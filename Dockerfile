FROM golang:1.22.2

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /argo-app-orchestrator

# Expose port
EXPOSE 9000

# Run
CMD ["/argo-app-orchestrator"]