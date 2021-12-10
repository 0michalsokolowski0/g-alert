FROM golang:1.17-alpine
WORKDIR /app
RUN apk update && apk add wget
RUN apk add \
  --no-cache \
  --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing \
  --repository http://dl-cdn.alpinelinux.org/alpine/edge/main \
  googler

COPY g-alert    /
CMD ["/g-alert"]