# syntax=docker/dockerfile:1.4
FROM alpine:3

WORKDIR /data

RUN apk add git

COPY --chown=1000:1000 helm-changelog /app/helm-changelog

USER 1000:1000

CMD ["/app/helm-changelog"]
