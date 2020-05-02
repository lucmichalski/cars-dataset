FROM alpine:3.11

MAINTAINER Michalski Luc "michalski.luc@gmail.com"

ENV TZ Europe/Paris
ENV NGINX_VERSION 1.17.7
ENV LUA_MODULE_VERSION 0.10.13
ENV DEVEL_KIT_MODULE_VERSION 0.3.0
ENV LUAJIT_LIB=/usr/lib
ENV LUAJIT_INC=/usr/include/luajit-2.1
ENV MAXMIND_VERSION=1.4.2
ENV NGX_BROTLI_COMMIT e505dce68acc190cc5a1e780a3b0275e39f160ca
# https://github.com/openresty/headers-more-nginx-module/releases
ENV HEADERS_MORE_VERSION=0.33

# TODO fix issue with non zero return code

RUN ldconfig || :

RUN GPG_KEYS=a \
	&& CONFIG="\
		--prefix=/etc/nginx \
		--sbin-path=/usr/sbin/nginx \
		--modules-path=/usr/lib/nginx/modules \
		--conf-path=/etc/nginx/nginx.conf \
		--error-log-path=/var/log/nginx/error.log \
		--http-log-path=/var/log/nginx/access.log \
		--pid-path=/var/run/nginx.pid \
		--lock-path=/var/run/nginx.lock \
		--http-client-body-temp-path=/var/cache/nginx/client_temp \
		--http-proxy-temp-path=/var/cache/nginx/proxy_temp \
		--http-fastcgi-temp-path=/var/cache/nginx/fastcgi_temp \
		--http-uwsgi-temp-path=/var/cache/nginx/uwsgi_temp \
		--http-scgi-temp-path=/var/cache/nginx/scgi_temp \
		--user=nginx \
		--group=nginx \
		--with-http_ssl_module \
		--with-http_realip_module \
		--with-http_addition_module \
		--with-http_sub_module \
		--with-http_dav_module \
		--with-http_flv_module \
		--with-http_mp4_module \
		--with-http_gunzip_module \
		--with-http_gzip_static_module \
		--with-http_random_index_module \
		--with-http_secure_link_module \
		--with-http_stub_status_module \
		--with-http_auth_request_module \
		--with-http_xslt_module=dynamic \
		--with-http_image_filter_module=dynamic \
		--with-http_geoip_module=dynamic \
		--with-http_perl_module=dynamic \
		--with-threads \
		--with-stream \
		--with-stream_ssl_module \
		--with-http_slice_module \
		--with-mail \
		--with-mail_ssl_module \
		--with-file-aio \
		--with-http_v2_module \
		--with-ipv6 \
		--add-dynamic-module=/ngx_http_geoip2_module \
		--add-module=/usr/src/ngx_devel_kit-$DEVEL_KIT_MODULE_VERSION \
    	--add-module=/usr/src/lua-nginx-module-$LUA_MODULE_VERSION \
    	--add-module=/usr/src/ngx_brotli \
    	--add-module=/usr/src/headers-more-nginx-module-$HEADERS_MORE_VERSION \
		--add-dynamic-module=/usr/src/ModSecurity-nginx \
        --with-cc-opt=-Wno-error \
	" \
	&& addgroup -S nginx \
	&& adduser -D -S -h /var/cache/nginx -s /sbin/nologin -G nginx nginx \
	\
	&& set -x \
	&& apk add --no-cache --virtual .build-deps alpine-sdk perl \
	&& git clone https://github.com/leev/ngx_http_geoip2_module /ngx_http_geoip2_module \
	&& wget https://github.com/maxmind/libmaxminddb/releases/download/${MAXMIND_VERSION}/libmaxminddb-${MAXMIND_VERSION}.tar.gz \
	&& tar xf libmaxminddb-${MAXMIND_VERSION}.tar.gz \
	&& cd libmaxminddb-${MAXMIND_VERSION} \
	&& ./configure \
	&& make -j$(getconf _NPROCESSORS_ONLN) \
	&& make check \
	&& make install \
	&& apk del .build-deps \
	\
	&& apk add --no-cache tzdata curl wget \
	&& apk add --no-cache --virtual .build-deps \
		gcc \
		libc-dev \
		make \
		openssl-dev \
		pcre-dev \
		zlib-dev \
		linux-headers \
		gnupg \
		libxslt-dev \
		gd-dev \
		geoip-dev \
		perl-dev \
		luajit-dev \
	&& apk add --no-cache --virtual .modsec-build-deps \
		libxml2-dev \
		flex \
		bison \
		yajl-dev \
		file \
	&& apk add --no-cache --virtual .brotli-build-deps \
		autoconf \
		libtool \
		automake \
		git \
		g++ \
		cmake \
	&& mkdir -p /usr/src \
	\
	&& cd /usr/src \
	&& git clone --recursive https://github.com/google/ngx_brotli.git \
	&& cd ngx_brotli \
	&& git checkout -b $NGX_BROTLI_COMMIT $NGX_BROTLI_COMMIT \
	&& cd .. \
	\
	####################### Add ModSecurity module
	\
	&& cd /usr/src \
	&& git clone --depth 1 --single-branch https://github.com/SpiderLabs/ModSecurity \
	&& cd ModSecurity \
	&& git submodule init \
	&& git submodule update \
	&& ./build.sh \
	&& ./configure \
	&& make -j$(getconf _NPROCESSORS_ONLN) \
	&& make install \
	&& rm -fR /usr/src/ModSecurity \
	\
	&& cd /usr/src \
	&& git clone --depth 1 https://github.com/SpiderLabs/ModSecurity-nginx.git \
	&& mkdir -p /etc/nginx/modsec \
	&& wget -qc -P /etc/nginx/modsec/ https://raw.githubusercontent.com/SpiderLabs/ModSecurity/v3/master/modsecurity.conf-recommended \
	&& mv /etc/nginx/modsec/modsecurity.conf-recommended /etc/nginx/modsec/modsecurity.conf \
	&& sed -i 's/SecRuleEngine DetectionOnly/SecRuleEngine On/' /etc/nginx/modsec/modsecurity.conf \
	\
	&& cd /usr/src \
	&& curl -fSL https://nginx.org/download/nginx-$NGINX_VERSION.tar.gz -o nginx.tar.gz \
	&& curl -fSL https://nginx.org/download/nginx-$NGINX_VERSION.tar.gz.asc  -o nginx.tar.gz.asc \
	&& curl -fSL https://github.com/simpl/ngx_devel_kit/archive/v$DEVEL_KIT_MODULE_VERSION.tar.gz -o ndk.tar.gz \
  	&& curl -fSL https://github.com/openresty/lua-nginx-module/archive/v$LUA_MODULE_VERSION.tar.gz -o lua.tar.gz \
	&& curl -fSL https://github.com/openresty/headers-more-nginx-module/archive/v$HEADERS_MORE_VERSION.tar.gz -o headers-more-nginx-module-$HEADERS_MORE_MODULE_PATH.tar.gz \
	&& sha512sum nginx.tar.gz nginx.tar.gz.asc \
	&& export GNUPGHOME="$(mktemp -d)" \
	&& rm -rf "$GNUPGHOME" nginx.tar.gz.asc \
	&& mkdir -p /usr/src \
	&& tar -zxC /usr/src -f nginx.tar.gz \
	&& tar -zxC /usr/src -f ndk.tar.gz \
	&& tar -zxC /usr/src -f lua.tar.gz \
	&& tar -zxC /usr/src -f headers-more-nginx-module-$HEADERS_MORE_MODULE_PATH.tar.gz \
	&& rm nginx.tar.gz ndk.tar.gz lua.tar.gz headers-more-nginx-module-$HEADERS_MORE_MODULE_PATH.tar.gz \
	&& cd /usr/src/nginx-$NGINX_VERSION \
	&& ./configure $CONFIG --with-debug \
	&& make -j$(getconf _NPROCESSORS_ONLN) \
	&& mv objs/nginx objs/nginx-debug \
	&& mv objs/ngx_http_xslt_filter_module.so objs/ngx_http_xslt_filter_module-debug.so \
	&& mv objs/ngx_http_image_filter_module.so objs/ngx_http_image_filter_module-debug.so \
	&& mv objs/ngx_http_geoip_module.so objs/ngx_http_geoip_module-debug.so \
	&& mv objs/ngx_http_perl_module.so objs/ngx_http_perl_module-debug.so \
	&& ./configure $CONFIG \
	&& make -j$(getconf _NPROCESSORS_ONLN) \
	&& make install \
	&& rm -rf /etc/nginx/html/ \
	&& mkdir /etc/nginx/conf.d/ \
	&& mkdir -p /usr/share/nginx/html/ \
	&& install -m644 html/index.html /usr/share/nginx/html/ \
	&& install -m644 html/50x.html /usr/share/nginx/html/ \
	&& install -m755 objs/nginx-debug /usr/sbin/nginx-debug \
	&& install -m755 objs/ngx_http_xslt_filter_module-debug.so /usr/lib/nginx/modules/ngx_http_xslt_filter_module-debug.so \
	&& install -m755 objs/ngx_http_image_filter_module-debug.so /usr/lib/nginx/modules/ngx_http_image_filter_module-debug.so \
	&& install -m755 objs/ngx_http_geoip_module-debug.so /usr/lib/nginx/modules/ngx_http_geoip_module-debug.so \
	&& install -m755 objs/ngx_http_perl_module-debug.so /usr/lib/nginx/modules/ngx_http_perl_module-debug.so \
	&& ln -s ../../usr/lib/nginx/modules /etc/nginx/modules \
	&& strip /usr/sbin/nginx* \
	&& strip /usr/lib/nginx/modules/*.so \
	&& rm -rf /usr/src/nginx-$NGINX_VERSION \
	&& rm -rf /usr/src/ngx_brotli \
	\
    # Install mime types and other files
    && cd /usr/src \
    && curl -fSL https://raw.githubusercontent.com/nginx/nginx/master/conf/mime.types -o mime.types \
    && curl -fSL https://raw.githubusercontent.com/nginx/nginx/master/conf/fastcgi_params -o fastcgi_params \
	&& curl -fSL https://gist.githubusercontent.com/romanoffs/29b981cccff51b0ea564e258e1ed2e85/raw/17bdc5b7c940604d84a9481fe28010d7b93ab043/cloudflare.conf -o cloudflare.conf \
	&& mv /usr/src/mime.types /etc/ \
	&& mv /usr/src/fastcgi_params /etc/ \
	&& mv /usr/src/cloudflare.conf /etc/ \
	\
	# Bring in gettext so we can get `envsubst`, then throw
	# the rest away. To do this, we need to install `gettext`
	# then move `envsubst` out of the way so `gettext` can
	# be deleted completely, then move `envsubst` back.
	&& apk add --no-cache --virtual .gettext gettext \
	&& mv /usr/bin/envsubst /tmp/ \
	\
	&& runDeps="$( \
		scanelf --needed --nobanner /usr/sbin/nginx /usr/lib/nginx/modules/*.so /tmp/envsubst \
			| awk '{ gsub(/,/, "\nso:", $2); print "so:" $2 }' \
			| sort -u \
			| xargs -r apk info --installed \
			| sort -u \
	)" \
	&& apk add --no-cache --virtual .nginx-rundeps $runDeps \
	&& apk del .build-deps \
	&& apk del .brotli-build-deps \
	&& apk del .gettext \
	&& mv /tmp/envsubst /usr/local/bin/ \
	\
	&& rm -fR /usr/src/ModSecurity-nginx \
	\
	# forward request and error logs to docker log collector
	&& ln -sf /dev/stdout /var/log/nginx/access.log \
	&& ln -sf /dev/stderr /var/log/nginx/error.log

RUN apk del .modsec-build-deps

COPY conf/ /etc/nginx/

EXPOSE 80 443

CMD ["nginx", "-g", "daemon off;"]
