# GPSD Proxy

This add-on connects to a USB GPS antenna (or network GPS) and serves the GPSD protocol on port 2947, allowing the Home Assistant GPSD integration to receive GPS data directly.

## Configuration

### Option: `device_type`

Select the connection type:
- **serial** - Direct attached GPS dongle via USB or similar (most common)
- **tcp** - Network connected GPS devices (fill in TCP Host and TCP Port)

### Option: `device`

Select your GPS device from the dropdown. The add-on will discover all available serial/TTY devices.

For TCP connections, this setting is ignored but a device must still be selected.

### Option: `tcp_host`

Hostname or IP address of the network GPS device. Only used when device type is **tcp**.

### Option: `tcp_port`

Port number for the network GPS device. Common values:
- `10110` - NMEA over TCP
- `2947` - GPSD protocol

Only used when device type is **tcp**.

### Option: `baud_rate`

Serial communication speed. Common values:
- `4800` - Older GPS receivers
- `9600` - Most common default
- `38400` - Some newer receivers
- `115200` - High-speed receivers

### Option: `fix_3d_only`

When enabled, only publishes position updates when the GPS has a 3D fix (minimum 3 satellites for accurate positioning). Recommended to leave enabled to avoid publishing inaccurate positions.

### Option: `log_level`

Controls logging verbosity:
- `debug` - All messages including raw NMEA sentences
- `info` - Normal operation messages
- `warn` - Warnings only
- `error` - Errors only

## Home Assistant Integration

After installing this add-on, configure the GPSD integration in Home Assistant:

1. Go to **Settings** > **Devices & Services**
2. Click **Add Integration**
3. Search for **GPSD**
4. Enter:
   - **Host**: `localhost` or your Home Assistant IP
   - **Port**: `2947`

## Supported GPS Receivers

Any GPS receiver that outputs NMEA sentences should work:
- u-blox based receivers (like Nabu Casa SkyConnect with GPS)
- SiRF based receivers
- MediaTek based receivers
- Most USB GPS antennas

## Troubleshooting

### Device not found

1. Check that your GPS is plugged in
2. Restart the add-on after plugging in a new device
3. Check the add-on logs for connection errors

### No GPS fix

1. Ensure your GPS antenna has a clear view of the sky
2. First fix can take 1-5 minutes (cold start)
3. Enable debug logging to see NMEA sentences being received
4. Make sure the baud rate matches your GPS device
