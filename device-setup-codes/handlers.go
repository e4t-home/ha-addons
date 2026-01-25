package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

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
