# GPSD Proxy

This add-on runs the `gpsd` daemon, connecting to a USB GPS antenna and serving the GPSD protocol on port 2947. This allows the Home Assistant GPSD integration to receive GPS data directly.

## Configuration

### Option: `device`

Select your GPS device from the dropdown. The add-on will discover all available serial/TTY devices. Most USB GPS antennas appear as `/dev/ttyUSB0` or `/dev/ttyACM0`.

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
- u-blox based receivers
- SiRF based receivers
- MediaTek based receivers
- Most USB GPS antennas

## Troubleshooting

### Device not found

1. Check that your GPS is plugged in
2. Restart the add-on after plugging in a new device

### No GPS fix

1. Ensure your GPS antenna has a clear view of the sky
2. First fix can take 1-5 minutes (cold start)
