#!/usr/bin/with-contenv bashio

DEVICE=$(bashio::config 'device')
BAUD_RATE=$(bashio::config 'baud_rate')
FIX_3D_ONLY=$(bashio::config 'fix_3d_only')
LOG_LEVEL=$(bashio::config 'log_level')

bashio::log.info "Starting GPSD Proxy..."
bashio::log.info "Device: ${DEVICE}"
bashio::log.info "Baud Rate: ${BAUD_RATE}"

FIX_3D_FLAG=""
if [ "${FIX_3D_ONLY}" = "true" ]; then
    FIX_3D_FLAG="-fix-3d-only"
fi

exec /usr/local/bin/gpsd-proxy \
    -device "${DEVICE}" \
    -baud "${BAUD_RATE}" \
    -port 2947 \
    -log-level "${LOG_LEVEL}" \
    ${FIX_3D_FLAG}
