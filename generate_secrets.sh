#!/bin/bash

KEY_PATH_BASE="/etc/pki"
KEY_PATH_JWT="$KEY_PATH_BASE/jwt"
KEY_PATH_DB="$KEY_PATH_BASE/db"

mkdir -p $KEY_PATH

#openssl req -new -newkey rsa:4096 -days 70000 -nodes -x509 -subj "/C=US/ST=/L=/O=/CN=idp" -keyout $KEY_PATH/private.pem 1>/dev/null 2>&1
openssl genrsa -out $KEY_PATH/private.pem -passout pass:keysecret 2048 1>/dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "error generating private key"
  exit 1
fi
openssl rsa -in $KEY_PATH/private.pem -out $KEY_PATH/public.pem -RSAPublicKey_out 1>/dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "error generating public key"
  exit 1
fi

mkdir -p $KEY_PATH_DB
echo "example" > /etc/pki/db/secret
