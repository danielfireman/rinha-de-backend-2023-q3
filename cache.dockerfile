FROM golang:1.19.0-alpine AS builder
RUN apk add git
WORKDIR /build

# Cache dependencies.
COPY ./go.* .
RUN go mod download

# Build the binary.
COPY . .
RUN cd cache/servidor && go build -o servidor . && mv servidor ../../

FROM alpine
RUN apk add --no-cache tzdata
COPY --from=builder /build/servidor /
EXPOSE 1313

# Inicia o servidor de cache
CMD ["/servidor"]