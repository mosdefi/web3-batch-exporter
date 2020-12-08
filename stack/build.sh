#!/bin/bash
set -e

docker-compose build --no-cache web3-batch-service web3-batch-exporter
docker-compose up -d