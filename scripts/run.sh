#!/bin/sh

cd "$(dirname "$0")/.."
go run -ldflags "-w -s" ./cmd/app
