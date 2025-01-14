# Use the official Golang image
FROM golang:1.23.4

# Set the working directory
WORKDIR /app

# Install Air
RUN export PATH=$PATH:$(go env GOPATH)/bin && go install github.com/air-verse/air@latest

# Copy Air config
COPY .air.toml .

# Copy go.mod and go.sum first for caching dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Create the tmp directory
RUN mkdir -p /app/tmp

# Copy the rest of the code
COPY . .

# Copy the entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Use the entrypoint script to handle migrations and start the app
ENTRYPOINT ["/app/entrypoint.sh"]

