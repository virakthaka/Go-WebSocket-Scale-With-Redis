FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY bin/app .

CMD ["./app"]
