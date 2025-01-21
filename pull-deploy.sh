#!/bin/sh

git pull
docker-compose down
docker-compose rm -f
docker-compose pull
docker-compose up --build -d