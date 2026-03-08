ARG BUILD_FROM=ghcr.io/home-assistant/amd64-base:3.18
FROM ${BUILD_FROM}

RUN apk add --no-cache gpsd

COPY run.sh /etc/services.d/gpsd/run
RUN chmod a+x /etc/services.d/gpsd/run
