package main

import "time"

type DeviceType string

const (
	DeviceTypeHomeKit DeviceType = "homekit"
	DeviceTypeMatter  DeviceType = "matter"
	DeviceTypeRing    DeviceType = "ring"
)

var DeviceTypes = []DeviceType{
	DeviceTypeHomeKit,
	DeviceTypeMatter,
	DeviceTypeRing,
}

type Device struct {
	ID           int64
	Name         string
	Type         DeviceType
	Model        string // e.g., "Ring Stickup Cam", "HM-BSD"
	Manufacturer string // e.g., "Ring", "eQ-3"
	SetupCode    string
	Notes        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
