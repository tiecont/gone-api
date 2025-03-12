# Use the official Golang image
FROM golang:1.23.4

RUN apt update && apt install -y default-mysql-client

# Install Air for live reload
RUN go install github.com/air-verse/air@latest && \
    mv $(go env GOPATH)/bin/air /usr/local/bin/

# Set the working directory
WORKDIR /app

# Copy Air config
COPY .air.toml ./

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# Copy the rest of the application
COPY . .

# Ensure Air config exists
RUN air init || true

# Copy the entrypoint script and set permissions
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh && ls -l /app/entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/bin/sh", "-c", "/app/entrypoint.sh"]
