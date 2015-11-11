# Micro platform HTTP Router

HTTP Router is proxy towards gRPC.
You have to run [Platform gRPC Router] in order to use http router.

### How to start router

Certificate used here is just example. Should be used only while testing.

```sh
export RABBITMQ_USER=admin
export RABBITMQ_PASS=admin
export RABBITMQ_PORT_5672_TCP_ADDR=127.0.0.1
export RABBITMQ_PORT_5672_TCP_PORT=5672
export SERVER_PROTOCOL=https

# By default you don't have to overwrite this but in case that you're getting issues
# by assigning address you can remove comment and use SERVER_IP instead
export SERVER_IP=127.0.0.1

export PORT=443
export SSL_CERT="test_data/server.pem"
export SSL_KEY="test_data/server.key"

go build && ./platform-router-http
```

[Platform gRPC Router]: <https://github.com/microplatform-io/platform-router-grpc>
