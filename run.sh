#!/bin/sh

cd "$(dirname "$0")/src"
go build -o ../app
cd ..
./app
