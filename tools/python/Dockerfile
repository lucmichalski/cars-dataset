FROM python:3.7.7
MAINTAINER Michalski Luc <michalski.luc@gmail.com>

# Install runtime/build packages
# RUN apk add --no-cache --no-progress -X http://dl-cdn.alpinelinux.org/alpine/edge/testing nano bash jq py3-numpy-dev py3-pandas py3-scipy gcc make automake gcc g++ python3-dev openblas-dev

WORKDIR /opt

# Install requirements
COPY requirements.txt /opt/requirements.txt
RUN pip3 install -U pip && \
    pip install -r requirements.txt

# Copy python file
COPY stanford-map2csv.py /stanford-map2csv.py

# Defined shared volume
VOLUME ["/opt/data"]

# Define entrypoint
ENTRYPOINT ["python3", "/stanford-map2csv.py"]
