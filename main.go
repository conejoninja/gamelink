package main

import (
	_ "embed"
	"image/color"
	"machine"
	"time"

	pio "github.com/tinygo-org/pio/rp2-pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"

	"github.com/conejoninja/gamelink/gamelink"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/encoders"
	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

const (
	SCREENSAVER = iota
	LAYER
)

const (
	MENU = iota
	GAME_START
	GAME_WAIT_OTHER
	GAME_WAIT_KEY
)

var (
	invertRotaryPins = false
	currentLayer     = 0
	displayShowing   = SCREENSAVER
	displayFrame     = 0

	textWhite = color.RGBA{255, 255, 255, 255}
	textBlack = color.RGBA{0, 0, 0, 255}

	rotaryOldValue, rotaryNewValue int

	state    = MENU
	hostGame = false

	colPins = []machine.Pin{
		machine.GPIO5,
		machine.GPIO6,
		machine.GPIO7,
		machine.GPIO8,
	}

	rowPins = []machine.Pin{
		machine.GPIO9,
		machine.GPIO10,
		machine.GPIO11,
	}

	matrixBtn [12]bool
)

const (
	white = 0x3F3F3FFF
	red   = 0x00FF00FF
	green = 0xFF0000FF
	blue  = 0x0000FFFF
	black = 0x000000FF
)

type WS2812B struct {
	Pin machine.Pin
	ws  *piolib.WS2812B
}

func NewWS2812B(pin machine.Pin) *WS2812B {
	s, _ := pio.PIO0.ClaimStateMachine()
	ws, _ := piolib.NewWS2812B(s, pin)
	ws.EnableDMA(true)
	return &WS2812B{
		ws: ws,
	}
}

func (ws *WS2812B) WriteRaw(rawGRB []uint32) error {
	return ws.ws.WriteRaw(rawGRB)
}

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

	enc := encoders.NewQuadratureViaInterrupt(
		machine.GPIO4,
		machine.GPIO3,
	)

	enc.Configure(encoders.QuadratureConfig{
		Precision: 4,
	})
	rotaryBtn := machine.GPIO2
	rotaryBtn.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	for _, c := range colPins {
		c.Configure(machine.PinConfig{Mode: machine.PinOutput})
		c.Low()
	}

	for _, c := range rowPins {
		c.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	}

	colors := []uint32{
		black, black, black, black,
		black, black, black, black,
		black, black, black, black,
	}
	ws := NewWS2812B(machine.GPIO1)

	menuOption := 0
	pressed := -1

	for {
		display.ClearBuffer()

		getMatrixState()

		switch state {
		case MENU:

			if rotaryNewValue = enc.Position(); rotaryNewValue != rotaryOldValue {
				println("value: ", rotaryNewValue)
				if rotaryNewValue > rotaryOldValue {
					menuOption = 1
				} else {
					menuOption = 0
				}
				rotaryOldValue = rotaryNewValue
			}

			if menuOption == 0 {
				tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 10, 20, "[+] HOST GAME", textWhite)
				tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 10, 34, "[ ] JOIN GAME", textWhite)
			} else {
				tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 10, 20, "[ ] HOST GAME", textWhite)
				tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 10, 34, "[+] JOIN GAME", textWhite)
			}

			if !rotaryBtn.Get() {
				println("pressed")
				if menuOption == 0 {
					state = GAME_WAIT_KEY
					hostGame = true
				} else {
					state = GAME_WAIT_OTHER
					hostGame = false
				}
			}

			break
		case GAME_WAIT_KEY:
			pressed = -1
			for i := 0; i < 12; i++ {
				if matrixBtn[i] {
					pressed = i
					break
				}
			}
			if pressed == -1 {
				tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 10, 20, "Press any key", textWhite)
				break
			}

			if colors[pressed] != black {
				tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 10, 20, "Invalid key", textWhite)
				break
			}

			if hostGame {
				colors[pressed] = red
			} else {
				colors[pressed] = blue
			}
			gl.Write([]uint8{GAME_WAIT_KEY, uint8(pressed)})
			state = GAME_WAIT_OTHER
			break
		case GAME_WAIT_OTHER:
			buffer, err := gl.Read()
			println("WAITING", err, len(buffer), buffer[0])
			tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 10, 20, "Waiting for other", textWhite)
			tinyfont.WriteLine(&display, &proggy.TinySZ8pt7b, 10, 34, "player's move", textWhite)
			break
		}

		ws.WriteRaw(colors)
		display.Display()
		time.Sleep(100 * time.Millisecond)
	}

	//buf, err := gl.Read()
	//println("READ", buf[0], err)

	err := gl.Write([]byte{3, 1, 2, 3})
	println("ERROR", err)

	buf, err := gl.Read()
	println("READ", buf[0], len(buf), err)

}

func getMatrixState() {
	colPins[0].High()
	colPins[1].Low()
	colPins[2].Low()
	colPins[3].Low()
	time.Sleep(1 * time.Millisecond)

	matrixBtn[0] = rowPins[0].Get()
	matrixBtn[1] = rowPins[1].Get()
	matrixBtn[2] = rowPins[2].Get()

	// COL2
	colPins[0].Low()
	colPins[1].High()
	colPins[2].Low()
	colPins[3].Low()
	time.Sleep(1 * time.Millisecond)

	matrixBtn[3] = rowPins[0].Get()
	matrixBtn[4] = rowPins[1].Get()
	matrixBtn[5] = rowPins[2].Get()

	// COL3
	colPins[0].Low()
	colPins[1].Low()
	colPins[2].High()
	colPins[3].Low()
	time.Sleep(1 * time.Millisecond)

	matrixBtn[6] = rowPins[0].Get()
	matrixBtn[7] = rowPins[1].Get()
	matrixBtn[8] = rowPins[2].Get()

	// COL4
	colPins[0].Low()
	colPins[1].Low()
	colPins[2].Low()
	colPins[3].High()
	time.Sleep(1 * time.Millisecond)

	matrixBtn[9] = rowPins[0].Get()
	matrixBtn[10] = rowPins[1].Get()
	matrixBtn[11] = rowPins[2].Get()
}
