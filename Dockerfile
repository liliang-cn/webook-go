FROM golang:1.22

COPY webook /app/webook

WORKDIR /app

ENTRYPOINT ["/app/webook"]