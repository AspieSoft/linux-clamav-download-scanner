#!/bin/bash

cd $(dirname "$0")

cd bin

go get -u
go mod tidy
go build
