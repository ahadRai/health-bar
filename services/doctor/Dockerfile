FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY services/doctor/main .

EXPOSE 8003

CMD ["./main"]
