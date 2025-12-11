FROM golang:1.22-bookworm as builder

WORKDIR /app
COPY . .

RUN make build

FROM alpine:3.16.2
COPY --from=builder /app/bin/consul-alerts /bin/consul-alerts

ENV CONSUL_VERSION=0.8.5
RUN apk upgrade --update --no-cache; \
    apk add --update --no-cache curl util-linux; \
    curl -sSLo /tmp/consul.zip https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_amd64.zip; \
    unzip -d /bin /tmp/consul.zip; \
    rm /tmp/consul.zip; \
    apk del curl

EXPOSE 9000
CMD []
ENTRYPOINT [ "/bin/consul-alerts", "--alert-addr=0.0.0.0:9000" ]
