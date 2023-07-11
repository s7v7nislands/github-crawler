# syntax=docker/dockerfile:1

FROM golang:1.20 AS builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /github-crawler


FROM golang:1.20
WORKDIR /
COPY --from=builder /github-crawler ./

EXPOSE 9090

# Run
CMD ["/github-crawler"]