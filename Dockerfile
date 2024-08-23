# final stage
FROM alpine:latest

WORKDIR /root/

# Copy the pre-built binary from the pipeline workspace
COPY azflow-api .
COPY .env .
COPY db db

EXPOSE 8080

CMD ["./azflow-api"]
