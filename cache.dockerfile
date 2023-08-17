FROM golang:1.19.0-alpine AS builder
RUN apk add git
WORKDIR /build

# Cache dependencies.
COPY ./go.* .
RUN go mod download

# Build the binary.
COPY ./cache/* .
RUN go build -o cache .

FROM alpine
RUN apk add --no-cache tzdata
COPY --from=builder /build/cache /
EXPOSE 1313

# Inicia o servidor de cache
CMD ["/cache"]