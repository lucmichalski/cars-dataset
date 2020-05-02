#!/bin/sh

set -x
set -e

nohup go run main.go --plugins autoscout24.be --extract --no-cache --clean --verbose > ./shared/nohup/autoscout24.be.out &
nohup go run main.go --plugins autoscout24.fr --extract --no-cache --clean --verbose > ./shared/nohup/autoscout24.fr.out &
nohup go run main.go --plugins autosphere.fr --extract --no-cache --clean --verbose > ./shared/nohup/autoscout24.fr.out &
