FROM ubuntu:latest
LABEL authors="liliang"

COPY webook /app/webook

WORKDIR /app

ENTRYPOINT ["/app/webook"]