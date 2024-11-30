package main

import (
	_ "embed"
	"image/color"
	"machine"
	"time"

	"github.com/conejoninja/gamelink/gamelink"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ssd1306"
)

const (
	SCREENSAVER = iota
	LAYER
)

var (
	invertRotaryPins = false
	currentLayer     = 0
	displayShowing   = SCREENSAVER
	displayFrame     = 0

	textWhite = color.RGBA{255, 255, 255, 255}
	textBlack = color.RGBA{0, 0, 0, 255}
)

const (
	white = 0x3F3F3FFF
	red   = 0x00FF00FF
	green = 0xFF0000FF
	blue  = 0x0000FFFF
	black = 0x000000FF
)

func main() {
	time.Sleep(3 * time.Second)
	i2c := machine.I2C0
	i2c.Configure(machine.I2CConfig{
		Frequency: 2.8 * machine.MHz,
		SDA:       machine.GPIO12,
		SCL:       machine.GPIO13,
	})

	display := ssd1306.NewI2C(i2c)
	display.Configure(ssd1306.Config{
		Address:  0x3C,
		Width:    128,
		Height:   64,
		Rotation: drivers.Rotation180,
	})
	display.ClearDisplay()

	gl := gamelink.New(i2c)
	gl.Configure()
	//buf, err := gl.Read()
	//println("READ", buf[0], err)

	err := gl.Write([]byte{3, 1, 2, 3})
	println("ERROR", err)

	buf, err := gl.Read()
	println("READ", buf[0], err)

}
