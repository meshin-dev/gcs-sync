#!/usr/bin/env bash

docker-compose -f docker-compose.dev.yaml build || exit 1
docker-compose -f docker-compose.dev.yaml up -d
