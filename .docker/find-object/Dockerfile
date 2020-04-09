FROM lucmichalski/gocv:alpine-cpu-latest

RUN apk add --no-cache qt5-qtbase-dev cmake \
    && git clone --depth=1 https://github.com/introlab/find-object /find-object \
    && cd /find-object/build \
    && cmake -DCMAKE_BUILD_TYPE=Release .. \
    && make -j4

CMD ["/bin/bash"]
