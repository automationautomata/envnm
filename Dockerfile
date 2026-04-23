FROM golang:latest AS builder

RUN apk add --no-cache git

WORKDIR /envmn

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/. logs/. config/. cmd/. internal/. ./

RUN CGO_ENABLED=1 GOOS=linux \
    go build -o envmn-server cmd/server \ 
    go build -o envmn-cli cmd/cli 

FROM alpine:latest

RUN apk add --no-cache openssl

WORKDIR /envmn

COPY --from=builder /envmn .

COPY scripts ./scripts 

RUN chmod -R +x scripts/

ENV CERTS_DIR=/envmn/certs \
    KEY_SEED_FILE=/envmn/seed

ENV LOG_LEVEL=INFO \
    ENVMN_HOST=0.0.0.0 \    
    ENVMN_PORT=8080 \
    ENVMN_PASSWORD=simple \
    METRICS_SERVER_HOST=0.0.0.0 \
    METRICS_SERVER_PORT=9090 \
    POSTGRES_MIN_CONNECTIONS=5 \ 
    POSTGRES_MAX_CONNECTIONS=20 \
    MAX_RETRIES=3 \
    RETRY_TIMEOUT=30s \
    CACHE_TTL=2h \
    ENVMN_SERVER_CA_CERT_PATH="${CERTS_DIR}/certs/ca.key" \
    ENVMN_SERVER_CERT_PATH="${CERTS_DIR}/server.crt" \
    ENVMN_SERVER_KEY_PATH="${CERTS_DIR}/server.key" \
    ENVMN_CLIENT_CA_CERT_PATH="${CERTS_DIR}/ca.crt" \ 
    ENVMN_CLIENT_CERT_PATH="${CERTS_DIR}/client.crt" \
    ENVMN_CLIENT_KEY_PATH="${CERTS_DIR}/client.key"
    
RUN ln -s ./envmn-cli /usr/bin/envmn && \
    mkdir $CERTS_DIR && \
    ./scripts/gen-certs.sh $CERTS_DIR && \
    ./scripts/gen-seed.sh $KEY_SEED_FILE
    
EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
