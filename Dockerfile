FROM alpine:3.24@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

RUN apk --no-cache add openjdk21 python3 ca-certificates && apk upgrade --no-cache

LABEL authors="valeriovinciarelli"

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /opt/converter

COPY converter-service converter-service

RUN mkdir /opt/converter/plugins

RUN chown -R appuser:appgroup /opt/converter

USER appuser:appgroup

CMD ["./converter-service"]
