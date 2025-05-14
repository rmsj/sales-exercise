# Simple Sales API

[![Go Report Card](https://goreportcard.com/badge/github.com/rmsj/service)](https://goreportcard.com/report/github.com/rmsj/service)
[![go.mod Go version](https://img.shields.io/github/go-mod/go-version/rmsj/service)](https://github.com/rmsj/service)

## Description

This is a simple sales API using as base the [Ardanlabs starter kit repo](https://github.com/ardanlabs/service)
It has been simplified, to run only using docker compose.

## Index

* [Installation](https://github.com/rmsj/service?tab=readme-ov-file#installation)
* [Running The Project](https://github.com/rmsj/service?tab=readme-ov-file#running-the-project)

## Installation

To clone the project, create a folder and use the git clone command. Then please read the [makefile](makefile) file to learn how to install all the tooling and docker images.

```
$ cd $HOME
$ mkdir code
$ cd code
$ git clone https://github.com/rmsj/sales-exercise or git@github.com:rmsj/sales-exercise.git
$ cd sales-exerciset
```

From the root of the project run:
```
cp ./zarf/compose/.env.example ./zarf/compose/.env
```
This will copy the file `/zarf/compose/.env.example` to `/zarf/compose/.env`.

## Running The Project

To run the project use the following commands.

```
# Install Tooling
$ make dev-gotooling
$ make dev-docker

# Run Tests
$ make test

# Shutdown Tests
$ make test-down

# Run Project
$ make compose-build-up

# Check Logs
$ make compose-logs

# Shut Project
$ make compose-down
```

## Testing The Project

Import the Postman collection file `sales.postman_collection.json` in the root of the project into Postman.

You need to run the `login` request first - this will log you to be able to create products, sales, etc.