package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// HADevice represents a device from Home Assistant's device registry
type HADevice struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	NameByUser      string   `json:"name_by_user"`
	Manufacturer    string   `json:"manufacturer"`
	Model           string   `json:"model"`
	AreaID          string   `json:"area_id"`
	ConfigEntries   []string `json:"config_entries"`
	Identifiers     [][]any  `json:"identifiers"`
	Connections     [][]any  `json:"connections"`
	SWVersion       string   `json:"sw_version"`
	HWVersion       string   `json:"hw_version"`
	SerialNumber    string   `json:"serial_number"`
	ViaDeviceID     string   `json:"via_device_id"`
	DisabledBy      string   `json:"disabled_by"`
	ConfigurationURL string  `json:"configuration_url"`
}

// HAArea represents an area from Home Assistant
type HAArea struct {
	AreaID string `json:"area_id"`
	Name   string `json:"name"`
}

type Server struct {
	db   *DB
	tmpl *template.Template
}

func NewServer(db *DB) (*Server, error) {
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{db: db, tmpl: tmpl}, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/devices", s.handleDevices)
	mux.HandleFunc("/devices/new", s.handleNewDeviceForm)
	mux.HandleFunc("/devices/search", s.handleSearch)
	mux.HandleFunc("/devices/", s.handleDevice)
	mux.HandleFunc("/ha/devices", s.handleHADevices)

	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	devices, err := s.db.ListDevices()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Devices":     devices,
		"DeviceTypes": DeviceTypes,
	}

	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListDevices(w, r)
	case http.MethodPost:
		s.handleCreateDevice(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDevice(w http.ResponseWriter, r *http.Request) {
	// Parse path: /devices/{id} or /devices/{id}/edit
	path := strings.TrimPrefix(r.URL.Path, "/devices/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	// Handle special routes that might be caught here
	if parts[0] == "search" {
		s.handleSearch(w, r)
		return
	}
	if parts[0] == "new" {
		s.handleNewDeviceForm(w, r)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if len(parts) == 2 && parts[1] == "edit" {
		s.handleEditDeviceForm(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodPut:
		s.handleUpdateDevice(w, r, id)
	case http.MethodDelete:
		s.handleDeleteDevice(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := s.db.ListDevices()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "device-list", devices); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var devices []Device
	var err error

	if query == "" {
		devices, err = s.db.ListDevices()
	} else {
		devices, err = s.db.SearchDevices(query)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "device-list", devices); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (s *Server) handleNewDeviceForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Device":      &Device{},
		"DeviceTypes": DeviceTypes,
		"IsNew":       true,
	}
	if err := s.tmpl.ExecuteTemplate(w, "device-form", data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (s *Server) handleCreateDevice(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	device := &Device{
		Name:         r.FormValue("name"),
		Type:         DeviceType(r.FormValue("type")),
		Model:        r.FormValue("model"),
		Manufacturer: r.FormValue("manufacturer"),
		SetupCode:    r.FormValue("setup_code"),
		Notes:        r.FormValue("notes"),
	}

	if err := s.db.CreateDevice(device); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	devices, err := s.db.ListDevices()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "device-list", devices); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (s *Server) handleEditDeviceForm(w http.ResponseWriter, r *http.Request, id int64) {
	device, err := s.db.GetDevice(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	data := map[string]any{
		"Device":      device,
		"DeviceTypes": DeviceTypes,
		"IsNew":       false,
	}
	if err := s.tmpl.ExecuteTemplate(w, "device-form", data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (s *Server) handleUpdateDevice(w http.ResponseWriter, r *http.Request, id int64) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	device := &Device{
		ID:           id,
		Name:         r.FormValue("name"),
		Type:         DeviceType(r.FormValue("type")),
		Model:        r.FormValue("model"),
		Manufacturer: r.FormValue("manufacturer"),
		SetupCode:    r.FormValue("setup_code"),
		Notes:        r.FormValue("notes"),
	}

	if err := s.db.UpdateDevice(device); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	devices, err := s.db.ListDevices()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "device-list", devices); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (s *Server) handleDeleteDevice(w http.ResponseWriter, r *http.Request, id int64) {
	if err := s.db.DeleteDevice(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	devices, err := s.db.ListDevices()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "device-list", devices); err != nil {
		log.Printf("template error: %v", err)
	}
}

// handleHADevices fetches devices from Home Assistant's device registry
func (s *Server) handleHADevices(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("SUPERVISOR_TOKEN")
	if token == "" {
		http.Error(w, "SUPERVISOR_TOKEN not available - are you running as a Home Assistant add-on?", http.StatusServiceUnavailable)
		return
	}

	// Fetch devices from Home Assistant
	devices, err := s.fetchHADevices(token)
	if err != nil {
		log.Printf("Error fetching HA devices: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch areas to map area IDs to names
	areas, err := s.fetchHAAreas(token)
	if err != nil {
		log.Printf("Error fetching HA areas: %v", err)
		// Continue without areas - not critical
		areas = make(map[string]string)
	}

	// Build response with area names
	type HADeviceWithArea struct {
		HADevice
		AreaName string `json:"area_name"`
	}

	result := make([]HADeviceWithArea, 0, len(devices))
	for _, d := range devices {
		// Skip devices without a name
		name := d.NameByUser
		if name == "" {
			name = d.Name
		}
		if name == "" {
			continue
		}

		dwa := HADeviceWithArea{HADevice: d}
		if d.AreaID != "" {
			dwa.AreaName = areas[d.AreaID]
		}
		result = append(result, dwa)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) fetchHADevices(token string) ([]HADevice, error) {
	req, err := http.NewRequest("POST", "http://supervisor/core/api/config/device_registry/list", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &httpError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	var devices []HADevice
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, err
	}

	return devices, nil
}

func (s *Server) fetchHAAreas(token string) (map[string]string, error) {
	req, err := http.NewRequest("POST", "http://supervisor/core/api/config/area_registry/list", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil // Areas are optional, don't fail
	}

	var areas []HAArea
	if err := json.NewDecoder(resp.Body).Decode(&areas); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, a := range areas {
		result[a.AreaID] = a.Name
	}
	return result, nil
}

type httpError struct {
	StatusCode int
	Message    string
}

func (e *httpError) Error() string {
	return e.Message
}
