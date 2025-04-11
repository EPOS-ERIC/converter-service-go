FROM alpine:3.20
RUN apk --no-cache add openjdk17 python3
# FROM amazoncorretto:17-alpine-jdk
# RUN apk --no-cache add python3

LABEL authors="valeriovinciarelli"

WORKDIR /opt/converter

COPY converter-service converter-service

CMD ["./converter-service"]
