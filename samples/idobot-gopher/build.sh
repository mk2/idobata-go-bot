#!/bin/sh

cd ../../; docker build --no-cache -f samples/idobot-gopher/Dockerfile -t gopher .
