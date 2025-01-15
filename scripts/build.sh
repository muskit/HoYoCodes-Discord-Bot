#!/bin/sh

cd "$(dirname "$0")/.."
go build -ldflags "-w -s" -o app ./cmd/app
