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

WORKDIR /envmn

COPY --from=builder /envmn .

ENV SERVER_HOST=0.0.0.0 \    
    SERVER_PORT=8080 \
    POSTGRES_MIN_CONNECTIONS=5 \ 
    POSTGRES_MAX_CONNECTIONS=20 

RUN ln -s /envmn/envmn-cli /usr/bin/envmn 

EXPOSE 8080

CMD ["./envmn-server"]
