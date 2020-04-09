FROM golang:1.14-alpine3.11

ENV OPENCV_VERSION=4.3.0
ENV BUILD="ca-certificates \
         musl \
         alpine-sdk \
         make \
         gcc \
         bash \
         nano \
         libc-dev \
         linux-headers \
         libjpeg-turbo \
         libpng \
         tiff \
         pkgconf \
         openblas"
ENV IMG_MAGICK_BUILD="pkgconfig libx11 lcms2 libbz2 glib libintl"

ENV DEV="binutils \
         clang clang-dev g++ cmake git wget \
         openblas-dev musl-dev libjpeg-turbo-dev \
         libpng-dev tiff-dev"
ENV IMG_MAGICK_DEV="jpeg-dev tiff-dev giflib-dev glib-dev libx11-dev lcms2-dev patch"
ENV PKG_CONFIG_PATH /usr/local/lib64/pkgconfig
ENV LD_LIBRARY_PATH /usr/local/lib64
ENV CGO_CPPFLAGS -I/usr/local/include
ENV CGO_CXXFLAGS "--std=c++1z"
ENV CGO_LDFLAGS "-L/usr/local/lib -lopencv_core -lopencv_face -lopencv_imgproc -lopencv_imgcodecs"
ENV IMAGEMAGICK_VERSION=6.9.8-3
ENV UFRAW_VERSION="0.22"
COPY ufraw.patch /
RUN apk update && \
    apk add --no-cache ${BUILD} ${IMG_MAGICK_BUILD} && \
    apk add --virtual dev-dependencies --no-cache ${DEV} ${IMG_MAGICK_DEV} && \
    # Fix libpng and xlocale.h path
    ln -vfs /usr/include/libpng16 /usr/include/libpng && \
    ln -vfs /usr/include/locale.h /usr/include/xlocale.h && \
    mkdir /tmp/opencv && \
    cd /tmp/opencv && \
    wget -O opencv.zip https://github.com/opencv/opencv/archive/${OPENCV_VERSION}.zip && \
    unzip opencv.zip && \
    wget -O opencv_contrib.zip https://github.com/opencv/opencv_contrib/archive/${OPENCV_VERSION}.zip && \
    unzip opencv_contrib.zip && \
    mkdir /tmp/opencv/opencv-${OPENCV_VERSION}/build && cd /tmp/opencv/opencv-${OPENCV_VERSION}/build && \
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
    -D OPENCV_GENERATE_PKGCONFIG=YES .. && \
    make -j4 && \
    make install && \
    cd && rm -rf /tmp/opencv && \
    go get -u -d gocv.io/x/gocv && go run ${GOPATH}/src/gocv.io/x/gocv/cmd/version/main.go && \
    cd && \
        mkdir /tmp/imagick && cd /tmp/imagick && \
	wget -O imagick.tar.gz https://github.com/ImageMagick/ImageMagick6/archive/${IMAGEMAGICK_VERSION}.tar.gz && \
	tar xvzf imagick.tar.gz && \
	cd ImageMagick* && \
	./configure \
	    --without-magick-plus-plus \
	    --without-perl \
	    --disable-openmp \
	    --with-gvc=no \
	    --without-threads \
	    --disable-docs && \
	make -j$(nproc) && make install && \
	ldconfig /usr/local/lib && \
    cd / && rm -rf /tmp/imagick && \
    curl -s --fail -L https://sourceforge.net/projects/ufraw/files/ufraw/ufraw-${UFRAW_VERSION}/ufraw-${UFRAW_VERSION}.tar.gz/download | tar -C /opt -xz && \
         patch /opt/ufraw-${UFRAW_VERSION}/dcraw.cc /ufraw.patch && \
         cd /opt/ufraw-${UFRAW_VERSION} && \
         ./configure --prefix=/usr/local --disable-dependency-tracking --without-gtk --without-gimp && \
         make && \
         make install && \
    cd / && rm -rf /opt/ufraw-${UFRAW_VERSION} && \
    apk del dev-dependencies && \
    rm -rf /var/cache/apk/*

CMD ["/bin/bash"]
