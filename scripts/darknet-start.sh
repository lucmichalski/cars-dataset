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

docker run --name darknet --runtime=nvidia -d -p 9003:9003 -e DARKNET_PORT=9003 -v `pwd`/darknet:/darknet -v `pwd`/models:/darknet/models lucmichalski/darknet:gpu-latest sh -c 'go run server.go --config-file=./models/stanford-1class.cfg --weights-file=./models/stanford-1class_final.weights'
docker run --name darknet2 --runtime=nvidia -d -p 9004:9004 -e DARKNET_PORT=9004 -v `pwd`/darknet:/darknet -v `pwd`/models:/darknet/models lucmichalski/darknet:gpu-latest sh -c 'go run server.go --config-file=./models/stanford-1class.cfg --weights-file=./models/stanford-1class_final.weights'
docker run --name darknet3 --runtime=nvidia -d -p 9005:9005 -e DARKNET_PORT=9005 -v `pwd`/darknet:/darknet -v `pwd`/models:/darknet/models lucmichalski/darknet:gpu-latest sh -c 'go run server.go --config-file=./models/stanford-1class.cfg --weights-file=./models/stanford-1class_final.weights'
docker run --name darknet4 --runtime=nvidia -d -p 9006:9006 -e DARKNET_PORT=9006 -v `pwd`/darknet:/darknet -v `pwd`/models:/darknet/models lucmichalski/darknet:gpu-latest sh -c 'go run server.go --config-file=./models/yolov4.cfg --weights-file=./models/yolov4.weights'
docker run --name darknet5 --runtime=nvidia -d -p 9007:9007 -e DARKNET_PORT=9007 -v `pwd`/darknet:/darknet -v `pwd`/models:/darknet/models lucmichalski/darknet:gpu-latest sh -c 'go run server.go --config-file=./models/stanford-1class.cfg --weights-file=./models/stanford-1class_final.weights'
