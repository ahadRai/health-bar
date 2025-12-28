FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY services/patient/main .

EXPOSE 8002

CMD ["./main"]
