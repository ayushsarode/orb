FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o orb cmd/orb/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o orbhub cmd/orbhub/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/orb .
COPY --from=builder /app/orbhub .
CMD ["./orb"]