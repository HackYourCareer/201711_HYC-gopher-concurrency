FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY ./attacker /root/attacker

ENTRYPOINT ["/root/attacker"]