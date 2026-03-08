ARG BUILD_FROM=ghcr.io/home-assistant/amd64-base:3.18
FROM ${BUILD_FROM}

RUN apk add --no-cache gpsd

COPY run.sh /run.sh
RUN chmod a+x /run.sh

CMD ["/run.sh"]
