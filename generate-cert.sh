#!/bin/bash
# Generate self-signed TLS certificate for testing
# Usage: ./generate-cert.sh [hostname]

HOSTNAME="${1:-localhost}"

echo "Generating self-signed certificate for ${HOSTNAME}..."

openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt \
    -days 365 -nodes \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=${HOSTNAME}" \
    -addext "subjectAltName=DNS:${HOSTNAME},DNS:*.${HOSTNAME},IP:127.0.0.1"

echo ""
echo "Certificate generated successfully!"
echo "  Certificate: server.crt"
echo "  Private Key: server.key"
echo ""
echo "To start the server with TLS, run:"
echo "  go run cmd/gather-requests/main.go -tls -tls-cert=server.crt -tls-key=server.key"
echo ""
echo "Or set environment variables:"
echo "  export TLS_ENABLED=true"
echo "  export TLS_CERT=server.crt"
echo "  export TLS_KEY=server.key"
echo "  go run cmd/gather-requests/main.go"
