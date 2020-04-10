#!/bin/sh

set -x
# set -e

wget -nc --output-document=sample.jpg https://cdn-images-1.medium.com/max/800/1*EYFejGUjvjPcc4PZTwoufw.jpeg
wget -nc --output-document=../models/coco.names https://raw.githubusercontent.com/AlexeyAB/darknet/master/data/coco.names
wget -nc --output-document=../models/yolov3.cfg https://raw.githubusercontent.com/AlexeyAB/darknet/master/cfg/yolov3.cfg
sed -i -e "\$anames = coco.names" ../models/yolov3.cfg
wget -nc --output-document=../models/yolov4-tiny.cfg https://raw.githubusercontent.com/ultralytics/yolov3/master/cfg/yolov4-tiny.cfg
wget -nc --output-document=../models/yolov4-tiny-1cls.cfg https://raw.githubusercontent.com/ultralytics/yolov3/master/cfg/yolov4-tiny-1cls.cfg
wget -nc --output-document=../models/yolov3.weights https://pjreddie.com/media/files/yolov3.weights
wget -nc --output-document=../models/yolov2.weights https://pjreddie.com/media/files/yolov2.weights


wget -O /usr/sbin/gdrivedl 'https://f.mjh.nz/gdrivedl'
chmod +x /usr/sbin/gdrivedl
gdrivedl https://drive.google.com/open?id=17G4W4JbSMHWIn9Qu9wBDKzSQEjGux_ps ../models/common.names
gdrivedl https://drive.google.com/open?id=1o1N0yV6R2HgcDsP9M9cVBmwAuGhC0OVF ../models/common.cfg
gdrivedl https://drive.google.com/open?id=1vY8MU0-_gcj65t7ya4KdLbv74dbxjmGQ ../models/common_inference.cfg
gdrivedl https://drive.google.com/open?id=1JWs98COfHm7Gz76_H3ezj4YrBYZXak9H ../models/common_220000.weights
