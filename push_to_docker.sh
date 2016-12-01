#!/bin/bash

VERSION=$(cat version.txt | sed "s@donations v@@g")

cp ./config/default.yaml ./dev

docker login -e="$DOCKER_EMAIL" -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

docker build -t donations .
docker tag donations:latest tfgco/donations:$VERSION.$TRAVIS_BUILD_NUMBER
docker tag donations:latest tfgco/donations:$VERSION
docker tag donations:latest tfgco/donations:latest
docker push tfgco/donations:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/donations:$VERSION
docker push tfgco/donations:latest

DOCKERHUB_LATEST=$(python get_latest_tag.py)

if [ "$DOCKERHUB_LATEST" != "$VERSION.$TRAVIS_BUILD_NUMBER" ]; then
    echo "Last version is not in docker hub!"
    echo "docker hub: $DOCKERHUB_LATEST, expected: $VERSION.$TRAVIS_BUILD_NUMBER"
    exit 1
fi
