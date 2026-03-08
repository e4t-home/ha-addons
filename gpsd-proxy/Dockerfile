ARG BUILD_FROM=ghcr.io/home-assistant/amd64-base:3.18
FROM ${BUILD_FROM}

RUN apk add --no-cache gpsd

# s6-overlay v3 service structure
COPY run.sh /etc/s6-overlay/s6-rc.d/gpsd/run
RUN chmod a+x /etc/s6-overlay/s6-rc.d/gpsd/run \
    && echo "longrun" > /etc/s6-overlay/s6-rc.d/gpsd/type \
    && touch /etc/s6-overlay/s6-rc.d/user/contents.d/gpsd
