FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Create uploads directory
RUN mkdir -p /app/uploads

COPY services/prescription/main .

EXPOSE 8005

CMD ["./main"]
