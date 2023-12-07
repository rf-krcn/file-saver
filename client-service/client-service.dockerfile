FROM alpine:latest

RUN mkdir /app

COPY build /build

COPY clientApp /app

CMD ["/app/clientApp"]
