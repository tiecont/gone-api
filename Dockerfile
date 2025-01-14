# Use the official Golang image
FROM golang:1.23.4

# Install Air
RUN export PATH=$PATH:$(go env GOPATH)/bin && go install github.com/air-verse/air@latest

# Copy Air config
COPY .air.toml .

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first for caching dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the code
COPY . .

# Install Air config if not already present
RUN air init || true

# Copy the entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Use the entrypoint script to handle migrations and start the app
ENTRYPOINT ["/app/entrypoint.sh"]