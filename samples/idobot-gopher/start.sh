#!/bin/sh

docker run --restart=always -e IDOBATA_API_TOKEN=${IDOBATA_API_TOKEN} -e STORE_FILE_PATH=/app_data/gopher.store -d -v `pwd`/app_data:/app_data gopher 
