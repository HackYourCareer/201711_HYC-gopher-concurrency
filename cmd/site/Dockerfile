FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY ./site /root/site

ENTRYPOINT ["/root/site"]
EXPOSE 5000 5000