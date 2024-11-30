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

type Message struct {
	read bool
	data [10]byte
}

type Stack struct {
	stack [20]Message
	ptr   byte
}

var (
	ports = []*machine.I2C{
		machine.I2C0,
		machine.I2C1,
	}
	err    error
	stacks [2]Stack
)

func main() {

	time.Sleep(3 * time.Second)

	for p := range ports {
		err = ports[p].Configure(machine.I2CConfig{
			Frequency: 2.8 * machine.MHz,
			Mode:      machine.I2CModeTarget,
			SDA:       machine.D0,
			SCL:       machine.D1,
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
	go portListener(PORTA)
	go portListener(PORTB)

	for {
		time.Sleep(1 * time.Second)
	}

}

func portListener(port byte) {
	println("LISTENING ON ", port)
	buf := make([]byte, 1)

	for {
		println("F", port)
		evt, n, err := ports[port].WaitForEvent(buf)
		println("G", evt, n, err)
		if err != nil {
			println("FOR ERROR", err)
		}

		switch evt {
		case machine.I2CReceive:
			println("RECEIVED", port, buf[0], n)

			if n > 10 {
				n = 10
			}
			stacks[port].ptr = (stacks[port].ptr + 1) % 20
			for o := 0; o < n; o++ {
				println("RECEIVED=", buf[o])
				stacks[port].stack[stacks[port].ptr].data[o] = buf[o]
			}
			for o := n; o < 10; o++ {
				stacks[port].stack[stacks[port].ptr].data[o] = 0
			}

		case machine.I2CRequest:
			portClient := (port + 1) % 2
			println("REQUESTED", port, portClient, stacks[portClient].stack[stacks[portClient].ptr].data[:])
			ports[port].Reply(stacks[portClient].stack[stacks[portClient].ptr].data[:])

		case machine.I2CFinish:
			println("I2C FINISH")

		default:
			println("default")
		}
	}
}
