wget --output-document=sample.jpg https://cdn-images-1.medium.com/max/800/1*EYFejGUjvjPcc4PZTwoufw.jpeg
wget --output-document=./models/coco.names https://raw.githubusercontent.com/AlexeyAB/darknet/master/data/coco.names
wget --output-document=./models/yolov3.cfg https://raw.githubusercontent.com/AlexeyAB/darknet/master/cfg/yolov3.cfg
sed -i -e "\$anames = coco.names" ./models/yolov3.cfg
wget --output-document=./models/yolov3.weights https://pjreddie.com/media/files/yolov3.weights
