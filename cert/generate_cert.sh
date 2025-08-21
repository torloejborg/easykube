#!/bin/bash

# Define variables
CA_KEY="localtest.me.ca.key"
CA_CRT="localtest.me.ca.crt"
SERVER_KEY="localtest.me.key"
SERVER_CSR="localtest.me.csr"
SERVER_CRT="localtest.me.crt"
CONFIG_FILE="localtest.me.cnf"
DOMAIN="localtest.me"

# Step 1: Create CA

# Generate CA private key
openssl genpkey -algorithm RSA -out $CA_KEY -pkeyopt rsa_keygen_bits:2048

# Generate self-signed CA certificate
openssl req -x509 -new -nodes -key $CA_KEY -sha256 -days 3650 -out $CA_CRT -subj "/C=DK/ST=DK/L=Copenhagen/O=UFST/OU=Appudvikling/CN=RootCA"

# Step 2: Create Certificate for Wildcard Domain

# Generate server private key
openssl genpkey -algorithm RSA -out $SERVER_KEY -pkeyopt rsa_keygen_bits:2048

# Create configuration file for the wildcard domain
cat > $CONFIG_FILE <<EOF
[req]
default_bits       = 2048
default_md         = sha256
default_keyfile    = $SERVER_KEY
distinguished_name = req_distinguished_name
req_extensions     = req_ext
prompt            = no
encrypt_key       = no

[req_distinguished_name]
C  = DK
ST = DK
L  = Copenhagen
O  = UFST
OU = Appudvikling
CN = *.$DOMAIN

[req_ext]
subjectAltName = @alt_names

[alt_names]
DNS.1   = *.$DOMAIN
DNS.2   = $DOMAIN
EOF

# Generate certificate signing request (CSR)
openssl req -new -out $SERVER_CSR -config $CONFIG_FILE

# Sign the CSR with the CA
openssl x509 -req -in $SERVER_CSR -CA $CA_CRT -CAkey $CA_KEY -CAcreateserial -out $SERVER_CRT -days 3650 -sha256 -extfile $CONFIG_FILE -extensions req_ext

# Step 3: Verify the Certificate
openssl verify -CAfile $CA_CRT $SERVER_CRT

if [ $? -eq 0 ]; then
  echo "Certificate verification successful."
else
  echo "Certificate verification failed."
fi

echo "CA Certificate: $CA_CRT"
echo "CA Key: $CA_KEY"
echo "Server Certificate: $SERVER_CRT"
echo "Server Key: $SERVER_KEY"


kubectl create secret tls default-cert --cert=localtest.me.crt --key=localtest.me.key --dry-run=client -o yaml > default-cert.yaml