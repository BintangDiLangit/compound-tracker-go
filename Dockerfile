# Use Golang image as base
FROM golang:1.22

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Install dependencies
RUN go mod download

# Copy the rest of the app
COPY . .

# Build the app
RUN go build -o main .

# Expose port for HTTP server
EXPOSE 8080

# Run the app
CMD ["./main"]