FROM golang:1.6.2-alpine

MAINTAINER TFG Co <backend@tfgco.com>

EXPOSE 8888

RUN apk update
RUN apk add --update git make g++ apache2-utils bash

RUN go get -u github.com/Masterminds/glide/...

ADD . /go/src/github.com/topfreegames/donations

WORKDIR /go/src/github.com/topfreegames/donations
RUN glide install
RUN go install github.com/topfreegames/donations

ENV DONATIONS_MONGODB_HOST 0.0.0.0
ENV DONATIONS_MONGODB_PORT 27017
ENV DONATIONS_MONGODB_DB donations
ENV DONATIONS_REDIS_URL redis://0.0.0.0:3456/0
ENV DONATIONS_REDIS_MAXIDLE 3
ENV DONATIONS_REDIS_IDLETIMEOUTSECONDS 240

ENV DONATIONS_SENTRY_URL ""
ENV DONATIONS_NEWRELIC_KEY ""
ENV DONATIONS_NEWRELIC_APPNAME ""

ENV DONATIONS_BASICAUTH_USERNAME ""
ENV DONATIONS_BASICAUTH_PASSWORD ""

CMD /go/bin/donations start --bind 0.0.0.0 --port 8888 --fast --config /go/src/github.com/topfreegames/donations/config/default.yaml
