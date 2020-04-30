#!/bin/bash

export OPENCV_VERSION=4.3.0

mkdir -p /tmp/opencv
cd /tmp/opencv
wget -O opencv.zip https://github.com/opencv/opencv/archive/${OPENCV_VERSION}.zip
unzip opencv.zip
wget -O opencv_contrib.zip https://github.com/opencv/opencv_contrib/archive/${OPENCV_VERSION}.zip
unzip opencv_contrib.zip
mkdir -p /tmp/opencv/opencv-${OPENCV_VERSION}/build
cd /tmp/opencv/opencv-${OPENCV_VERSION}/build

cmake \
-D CMAKE_BUILD_TYPE=RELEASE \
-D CMAKE_INSTALL_PREFIX=/usr/local \
-D CMAKE_C_COMPILER=/usr/bin/clang \
-D CMAKE_CXX_COMPILER=/usr/bin/clang++ \
-D OPENCV_EXTRA_MODULES_PATH=/tmp/opencv/opencv_contrib-${OPENCV_VERSION}/modules \
-D INSTALL_C_EXAMPLES=NO \
-D INSTALL_PYTHON_EXAMPLES=NO \
-D BUILD_ANDROID_EXAMPLES=NO \
-D BUILD_DOCS=NO \
-D BUILD_TESTS=NO \
-D BUILD_PERF_TESTS=NO \
-D BUILD_EXAMPLES=NO \
-D WITH_JASPER=OFF \
-D WITH_FFMPEG=NO \
-D BUILD_opencv_java=NO \
-D BUILD_opencv_python=NO \
-D BUILD_opencv_python2=NO \
-D BUILD_opencv_python3=NO \
-D OPENCV_GENERATE_PKGCONFIG=YES ..

make -j12
make install

cd
rm -rf /tmp/opencv
