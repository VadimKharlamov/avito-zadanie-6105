FROM golang:latest as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp ./cmd/tender-service

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/myapp .
COPY --from=builder /app/config.env .
EXPOSE 8080:8080
CMD ["./myapp"]
