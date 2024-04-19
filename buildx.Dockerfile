# syntax=docker/dockerfile:1.4
FROM alpine:3

WORKDIR /data

COPY helm-changelog --chown=1000:1000 /app/helm-changelog

USER 1000:1000

CMD ["/helm-changelog"]
