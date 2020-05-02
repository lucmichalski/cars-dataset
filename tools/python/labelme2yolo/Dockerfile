FROM python:3.7.7
MAINTAINER Michalski Luc <michalski.luc@gmail.com>

WORKDIR /opt

# Install requirements
COPY requirements.txt /opt/requirements.txt
RUN pip3 install -U pip && \
    pip install -r requirements.txt

# Copy python file
COPY labelme2yolo.py /labelme2yolo.py

# Defined shared volume
VOLUME ["/opt/data"]

# Define entrypoint
# ENTRYPOINT ["python3", "/labelme2yolo.py"]
CMD ["/bin/bash"]
