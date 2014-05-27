#!/bin/sh

# TODO : switch to better SSL alternative ASAP
which openssl || (echo 'OpenSSL not found.' && exit 1)

# Generate sample auto-sign x509 certificate.
# Not using crypto/tls/generate_cert.go because
# the certificate it generates couldn't be use for
# both API & UI.

# using https://www.openssl.org/docs/HOWTO/keys.txt
# removed -des3 (password protection)
echo '# Generating private key (key.pem)'
openssl genrsa -out key.pem 2048
echo ''

# using https://www.openssl.org/docs/HOWTO/certificates.txt
echo '# Generating associated certificate (cert.pem)'
echo '# (WATCH OUT for the Common Name field, eg. www.mywebsite.com)'
openssl req -new -x509 -key key.pem -out cert.pem -days 1095

echo '# Installing cert/key for example/'
cp *.pem example/
cp cert.pem example/conf/auth-cert.pem 

echo '# Installing cert/key for storeexample/'
cp *.pem storexample/
cp cert.pem storexample/auth-cert.pem 

