#!/bin/bash

cd "$1"

CA_CN="certs-ca"
SERVER_CN="server"
CLIENT_CN="client"
DAYS="365"
KEY_SIZE="4096"

# ---------------------- CA ----------------------
openssl genrsa -out ca.key "$KEY_SIZE"
openssl req -new -x509 -key ca.key -sha256 -days "$DAYS" \
  -subj "/C=US/ST=CA/O=MyOrg/CN=$CA_CN" \
  -out ca.crt

# ---------------------- Server ----------------------
openssl genrsa -out server.key "$KEY_SIZE"

cat > server.cnf << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = US
ST = CA
O = MyOrg
CN = $SERVER_CN

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = $SERVER_CN
DNS.2 = localhost
DNS.3 = *.localhost
IP.1 = 127.0.0.1
IP.2 = 0.0.0.0
EOF

openssl req -new -key server.key \
  -config server.cnf \
  -out server.csr

openssl x509 -req -in server.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -out server.crt \
  -days "$DAYS" \
  -sha256 \
  -extensions v3_req \
  -extfile server.cnf

# ---------------------- Client ----------------------
openssl genrsa -out client.key "$KEY_SIZE"

cat > client.cnf << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = US
ST = CA
O = MyOrg
CN = $CLIENT_CN

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = clientAuth
EOF

openssl req -new -key client.key \
  -config client.cnf \
  -out client.csr

openssl x509 -req -in client.csr \
  -CA ca.crt \
  -CAkey ca.key \
  -CAcreateserial \
  -out client.crt \
  -days "$DAYS" \
  -sha256 \
  -extensions v3_req \
  -extfile client.cnf

# ---------------------- Cleanup ----------------------
echo cleanup...
rm -f *.csr *.cnf *.srl

# # Secure permissions
# chmod 600 "$CERT_DIR"/*.key
# chmod 644 "$CERT_DIR"/*.crt

