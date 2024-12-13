// Package gamelink provides a driver for the Keeb GameLink
package gamelink

import (
	"tinygo.org/x/drivers"
)

const Address = 0x76

// Device wraps an I2C connection to a GameLink device.
type Device struct {
	bus     drivers.I2C
	Address uint16
}

// New returns a GameLink Device
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

// Configure does nothing (for the moment)
func (d *Device) Configure() {
	println("GL configure")
}

// Write sends a message to the GameLink to be read later by the other device connected to
func (d *Device) Write(data []uint8) error {
	return d.bus.Tx(d.Address, data, nil)
}

// Read returns a message from the GameLink sent by the other device connected to
func (d *Device) Read() ([]uint8, error) {
	data := make([]uint8, 10)
	err := d.bus.Tx(d.Address, nil, data)
	return data, err
}
