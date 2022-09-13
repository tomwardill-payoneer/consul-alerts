FROM golang:1.19.1-buster as builder

WORKDIR /app
COPY . .

RUN make build

FROM consul:1.13.1 as consul

FROM alpine:3.16.2
COPY --from=builder /app/bin/consul-alerts /bin/consul-alerts
COPY --from=consul /consul /bin/consul
EXPOSE 9000
CMD []
ENTRYPOINT [ "/bin/consul-alerts", "--alert-addr=0.0.0.0:9000" ]
