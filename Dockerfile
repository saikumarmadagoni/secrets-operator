# Start with Amazon Linux 2023 as the base image for compatibility with AWS SDKs
FROM amazonlinux:2023 AS base

# Install necessary dependencies
RUN yum update -y && \
    yum install -y \
    gcc \
    tar \
    gzip \
    make \
    git \
    aws-cli \
    # Clean up
    && yum clean all

# Install Go (replace with desired version)
ARG GOLANG_VERSION=1.21.0
RUN curl -OL https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm -f go${GOLANG_VERSION}.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

# Build the Go application
WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Download dependencies
RUN go mod download

# Copy the Go source code
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/

# Build the application binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager cmd/main.go

# Final stage: Use Amazon Linux as the runtime environment
FROM amazonlinux:2023 AS runtime

# Add AWS CLI and necessary packages
RUN yum update -y && \
    yum install -y aws-cli && \
    yum clean all

# Set up the application user and directories
RUN useradd -u 1000 appuser && mkdir -p /app && chown appuser:appuser /app

# Set the working directory
WORKDIR /app

# Copy the Go application binary from the build stage
COPY --from=base /workspace/manager /app/manager

# Ensure permissions for the AWS credentials directory
RUN mkdir -p /root/.aws && chown -R appuser:appuser /root/.aws

# Switch to non-root user
USER appuser

# Entrypoint for the application
ENTRYPOINT ["/app/manager"]
