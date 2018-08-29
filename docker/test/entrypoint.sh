#!/bin/bash

# Setup SSL
cp /mongod.key /tmp/mongod.key
cp /mongod.pem /tmp/mongod.pem
cp /rootCA.crt /tmp/mongod-rootCA.crt
chmod 600 /tmp/mongod.key /tmp/mongod.pem /tmp/mongod-rootCA.pem

/usr/bin/mongod \
	--replSet=${TEST_RS_NAME:-rs} \
	--keyFile=/tmp/mongod.key \
	--sslMode=preferSSL \
	--sslCAFile=/tmp/mongod-rootCA.crt \
	--sslPEMKeyFile=/tmp/mongod.pem \
	$*
