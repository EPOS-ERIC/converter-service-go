FROM alpine:3.22@sha256:310c62b5e7ca5b08167e4384c68db0fd2905dd9c7493756d356e893909057601

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
