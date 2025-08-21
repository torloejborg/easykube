#!/bin/bash

#
# Note: This script written for Arch based linux distros.
#

CERT_FILE="localtest.me.ca.crt"

if [ -z "$CERT_FILE" ]; then
  echo "Certificate not found"
  exit 1
fi

if ! grep -q "BEGIN CERTIFICATE" "$CERT_FILE"; then
  echo "Error: $CERT_FILE does not appear to be a valid PEM certificate."
  exit 1
fi

CERT_NAME=$(basename "$CERT_FILE")

# Copy to CA anchor directory
echo "Copying $CERT_NAME to /etc/ca-certificates/trust-source/anchors/ ..."
sudo cp "$CERT_FILE" "/etc/ca-certificates/trust-source/anchors/$CERT_NAME"

# Update the system trust store
echo "Updating system trust store ..."
sudo trust extract-compat

# Confirm it was added as a trusted anchor
echo
echo "Verifying trust installation..."
trust list | grep -A 10 "$(openssl x509 -in "$CERT_FILE" -noout -subject | sed 's/subject= //')"

echo
echo "✅ Certificate installed as a trusted CA anchor."
