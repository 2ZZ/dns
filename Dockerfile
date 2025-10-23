# Multi-stage build for node-cache with local CoreDNS cache plugin
FROM golang:1.24.8-bookworm AS builder

# Set working directory
WORKDIR /go/src/k8s.io/dns

# Copy the DNS project
COPY . .

# Replace the vendored cache plugin with our local version
RUN rm -rf vendor/github.com/coredns/coredns/plugin/cache
COPY local-cache-plugin/ vendor/github.com/coredns/coredns/plugin/cache/

# Build arguments for multi-arch
ARG TARGETARCH
ARG TARGETOS

# Set environment variables for Go build (matching the original build script)
ENV CGO_ENABLED=0
ENV GOARCH=${TARGETARCH}
ENV PKG=k8s.io/dns
ENV VERSION=local-build

# Create necessary directories
RUN mkdir -p /go/bin/linux_${TARGETARCH} /go/pkg

# Build the node-cache binary using the original build approach
RUN go install -mod=mod \
    -installsuffix "static" \
    -ldflags "-X ${PKG}/pkg/version.VERSION=${VERSION}" \
    ./cmd/node-cache

# Find and copy the binary to a known location
RUN find /go/bin -name "node-cache" -exec cp {} /node-cache \; || \
    (ls -la /go/bin/ && exit 1)

# Final stage - use debian slim with basic tools
FROM debian:bookworm-slim

# Install iptables, curl and basic tools
RUN apt-get update && apt-get install -y \
    iptables \
    curl \
    jq \
    dnsutils \
    findutils \
    coreutils \
    && rm -rf /var/lib/apt/lists/*

# Copy the binary from builder stage
COPY --from=builder /node-cache /node-cache

# Expose DNS ports
EXPOSE 53 53/udp
EXPOSE 53 53/tcp

# Set entrypoint
ENTRYPOINT ["/node-cache"]