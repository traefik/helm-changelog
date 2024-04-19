# syntax=docker/dockerfile:1.4
FROM alpine:3

COPY helm-changelog /

CMD ["/helm-changelog"]
