package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &DB{db}
	if err := d.migrate(); err != nil {
		return nil, err
	}

	return d, nil
}

func (db *DB) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		model TEXT DEFAULT '',
		manufacturer TEXT DEFAULT '',
		setup_code TEXT NOT NULL,
		notes TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(query); err != nil {
		return err
	}

	// Add new columns if they don't exist (for existing databases)
	db.Exec("ALTER TABLE devices ADD COLUMN model TEXT DEFAULT ''")
	db.Exec("ALTER TABLE devices ADD COLUMN manufacturer TEXT DEFAULT ''")

	return nil
}

func (db *DB) ListDevices() ([]Device, error) {
	rows, err := db.Query(`
		SELECT id, name, type, COALESCE(model, ''), COALESCE(manufacturer, ''), setup_code, notes, created_at, updated_at
		FROM devices
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		err := rows.Scan(&d.ID, &d.Name, &d.Type, &d.Model, &d.Manufacturer, &d.SetupCode, &d.Notes, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (db *DB) SearchDevices(query, deviceType string) ([]Device, error) {
	searchTerm := "%" + query + "%"
	
	var rows *sql.Rows
	var err error
	
	if query == "" && deviceType != "" {
		// Only filter by type
		rows, err = db.Query(`
			SELECT id, name, type, COALESCE(model, ''), COALESCE(manufacturer, ''), setup_code, notes, created_at, updated_at
			FROM devices
			WHERE type = ?
			ORDER BY name ASC
		`, deviceType)
	} else if query != "" && deviceType != "" {
		// Filter by both query and type
		rows, err = db.Query(`
			SELECT id, name, type, COALESCE(model, ''), COALESCE(manufacturer, ''), setup_code, notes, created_at, updated_at
			FROM devices
			WHERE (name LIKE ? OR model LIKE ? OR manufacturer LIKE ? OR setup_code LIKE ? OR notes LIKE ?)
			AND type = ?
			ORDER BY name ASC
		`, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, deviceType)
	} else {
		// Only filter by query
		rows, err = db.Query(`
			SELECT id, name, type, COALESCE(model, ''), COALESCE(manufacturer, ''), setup_code, notes, created_at, updated_at
			FROM devices
			WHERE name LIKE ? OR model LIKE ? OR manufacturer LIKE ? OR setup_code LIKE ? OR notes LIKE ?
			ORDER BY name ASC
		`, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		err := rows.Scan(&d.ID, &d.Name, &d.Type, &d.Model, &d.Manufacturer, &d.SetupCode, &d.Notes, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (db *DB) GetDevice(id int64) (*Device, error) {
	var d Device
	err := db.QueryRow(`
		SELECT id, name, type, COALESCE(model, ''), COALESCE(manufacturer, ''), setup_code, notes, created_at, updated_at
		FROM devices
		WHERE id = ?
	`, id).Scan(&d.ID, &d.Name, &d.Type, &d.Model, &d.Manufacturer, &d.SetupCode, &d.Notes, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (db *DB) CreateDevice(d *Device) error {
	now := time.Now()
	result, err := db.Exec(`
		INSERT INTO devices (name, type, model, manufacturer, setup_code, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, d.Name, d.Type, d.Model, d.Manufacturer, d.SetupCode, d.Notes, now, now)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	d.ID = id
	d.CreatedAt = now
	d.UpdatedAt = now
	return nil
}

func (db *DB) UpdateDevice(d *Device) error {
	now := time.Now()
	_, err := db.Exec(`
		UPDATE devices
		SET name = ?, type = ?, model = ?, manufacturer = ?, setup_code = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`, d.Name, d.Type, d.Model, d.Manufacturer, d.SetupCode, d.Notes, now, d.ID)
	if err != nil {
		return err
	}
	d.UpdatedAt = now
	return nil
}

func (db *DB) DeleteDevice(id int64) error {
	_, err := db.Exec(`DELETE FROM devices WHERE id = ?`, id)
	return err
}

func (db *DB) DeviceExists(id int64) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM devices WHERE id = ?)`, id).Scan(&exists)
	return exists, err
}

func (db *DB) CreateDeviceWithID(d *Device) error {
	_, err := db.Exec(`
		INSERT INTO devices (id, name, type, model, manufacturer, setup_code, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.Name, d.Type, d.Model, d.Manufacturer, d.SetupCode, d.Notes, d.CreatedAt, d.UpdatedAt)
	return err
}
