FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY services/timeline/main .

EXPOSE 8004

CMD ["./main"]
