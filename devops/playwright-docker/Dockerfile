FROM node:13.8-buster

RUN apt update \
    && apt install -y \
        gconf-service \
        libasound2 \
        libatk1.0-0 \
        libatk-bridge2.0-0 \
        libc6 \
        libcairo2 \
        libcups2 \
        libdbus-1-3 \
        libegl1 \
        libenchant1c2a \
        libexpat1 \
        libfontconfig1 \
        libgbm1 \
        libgcc1 \
        libgconf-2-4 \
        libgdk-pixbuf2.0-0 \
        libgles2 \
        libglib2.0-0 \
        libgtk-3-0 \
        libhyphen0 \
        libnotify-dev \
        libnspr4 \
        libopus0 \
        libpango-1.0-0 \
        libpangocairo-1.0-0 \
        libsecret-1-0 \
        libstdc++6 \
        libvpx5 \
        libwebkit2gtk-4.0-37 \
        libwoff1 \
        libx11-6 \
        libx11-xcb1 \
        libxcb1 \
        libxcomposite1 \
        libxcursor1 \
        libxdamage1 \
        libxext6 \
        libxfixes3 \
        libxi6 \
        libxrandr2 \
        libxrender1 \
        libxss1 \
        libxtst6 \
        ca-certificates \
        fonts-liberation \
        libappindicator1 \
        libnss3 \
        lsb-release \
        xdg-utils \
        wget \
        xvfb

# Install webkit dependencies

# Install an old version of libjpeg for webkit
RUN wget http://www.ijg.org/files/jpegsrc.v8d.tar.gz \
    && tar xvf jpegsrc.v8d.tar.gz \
    && cd jpeg-8d \
    && ./configure \
    && make \
    && make test \
    && make install \
    && ln -s /usr/local/lib/libjpeg.so.8 /usr/lib/x86_64-linux-gnu/libjpeg.so.8 \
    && mkdir -p /usr/local/lib/webkit2gtk-4.0/injected-bundle \
    && ln -s /usr/lib/x86_64-linux-gnu/webkit2gtk-4.0/injected-bundle/libwebkit2gtkinjectedbundle.so /usr/local/lib/webkit2gtk-4.0/injected-bundle/libwebkit2gtkinjectedbundle.so

RUN mkdir -p /example/screenshots
COPY example/*.json example/*.js /example/
