FROM golang:1.22-alpine3.19 AS builder

COPY ${PWD} /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -o factorio-relay relay.go
RUN apk add --no-cache libcap
RUN setcap cap_net_raw+pe /app/factorio-relay

FROM alpine:3.19
LABEL MAINTAINER Nicholas <nicholas@peyta.com>
ENV USER=udp-relay
ENV RELAYS=""
RUN adduser -D ${USER}
USER ${USER}
WORKDIR /home/${USER}
COPY --from=builder /app/factorio-relay /usr/bin/factorio-relay
ENTRYPOINT [ "/usr/bin/factorio-relay" ]
CMD [ "--relays=RELAYS" ]