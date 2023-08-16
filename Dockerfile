FROM golang:1.19.0-alpine AS builder
RUN apk add git
WORKDIR /build

# Cache dependencies.
COPY ./go.* ./
RUN go mod download

# Build the binary.
COPY . .
RUN go build -o api .

FROM alpine
RUN apk add --no-cache tzdata
COPY --from=builder /build/api /
EXPOSE 8080

# Inicia a API
CMD ["/api"]