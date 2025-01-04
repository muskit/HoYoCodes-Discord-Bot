#!/bin/sh

cd $(dirname $0)/..
go clean -testcache
godotenv -f ./.env go test ./...