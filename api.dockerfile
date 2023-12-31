FROM golang:1.19.0-alpine AS builder
RUN apk add git
WORKDIR /build

# Cache dependencies.
COPY ./go.* .
RUN go mod download

# Build the binary.
COPY . .
RUN cd api && go build -o servidor . && mv servidor ../

FROM alpine
RUN apk add --no-cache tzdata
COPY --from=builder /build/servidor /
EXPOSE 8080

ENV GOMAXPROCS=1 \
    GOGC=300 \
    MONGODB_URI="mongodb://root:rootpassword@host.docker.internal:27017"

# Inicia a API
CMD ["/servidor"]