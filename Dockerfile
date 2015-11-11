FROM golang:1.5

ENV RABBITMQ_USER=admin
ENV RABBITMQ_PASS=password
ENV RABBITMQ_PORT_5672_TCP_ADDR=127.0.0.1
ENV RABBITMQ_PORT_5672_TCP_PORT=5672
ENV SERVER_PROTOCOL=https
# By default you don't have to overwrite this but in case that you're getting issues
# by assigning address you can remove comment and use SERVER_IP instead
# ENV SERVER_IP=127.0.0.1
ENV PORT=443
ENV SSL_CERT="/certs/server.pem"
ENV SSL_KEY="/certs/server.key"

RUN mkdir /certs

# @TODO - We should import certificates here ...

EXPOSE 443

ADD . /go/src/microplatform-io/platform-router-http
WORKDIR /go/src/microplatform-io/platform-router-http
RUN go get ./...

ENTRYPOINT ["platform-router-http"]
