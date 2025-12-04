FROM alpine:3.20

RUN apk --no-cache add openjdk21 python3 ca-certificates

LABEL authors="valeriovinciarelli"

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /opt/converter

COPY converter-service converter-service

RUN chown -R appuser:appgroup /opt/converter

USER appuser:appgroup

CMD ["./converter-service"]
