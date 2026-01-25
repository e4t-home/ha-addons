#!/usr/bin/with-contenv bashio

PORT=$(bashio::config 'port')
DB_PATH="/config/device-codes.db"

bashio::log.info "Starting Device Setup Codes on port ${PORT}..."

exec /app/matter-code-db -port "${PORT}" -db "${DB_PATH}"
