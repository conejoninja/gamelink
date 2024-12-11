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

func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

func (d *Device) Configure() {
	println("GL configure")
}

func (d *Device) Write(data []uint8) error {
	println("GL write")
	return d.bus.Tx(d.Address, data, nil)
}

func (d *Device) Read() ([]uint8, error) {
	println("GL read")
	data := make([]uint8, 10)
	err := d.bus.Tx(d.Address, []byte{1}, data)
	return data, err
}
