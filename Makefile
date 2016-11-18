PACKAGES = $(shell glide novendor)
OS = "$(shell uname | awk '{ print tolower($$0) }')"

setup-docs:
	@pip install -q --log /tmp/pip.log --no-cache-dir sphinx recommonmark sphinx_rtd_theme

setup-hooks:
	@cd .git/hooks && ln -sf ../../hooks/pre-commit.sh pre-commit

setup:
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get -u github.com/Masterminds/glide/...
	@glide install

build:
	@go build $(PACKAGES)
	@go build -o bin/donations main.go

update-version:
	@go run main.go version > version.txt

cross:
	@mkdir -p ./bin
	@echo "Building for linux-i386..."
	@env GOOS=linux GOARCH=386 go build -o ./bin/donations-linux-i386
	@echo "Building for linux-x86_64..."
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/donations-linux-x86_64
	@echo "Building for darwin-i386..."
	@env GOOS=darwin GOARCH=386 go build -o ./bin/donations-darwin-i386
	@echo "Building for darwin-x86_64..."
	@env GOOS=darwin GOARCH=amd64 go build -o ./bin/donations-darwin-x86_64
	@chmod +x bin/*

docker-build:
	@mkdir -p ./_build/docker
	@docker build -t donations -f Dockerfile .

docker-services-dev: docker-services-shutdown
	@docker-compose -p donations-dev -f ./docker-compose-dev.yml up -d

docker-services-shutdown: docker-services-local-shutdown

docker-services-dev-shutdown:
	@docker-compose -p donations-dev -f ./docker-compose-dev.yml stop
	@docker-compose -p donations-dev -f ./docker-compose-dev.yml rm -f

services: docker-services-local
	@echo "Required services are up."

run: services
	@go run main.go start --fast -d -c ./config/local.yaml

run-basic: services
	@go run main.go start --fast -d -c ./config/basicAuth.yaml

test: update-version test-services test-migrate test-run test-coverage-func

test-run:
	@ginkgo --cover -r .

test-coverage:
	@rm -rf _build
	@mkdir -p _build
	@echo "mode: count" > _build/test-coverage-all.out
	@bash -c 'find . -name "*.coverprofile" | xargs tail -n +2 | egrep -v "\=\=>" | egrep -v "^$$" >> _build/test-coverage-all.out'

test-coverage-html: test-coverage
	@go tool cover -html=_build/test-coverage-all.out

test-coverage-func: test-coverage
	@go tool cover -func=_build/test-coverage-all.out | egrep -v "100\.0\%"

test-services: docker-services-local
	@echo "Required test services are up."

test-migrate:
	#@go run main.go migrate -d "mongodb://localhost:9999" -n "donations-test"

perf-migrate:
	@go run main.go migrate -d "mongodb://localhost:9999" -n "donations-perf"

ci-migrate:
	@go run main.go migrate -d "mongodb://mongo:27017" -n "donations-test"
	@go run main.go migrate -d "mongodb://mongo:27017" -n "donations-test" -v 0
	@go run main.go migrate -d "mongodb://mongo:27017" -n "donations-test"

test-all: test run-test-donations-docker run-perf

docker-services-local: docker-services-shutdown
	@docker-compose -p donations up -d

docker-services-local-shutdown:
	@docker-compose -p donations stop
	@docker-compose -p donations rm -f

schema-update:
	@go generate ./models/*.go
	@go generate ./api/payload.go

run-test-donations: docker-services-shutdown perf-migrate run-test-ci

run-test-ci: build kill-test-donations
	@rm -rf /tmp/donations-bench.log
	@./bin/donations start -p 8080 -q --fast -c ./config/perf.yaml 2>&1 > /tmp/donations-bench.log &
	@sleep 2

run-test-donations-fg: build kill-test-donations
	@./bin/donations start -p 8080 -q --fast -c ./config/perf.yaml

kill-test-donations:
	@-ps aux | egrep './donations.+perf.yaml' | egrep -v grep | awk ' { print $$2 } ' | xargs kill -9

run-test-donations-docker: build kill-test-donations-docker
	@rm -rf /tmp/donations-bench.log
	@./bin/donations start -p 8080 -q --fast -c ./config/perf.yaml 2>&1 > /tmp/donations-bench.log &

run-test-donations-docker-fg: build kill-test-donations-docker
	@./bin/donations start -p 8080 -q --fast -c ./config/perf.yaml

kill-test-donations-docker:
	@-ps aux | egrep './donations.+perf.yaml' | egrep -v grep | awk ' { print $$2 } ' | xargs kill -9

run-perf:
	@go test -bench . -benchtime 3s ./bench/...

docker-run-tests: docker-build docker-services-ci
	@docker-compose -p donations-ci -f ./docker-compose-ci.yml run --entrypoint="make docker-run-tests-exec" app

docker-run-perf: docker-build docker-services-ci
	@docker-compose -p donations-ci -f ./docker-compose-ci.yml run --entrypoint="make ci-migrate run-test-ci run-perf" app

docker-run-tests-exec: ci-migrate
	@ginkgo -r --randomizeAllSpecs --randomizeSuites --cover .
	@$(MAKE) test-coverage-func

docker-services-ci: docker-services-shutdown
	@docker-compose -p donations-ci -f ./docker-compose-ci.yml up -d

docker-services-ci-shutdown:
	@docker-compose -p donations-ci -f ./docker-compose-ci.yml stop
	@docker-compose -p donations-ci -f ./docker-compose-ci.yml rm -f

rtfd: update-version
	@rm -rf docs/_build
	@sphinx-build -b html -d ./docs/_build/doctrees ./docs/ docs/_build/html
	@open docs/_build/html/index.html
