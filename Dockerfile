# Stage 1: Build the Go binary
FROM amazonlinux:2 AS builder

# Install necessary tools, including tar and curl
RUN yum install -y tar gzip shadow-utils curl

# Install Go
RUN curl -LO https://golang.org/dl/go1.21.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz && \
    rm go1.21.0.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

# Set up the working directory
WORKDIR /app

# Copy the Go Modules manifests and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the remaining source code files
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager cmd/main.go

# Stage 2: Runtime environment
FROM amazonlinux:2

# Set up working directory
WORKDIR /

# Copy the Go binary from the builder stage
COPY --from=builder /app/manager /manager

# Run as root user to ensure access to /root/.aws
USER root

# Set the entrypoint to execute the binary
ENTRYPOINT ["/manager"]
