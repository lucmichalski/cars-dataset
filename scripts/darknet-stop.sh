#!/bin/sh

set -x
# set -e

docker stop darknet
docker stop darknet2
docker stop darknet3
docker stop darknet4
docker stop darknet5

docker rm darknet
docker rm darknet2
docker rm darknet3
docker rm darknet4
docker rm darknet5
