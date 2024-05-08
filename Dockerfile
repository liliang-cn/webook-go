FROM ubuntu:latest
LABEL authors="leoleecn"

COPY webook /app/webook

WORKDIR /app

ENTRYPOINT ["/app/webook"]