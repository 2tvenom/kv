#!/usr/bin/env bash

openssl genrsa -out ca.key 4096
openssl req -x509 -new -key ca.key -days 10000 -out ca.crt -subj "/C=DE/ST=NRW/L=Berlin/O=My Inc/OU=DevOps/CN=www.example.com/emailAddress=dev@www.example.com"


openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/C=DE/ST=NRW/L=Berlin/O=My Inc/OU=DevOps/CN=test.com/emailAddress=dev@test.com"

openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 5000
