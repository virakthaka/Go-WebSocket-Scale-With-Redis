FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY bin/app .
COPY views ./views

CMD ["./app"]
