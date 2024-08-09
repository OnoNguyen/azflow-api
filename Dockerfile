# Dockerfile
FROM golang:1.20 as builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o azflow-api .

# final stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/azflow-api .

EXPOSE 8080

CMD ["./azflow-api"]
