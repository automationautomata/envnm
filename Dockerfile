FROM golang:latest AS builder

RUN apk add --no-cache git

WORKDIR /envmn

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/. logs/. config/. cmd/. internal/. ./

RUN CGO_ENABLED=1 GOOS=linux \
    go build -o envmn-server cmd/server  \ 
    go build -o envmn-cli cmd/cli 

FROM alpine:latest

RUN apk add --no-cache openssl

WORKDIR /envmn

COPY --from=builder /envmn .

COPY scripts ./scripts 

RUN chmod -R +x scripts/

ENV CERTS_DIR=/envmn/certs \
    KEY_SEED_FILE=/envmn/seed

ENV ENVMN_HOST=0.0.0.0 \    
    ENVMN_PORT=8080 \
    POSTGRES_MIN_CONNECTIONS=5 \ 
    POSTGRES_MAX_CONNECTIONS=20 \
    MAX_RETRIES=3 \
    RETRY_TIMEOUT=30s \
    LOG_LEVEL=INFO \
    SERVER_CA_CERT_PATH="${CERTS_DIR}/certs/ca.key" \
    SERVER_CERT_PATH="${CERTS_DIR}/server.crt" \
    SERVER_KEY_PATH="${CERTS_DIR}/server.key" \
    CLIENT_CA_CERT_PATH="${CERTS_DIR}/ca.crt" \ 
    CLIENT_CERT_PATH="${CERTS_DIR}/client.crt" \
    CLIENT_KEY_PATH="${CERTS_DIR}/client.key"
    
RUN ln -s $CERTS_DIR /usr/bin/envmn && \
    mkdir $CERTS_DIR && \
    ./gen-certs.sh $CERTS_DIR && \
    ./gen-seed.sh $KEY_SEED_FILE
    
EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
