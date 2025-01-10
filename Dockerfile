FROM ubuntu:20.04
LABEL authors="Chis"

COPY webook /app/webook

WORKDIR /app

CMD ["/app/webook"]

