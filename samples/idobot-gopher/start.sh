#!/bin/sh

docker run -e IDOBATA_API_TOKEN=${IDOBATA_API_TOKEN} -d -v `pwd`/app_data:/app_data gopher 
