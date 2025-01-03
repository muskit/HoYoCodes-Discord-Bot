#!/bin/sh

cd "$(dirname "$0")/src"
go build ../app
cd ..
./app
