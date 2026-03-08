#!/usr/bin/with-contenv bashio

DEVICE=$(bashio::config 'device')

bashio::log.info "Starting gpsd on device ${DEVICE}..."

exec gpsd -N -n -G -S 2947 "${DEVICE}"
