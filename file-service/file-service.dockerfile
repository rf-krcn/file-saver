FROM alpine:latest

RUN mkdir /app

COPY fileApp /app

CMD ["/app/fileApp"]
