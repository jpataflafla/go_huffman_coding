# Use the official Golang image
FROM golang:1.22.1

# Set the working directory
WORKDIR /app

# Create a directory for the application source code
RUN mkdir -p /command-encoding-service

# Set the working directory for the source code
WORKDIR /command-encoding-service

# Copy the entire application source code into the container
COPY . .

# Download dependencies and build the binary
RUN go mod download
RUN go build -o bin/command-encoding-service

# Expose port
EXPOSE 3000


# Command to run the application
CMD ["./bin/command-encoding-service"] 
#, "-b", "0.0.0.0"]