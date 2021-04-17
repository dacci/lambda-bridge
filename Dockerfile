ARG GOLANG_VERSION=1.16
ARG ALPINE_VERSION=3.13

# Build
FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS builder

WORKDIR /build
COPY . .
RUN go build

# Runtime
FROM alpine:${ALPINE_VERSION}

COPY --from=builder /build/lambda-bridge /

ENTRYPOINT [ "/lambda-bridge" ]
