#!/bin/bash

curl -X POST -H 'Content-Type: application/json' -d@sample_request.json localhost:8000/
docker logs -f stack_web3-batch-exporter_1
