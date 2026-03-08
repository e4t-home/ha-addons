#!/usr/bin/with-contenv bashio

DEVICE_TYPE=$(bashio::config 'device_type')
DEVICE=$(bashio::config 'device')
TCP_HOST=$(bashio::config 'tcp_host')
TCP_PORT=$(bashio::config 'tcp_port')
BAUD_RATE=$(bashio::config 'baud_rate')
FIX_3D_ONLY=$(bashio::config 'fix_3d_only')
LOG_LEVEL=$(bashio::config 'log_level')

bashio::log.info "Starting GPSD Proxy..."
bashio::log.info "Device Type: ${DEVICE_TYPE}"

if [ "${DEVICE_TYPE}" = "serial" ]; then
    bashio::log.info "Device: ${DEVICE}"
    bashio::log.info "Baud Rate: ${BAUD_RATE}"
fi

if [ "${DEVICE_TYPE}" = "tcp" ]; then
    bashio::log.info "TCP Host: ${TCP_HOST}"
    bashio::log.info "TCP Port: ${TCP_PORT}"
fi

FIX_3D_FLAG=""
if [ "${FIX_3D_ONLY}" = "true" ]; then
    FIX_3D_FLAG="-fix-3d-only"
fi

exec /usr/local/bin/gpsd-proxy \
    -device-type "${DEVICE_TYPE}" \
    -device "${DEVICE}" \
    -tcp-host "${TCP_HOST}" \
    -tcp-port "${TCP_PORT}" \
    -baud "${BAUD_RATE}" \
    -port 2947 \
    -log-level "${LOG_LEVEL}" \
    ${FIX_3D_FLAG}
