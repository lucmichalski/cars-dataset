FROM python:3.7.7
MAINTAINER Michalski Luc <michalski.luc@gmail.com>

WORKDIR /opt

# Install requirements
COPY requirements.txt /opt/requirements.txt
RUN pip3 install -U pip && \
    pip install -r requirements.txt

# Copy python file
COPY stanford-maptcsv.py /stanford-mat2csv.py

# Defined shared volume
VOLUME ["/opt/data"]

# Define entrypoint
ENTRYPOINT ["python3", "/stanford-map2csv.py"]
