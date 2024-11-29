package main

import (
	"machine"
	"time"
)

const (
	PORTA = iota
	PORTB
)

const (
	portAddress = 0x76
	maxTxSize   = 16
)

var (
	ports = []*machine.I2C{
		machine.I2C0,
		//machine.I2C1,
	}
	err error
	mem [2][256]byte
)

func main() {

	time.Sleep(3 * time.Second)

	for p := range ports {
		err = ports[p].Configure(machine.I2CConfig{
			Mode: machine.I2CModeTarget,
			SDA:  machine.D0,
			SCL:  machine.D1,
		})
		if err != nil {
			panic("failed to config I2C0 as controller")
		}

		err = ports[p].Listen(portAddress)
		if err != nil {
			panic("failed to listen as I2C target")
		}
	}
	println("GOING TO LISTEN")
	portListener(PORTA)
	//go portListener(PORTB)

	for {
		time.Sleep(1 * time.Second)
	}

}

func portListener(port byte) {
	println("LISTENING N ")
	println("LISTENING ON ", port)
	buf := make([]byte, 1)
	var ptr uint8

	for {
		println("F", port)
		evt, n, err := ports[port].WaitForEvent(buf)
		println("G", evt, n, err)
		if err != nil {
			println("FOR ERROR", err)
		}

		switch evt {
		case machine.I2CReceive:
			println("RECEIVED", port, n)
			if n > 0 {
				ptr = buf[0]
			}
			println("RECEIVED_", buf[0])

			for o := 1; o < n; o++ {
				println("RECEIVED=", buf[o])
				mem[port][ptr] = buf[o]
				ptr++
			}

		case machine.I2CRequest:
			println("REQUESTED", port, mem[port][ptr:256])
			ports[port].Reply(mem[port][ptr:256])

		case machine.I2CFinish:
			// nothing to do
			println("I2C FINISH")

		default:
			println("default")
		}
	}
}
