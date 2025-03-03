# Use the official Golang image
FROM golang:1.23.4

# Install Air for live reload
RUN export PATH=$PATH:$(go env GOPATH)/bin && go install github.com/air-verse/air@latest

# Set the working directory
WORKDIR /app

# Copy Air config
COPY .air.toml ./

# Copy go.mod and create an empty go.sum if missing
COPY go.mod ./
RUN touch go.sum && go mod tidy && go mod download

# Copy the rest of the code
COPY . .

# Install Air config if not already present
RUN air init || true

# Copy the entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Set the entrypoint script as the default command
ENTRYPOINT ["/app/entrypoint.sh"]
