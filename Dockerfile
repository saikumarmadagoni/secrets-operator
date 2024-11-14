# Start with an Amazon Linux base image for AWS SDK compatibility
FROM amazonlinux:2 AS base

# Install necessary packages for running your Go application
RUN yum install -y shadow-utils && \
    yum clean all

# Set up the working directory and user
WORKDIR /app

# Copy the Go Modules manifests and source code
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager cmd/main.go

# Final stage: Use Amazon Linux as the runtime environment
FROM amazonlinux:2
WORKDIR /

# Copy the compiled binary from the build stage
COPY --from=base /app/manager /manager

# Run as root user to ensure access to /root/.aws
USER root

# Set entrypoint to run the application
ENTRYPOINT ["/manager"]
