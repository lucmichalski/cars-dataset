FROM alpine:3.11

RUN apk add --no-cache --no-progress nano jq bash curl wget

WORKDIR /sandbox
EXPOSE 8090
VOLUME ["/opt/data"]

CMD ["/bin/bash"]
