FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary from host
COPY services/auth/main .

EXPOSE 8001

CMD ["./main"]
