#!/bin/bash
openssl req -new -x509 -sha256 -newkey rsa:2048 -nodes -keyout backend.key -days 9999 -out backend.cert -subj '/CN=backend.local'
openssl req -new -x509 -sha256 -newkey rsa:2048 -nodes -keyout frontend.key -days 9999 -out frontend.cert -subj '/CN=frontend.local'