# Multi-arch image (linux/amd64, linux/arm64, linux/arm/v7):
#   docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t jgulick48/pro-check-server:<tag> --push .
# Scanning uses a raw HCI socket, so run it as root with CAP_NET_RAW and
# CAP_NET_ADMIN (or privileged) and host networking.
FROM --platform=$BUILDPLATFORM golang:1.24 AS builder
ARG TARGETOS TARGETARCH TARGETVARIANT
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN go test ./...
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /pro-check-server .

FROM gcr.io/distroless/static-debian12
COPY --from=builder /pro-check-server /bin/pro-check-server
EXPOSE 8000
ENTRYPOINT ["/bin/pro-check-server"]
