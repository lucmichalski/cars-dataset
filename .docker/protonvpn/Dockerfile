FROM archlinux

ENV COLUMNS 80
ENV LINES 24
# get these from https://protonvpn.com/support/vpn-login/
ARG PROTONVPN_USER=${PROTONVPN_USER}
ARG PROTONVPN_PASSWORD=${PROTONVPN_PASSWORD}

COPY ./vpn-setup.exp /tmp/vpn-setup.exp

RUN pacman -Sy && \
    pacman -S --noconfirm which expect openvpn dialog python-pip python-setuptools git dante libffi && \
    pip install git+https://github.com/krey/protonvpn-cli-ng.git && \
    expect /tmp/vpn-setup.exp && \
    rm /tmp/vpn-setup.exp
    
RUN ln -sf /usr/lib/libffi.so.7 /usr/lib/libffi.so.6

COPY ./sockd.conf /etc/sockd.conf
