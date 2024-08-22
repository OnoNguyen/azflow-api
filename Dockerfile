# final stage
FROM --platform=linux/arm64 alpine:latest

WORKDIR /root/

# Copy the pre-built binary from the pipeline workspace
COPY azflow-api .

EXPOSE 8080

CMD ["./azflow-api"]
