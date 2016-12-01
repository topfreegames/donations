# donations

[![Build Status](https://travis-ci.org/topfreegames/donations.svg?branch=master)](https://travis-ci.org/topfreegames/donations)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/donations/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/donations?branch=master)
[![Code Climate](https://codeclimate.com/github/topfreegames/donations/badges/gpa.svg)](https://codeclimate.com/github/topfreegames/donations)
[![Go Report Card](https://goreportcard.com/badge/github.com/topfreegames/donations)](https://goreportcard.com/report/github.com/topfreegames/donations)
[![Docs](https://readthedocs.org/projects/donations-api/badge/?version=latest
)](http://donations-api.readthedocs.io/en/latest/)
[![](https://imagelayers.io/badge/tfgco/donations:latest.svg)](https://imagelayers.io/?images=tfgco/donations:latest 'Donations Image Layers')

Donations is an HTTP API to control donations made by players in clans in games.

## Hacking Donations

### Setup

Make sure you have go installed on your machine.
If you use homebrew you can install it with `brew install go`.

Run `make setup`.

### Running the application

Run the api with `make run`.

### Running with docker

Provided you have docker installed, to build Donations's image run:

    $ make docker-build

To run a new donations instance, run:

    $ make docker-run

### Tests

Running tests can be done with `make test`.
