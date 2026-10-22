FROM golang:1.27-alpine AS builder

COPY . /src
WORKDIR /src

RUN CGO_ENABLED=0 go build -a -tags release -ldflags "-s -w" -o /usr/local/bin/supervisord . \
 && CGO_ENABLED=0 go build -a -ldflags "-s -w" -o /usr/local/bin/pidproxy ./pidproxy

FROM scratch

COPY --from=builder /usr/local/bin/supervisord /usr/local/bin/supervisord
COPY --from=builder /usr/local/bin/pidproxy /usr/local/bin/pidproxy

ENTRYPOINT ["/usr/local/bin/supervisord"]
