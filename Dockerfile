# Start with an Amazon Linux base image for AWS SDK compatibility
FROM amazonlinux:2 AS builder

# Install Go
RUN curl -LO https://golang.org/dl/go1.21.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz && \
    rm go1.21.0.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

# Set up the working directory
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the remaining source code files
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager cmd/main.go

# Final stage: Use Amazon Linux as the runtime environment
FROM amazonlinux:2
WORKDIR /

# Copy the compiled binary from the build stage
COPY --from=builder /app/manager /manager

# Run as root user to ensure access to /root/.aws
USER root

# Set entrypoint to run the application
ENTRYPOINT ["/manager"]
