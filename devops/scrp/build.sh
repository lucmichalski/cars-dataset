#!/bin/bash
FILE=gsvc
if [ -f $FILE ]; then
   echo "
        ------------------------------------------------------
        -- Building update.
        ------------------------------------------------------
   "
else
   echo "

        ------------------------------------------------------
        -- README FIRST TO GET IT WORKING
        ------------------------------------------------------

        Install protobuf compiler...

        $ sudo apt-get install autoconf automake libtool curl make g++ unzip #!!! THIS will work on debian/ubuntu
        $ git clone https://github.com/google/protobuf
        $ cd protobuf
        $ ./autogen.sh
        $ ./configure
        $ make
        $ make check
        $ sudo make install
        $ sudo ldconfig 

        Install the protoc Go plugin

        $ go get -u github.com/golang/protobuf/protoc-gen-go

        Rebuild the generated Go code

        $ protoc -I helloworld/ helloworld/helloworld.proto --go_out=plugins=grpc:helloworld

    "
fi

#Setup Cassandra Schema
#cqlsh --ssl -f schema.1.cql 

#Generate certificates for gRPC
#Common Name (e.g. server FQDN or YOUR name) []:backend.local
#openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./backend.key -out ./backend.cert -subj "/C=US/ST=San Francisco/L=San Francisco/O=SFPL/OU=IT Department/CN=backend.local"
#openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./frontend.key -out ./frontend.cert -subj "/C=US/ST=San Francisco/L=San Francisco/O=SFPL/OU=IT Department/CN=frontend.local"

#Store the config in consul when ready
#traefik storeconfig
#then can just run
#./traefik #instead of ./traefik -c traefik.toml

#go generate github.com/dioptre/scrp/src/proto
protoc -I proto/ proto/helloworld.proto proto/scrape.proto --go_out=plugins=grpc:proto/

#Build client & server
go build -o gsvc -tags netgo service/*.go
go build -o gcli -tags netgo client/*.go

#On Server... run consul and traefik too, login credentials to cassandra can be changed in execution arguments
GOCQL_HOST_LOOKUP_PREFER_V4=true ./gsvc localhost false false cassandra-ca.cert cassandra-client.cert cassandra-client.key  # &; ./consul agent -config-file consul.json -bind 127.0.0.1 -bootstrap-expect 1 &; ./traefik -c traefik.toml &;
#On your client
#./gcli &;

#Dont forget to change/remove bootstrap-expect

#Or generate certificates