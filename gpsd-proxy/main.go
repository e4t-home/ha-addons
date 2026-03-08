package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/adrianmo/go-nmea"
	"go.bug.st/serial"
)

type GPSDServer struct {
	mu         sync.RWMutex
	clients    map[net.Conn]bool
	tpv        *TPVReport
	sky        *SKYReport
	devicePath string
	baudRate   int
	port       int
	fix3DOnly  bool
	logLevel   string
}

// TPV - Time-Position-Velocity report
type TPVReport struct {
	Class  string   `json:"class"`
	Device string   `json:"device"`
	Mode   int      `json:"mode"`
	Time   string   `json:"time,omitempty"`
	Lat    *float64 `json:"lat,omitempty"`
	Lon    *float64 `json:"lon,omitempty"`
	Alt    *float64 `json:"alt,omitempty"`
	Speed  *float64 `json:"speed,omitempty"`
	Track  *float64 `json:"track,omitempty"`
	Climb  *float64 `json:"climb,omitempty"`
}

// SKY - Satellite information
type SKYReport struct {
	Class      string      `json:"class"`
	Device     string      `json:"device"`
	Satellites []Satellite `json:"satellites,omitempty"`
}

type Satellite struct {
	PRN  int     `json:"PRN"`
	El   float64 `json:"el"`
	Az   float64 `json:"az"`
	SS   float64 `json:"ss"`
	Used bool    `json:"used"`
}

// VERSION response
type VersionReport struct {
	Class      string `json:"class"`
	Release    string `json:"release"`
	Rev        string `json:"rev"`
	ProtoMajor int    `json:"proto_major"`
	ProtoMinor int    `json:"proto_minor"`
}

// DEVICES response
type DevicesReport struct {
	Class   string         `json:"class"`
	Devices []DeviceReport `json:"devices"`
}

type DeviceReport struct {
	Class  string `json:"class"`
	Path   string `json:"path"`
	Driver string `json:"driver"`
	Flags  int    `json:"flags"`
	Native int    `json:"native"`
}

// WATCH command/response
type WatchReport struct {
	Class  string `json:"class"`
	Enable bool   `json:"enable"`
	JSON   bool   `json:"json"`
}

func NewGPSDServer(device string, baud, port int, fix3DOnly bool, logLevel string) *GPSDServer {
	return &GPSDServer{
		clients:    make(map[net.Conn]bool),
		devicePath: device,
		baudRate:   baud,
		port:       port,
		fix3DOnly:  fix3DOnly,
		logLevel:   logLevel,
		tpv: &TPVReport{
			Class:  "TPV",
			Device: device,
			Mode:   0,
		},
		sky: &SKYReport{
			Class:      "SKY",
			Device:     device,
			Satellites: []Satellite{},
		},
	}
}

func (s *GPSDServer) logDebug(format string, v ...interface{}) {
	if s.logLevel == "debug" {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func (s *GPSDServer) logInfo(format string, v ...interface{}) {
	if s.logLevel == "debug" || s.logLevel == "info" {
		log.Printf("[INFO] "+format, v...)
	}
}

func (s *GPSDServer) logWarn(format string, v ...interface{}) {
	if s.logLevel != "error" {
		log.Printf("[WARN] "+format, v...)
	}
}

func (s *GPSDServer) logError(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

func (s *GPSDServer) readGPS() {
	for {
		reader, closer := s.connectSerial()

		if reader == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		s.processNMEAStream(bufio.NewReader(reader))

		if closer != nil {
			closer.Close()
		}

		time.Sleep(2 * time.Second)
	}
}

func (s *GPSDServer) connectSerial() (io.Reader, io.Closer) {
	s.logInfo("Opening serial GPS device: %s at %d baud", s.devicePath, s.baudRate)

	mode := &serial.Mode{
		BaudRate: s.baudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(s.devicePath, mode)
	if err != nil {
		s.logError("Failed to open GPS device: %v", err)
		return nil, nil
	}

	s.logInfo("Serial GPS device opened successfully")
	return port, port
}

func (s *GPSDServer) processNMEAStream(reader *bufio.Reader) {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				s.logError("Error reading from GPS: %v", err)
			}
			return
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "$") {
			continue
		}

		s.logDebug("NMEA: %s", line)
		s.parseNMEA(line)
	}
}

func (s *GPSDServer) parseNMEA(line string) {
	sentence, err := nmea.Parse(line)
	if err != nil {
		s.logDebug("Failed to parse NMEA: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch m := sentence.(type) {
	case nmea.GGA:
		lat := m.Latitude
		lon := m.Longitude
		alt := m.Altitude
		s.tpv.Lat = &lat
		s.tpv.Lon = &lon
		s.tpv.Alt = &alt
		if m.FixQuality != "" && m.FixQuality != "0" {
			s.tpv.Mode = 3 // 3D fix
		} else {
			s.tpv.Mode = 1 // No fix
		}
		s.tpv.Time = formatGPSTime(m.Time)

	case nmea.RMC:
		lat := m.Latitude
		lon := m.Longitude
		speed := m.Speed * 0.514444 // knots to m/s
		track := m.Course
		s.tpv.Lat = &lat
		s.tpv.Lon = &lon
		s.tpv.Speed = &speed
		s.tpv.Track = &track
		if m.Validity == "A" {
			if s.tpv.Mode < 2 {
				s.tpv.Mode = 2 // 2D fix
			}
		} else {
			s.tpv.Mode = 1
		}
		s.tpv.Time = formatGPSDateTime(m.Date, m.Time)

	case nmea.GSA:
		switch m.FixType {
		case "1":
			s.tpv.Mode = 1
		case "2":
			s.tpv.Mode = 2
		case "3":
			s.tpv.Mode = 3
		}

	case nmea.GSV:
		sats := make([]Satellite, 0)
		for _, sv := range m.Info {
			sats = append(sats, Satellite{
				PRN: int(sv.SVPRNNumber),
				El:  float64(sv.Elevation),
				Az:  float64(sv.Azimuth),
				SS:  float64(sv.SNR),
			})
		}
		if len(sats) > 0 {
			s.sky.Satellites = sats
		}

	case nmea.VTG:
		speed := m.GroundSpeedKPH / 3.6 // km/h to m/s
		track := m.TrueTrack
		s.tpv.Speed = &speed
		s.tpv.Track = &track
	}

	s.broadcastUpdates()
}

func formatGPSTime(t nmea.Time) string {
	return fmt.Sprintf("%02d:%02d:%02d.%03dZ",
		t.Hour, t.Minute, t.Second, t.Millisecond)
}

func formatGPSDateTime(d nmea.Date, t nmea.Time) string {
	return fmt.Sprintf("20%02d-%02d-%02dT%02d:%02d:%02d.%03dZ",
		d.YY, d.MM, d.DD, t.Hour, t.Minute, t.Second, t.Millisecond)
}

func (s *GPSDServer) shouldBroadcast() bool {
	if s.fix3DOnly && s.tpv.Mode < 3 {
		return false
	}
	return true
}

func (s *GPSDServer) broadcastUpdates() {
	if !s.shouldBroadcast() {
		return
	}

	tpvJSON, _ := json.Marshal(s.tpv)

	for conn, watching := range s.clients {
		if watching {
			conn.Write(append(tpvJSON, '\n'))
		}
	}
}

func (s *GPSDServer) handleClient(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	s.logInfo("Client connected: %s", conn.RemoteAddr())

	s.mu.Lock()
	s.clients[conn] = false
	s.mu.Unlock()

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				s.logDebug("Client read error: %v", err)
			}
			s.logInfo("Client disconnected: %s", conn.RemoteAddr())
			return
		}

		line = strings.TrimSpace(line)
		s.logDebug("Received from client: %s", line)
		s.handleCommand(conn, line)
	}
}

