FROM golang:1.17.6-stretch
WORKDIR /app
RUN apt-get update -y && apt-get upgrade -y
RUN apt-get install googler -y

COPY config.yml /etc/g-alert/config.yml
ENV CONFIG_PATH="/etc/g-alert/config.yml"

COPY . /usr/src/app
WORKDIR /usr/src/app
RUN make build
CMD ["./g-alert"]