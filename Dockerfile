FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -o github-release-notifier ./cmd 

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/github-release-notifier .

EXPOSE 8080
CMD ["./github-release-notifier"]