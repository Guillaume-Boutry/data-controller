# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.15 as builder

# Create and change to the app directory.
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build ./cmd/data-controller

RUN CGO_ENABLED=0 GOOS=linux go build -v ./cmd/data-controller

FROM alpine:3
RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/data-controller /data-controller

# Run the web service on container startup.
CMD ["/data-controller"]