FROM python:3.7.7
MAINTAINER Michalski Luc <michalski.luc@gmail.com>

RUN apt-get update && apt-get install -y nano jq bash

WORKDIR /opt

# Install requirements
COPY requirements.txt /opt/requirements.txt
RUN pip3 install -U pip && \
    pip install -r requirements.txt

# Copy python file
COPY compcars-mat2csv.py /compcars-mat2csv.py

# Defined shared volume
VOLUME ["/opt/data"]

# Define entrypoint
# ENTRYPOINT ["python3", "/compcars-map2csv.py"]
CMD ["/bin/bash"]
