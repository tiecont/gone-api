# Use the official Golang image
FROM golang:1.21.1

# Install Air
RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b /usr/local/bin

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

# Expose the application port
EXPOSE 2000

# Command to run Air
CMD ["air"]