func (s *GPSDServer) handleCommand(conn net.Conn, cmd string) {
	switch {
	case strings.HasPrefix(cmd, "?VERSION"):
		version := VersionReport{
			Class:      "VERSION",
			Release:    "gpsd-proxy 1.0.1",
			Rev:        "1.0.1",
			ProtoMajor: 3,
			ProtoMinor: 14,
		}
		s.sendJSON(conn, version)

	case strings.HasPrefix(cmd, "?DEVICES"):
		devices := DevicesReport{
			Class: "DEVICES",
			Devices: []DeviceReport{
				{
					Class:  "DEVICE",
					Path:   s.devicePath,
					Driver: "NMEA0183",
					Flags:  1,
					Native: 0,
				},
			},
		}
		s.sendJSON(conn, devices)

	case strings.HasPrefix(cmd, "?WATCH"):
		watch := WatchReport{
			Class:  "WATCH",
			Enable: true,
			JSON:   true,
		}

		if strings.Contains(cmd, "=") {
			parts := strings.SplitN(cmd, "=", 2)
			if len(parts) == 2 {
				var watchCmd struct {
					Enable bool `json:"enable"`
					JSON   bool `json:"json"`
				}
				if err := json.Unmarshal([]byte(parts[1]), &watchCmd); err == nil {
					watch.Enable = watchCmd.Enable
					watch.JSON = watchCmd.JSON
				}
			}
		}

		s.mu.Lock()
		s.clients[conn] = watch.Enable
		s.mu.Unlock()

		s.sendJSON(conn, watch)

		devices := DevicesReport{
			Class: "DEVICES",
			Devices: []DeviceReport{
				{
					Class:  "DEVICE",
					Path:   s.devicePath,
					Driver: "NMEA0183",
					Flags:  1,
					Native: 0,
				},
			},
		}
		s.sendJSON(conn, devices)

		s.mu.RLock()
		if s.tpv.Mode > 0 && s.shouldBroadcast() {
			s.sendJSON(conn, s.tpv)
		}
		if len(s.sky.Satellites) > 0 {
			s.sendJSON(conn, s.sky)
		}
		s.mu.RUnlock()

	case strings.HasPrefix(cmd, "?POLL"):
		s.mu.RLock()
		if s.shouldBroadcast() {
			s.sendJSON(conn, s.tpv)
		}
		s.sendJSON(conn, s.sky)
		s.mu.RUnlock()

	default:
		s.logDebug("Unknown command: %s", cmd)
	}
}

func (s *GPSDServer) sendJSON(conn net.Conn, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		s.logError("JSON marshal error: %v", err)
		return
	}
	conn.Write(append(data, '\n'))
}

func (s *GPSDServer) Start() error {
	go s.readGPS()

	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.logInfo("GPSD server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logError("Accept error: %v", err)
			continue
		}
		go s.handleClient(conn)
	}
}

func main() {
	device := flag.String("device", "/dev/ttyUSB0", "Serial GPS device path")
	baud := flag.Int("baud", 9600, "Baud rate for serial")
	port := flag.Int("port", 2947, "GPSD server port")
	fix3DOnly := flag.Bool("fix-3d-only", false, "Only broadcast 3D fixes")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)

	server := NewGPSDServer(*device, *baud, *port, *fix3DOnly, *logLevel)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
