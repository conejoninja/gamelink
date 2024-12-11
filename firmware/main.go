package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

const (
	PORTA = iota
	PORTB
)

const (
	portAddress = 0x76
	stackSize   = 20
	dataSize    = 10
)

type Message struct {
	read bool
	data [dataSize]byte
}

type Stack struct {
	stack    [stackSize]Message
	writePtr byte
	readPtr  byte
}

var (
	ports = []*machine.I2C{
		machine.I2C0,
		machine.I2C1,
	}
	pinSDA = []machine.Pin{
		machine.D0,
		machine.D2,
	}
	pinSCL = []machine.Pin{
		machine.D1,
		machine.D3,
	}
	err    error
	stacks [2]Stack

	led [1]color.RGBA
)

func main() {

	time.Sleep(3 * time.Second)

	machine.WS2812.Configure(machine.PinConfig{Mode: machine.PinOutput})

	ws := ws2812.NewWS2812(machine.WS2812)

	for p := range ports {
		err = ports[p].Configure(machine.I2CConfig{
			Frequency: 2.8 * machine.MHz,
			Mode:      machine.I2CModeTarget,
			SDA:       pinSDA[p],
			SCL:       pinSCL[p],
		})
		if err != nil {
			led[0] = color.RGBA{0xFF, 0x00, 0x00, 0xFF}
			ws.WriteColors(led[:])
			panic("failed to config I2C0 as controller")
		}

		err = ports[p].Listen(portAddress)
		if err != nil {
			led[0] = color.RGBA{0x00, 0x00, 0xFF, 0xFF}
			ws.WriteColors(led[:])
			panic("failed to listen as I2C target")
		}

		for s := range stackSize {
			stacks[p].stack[s].read = true
		}
	}
	println("GOING TO LISTEN")
	go portListener(PORTA)
	go portListener(PORTB)
	led[0] = color.RGBA{0x00, 0xFF, 0x00, 0xFF}
	ws.WriteColors(led[:])

	for {
		time.Sleep(1 * time.Second)
	}

}

func portListener(port byte) {
	println("LISTENING ON ", port)
	buf := make([]byte, 10)

	for {
		evt, n, err := ports[port].WaitForEvent(buf)
		if err != nil {
			println("FOR ERROR", err)
		}

		switch evt {
		case machine.I2CReceive:
			println("RECEIVED", port, buf[0], n)

			if n > dataSize {
				n = dataSize
			}
			stacks[port].writePtr = (stacks[port].writePtr + 1) % stackSize
			for o := 0; o < n; o++ {
				println("RECEIVED=", buf[o], stacks[port].writePtr, port)
				stacks[port].stack[stacks[port].writePtr].data[o] = buf[o]
			}
			for o := n; o < dataSize; o++ {
				stacks[port].stack[stacks[port].writePtr].data[o] = 0
			}
			stacks[port].stack[stacks[port].writePtr].read = false

		case machine.I2CRequest:
			portClient := (port + 1) % 2
			ptr := (stacks[portClient].readPtr + 1) % stackSize
			if stacks[portClient].stack[ptr].read {
				ports[port].Reply([]byte{0})
				continue
			}
			println("REQUESTED", port, portClient, stacks[portClient].stack[ptr].data[:], ptr)
			stacks[portClient].readPtr = ptr
			ports[port].Reply(stacks[portClient].stack[ptr].data[:])
			stacks[portClient].stack[ptr].read = true

		case machine.I2CFinish:

		default:
			println("default")
		}
	}
}
