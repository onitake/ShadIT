// +build linux

/* Copyright (c) 2017-2018, Gregor Riepl <onitake@gmail.com>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 *     Redistributions of source code must retain the above copyright notice,
 *     this list of conditions and the following disclaimer.
 *
 *     Redistributions in binary form must reproduce the above copyright notice,
 *     this list of conditions and the following disclaimer in the documentation
 *     and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
 * ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
 * ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"log"
	"os"
	"strconv"
	"errors"
)

// linuxGpio is the Linux-specific implementation of the GPIO interface.
type linuxGpio struct {
	// Line is the GPIO line number.
	// Linux uses a flat GPIO interface model with continuous numbering.
	// Which GPIO is accessed, depends on the hardware configuration, usually
	// from a system-specific configuration block such as ACPI or DeviceTree.
	Line int
	// Output is the type of interface. If true, the line is configured as output.
	Output bool
}

// NewGpio creates a GPIO handler for a single GPIO line.
// The implementation and the line ID format is platform specific.
// For example, on Linux, the line is interpreted as an unsigned integer and
// refers to a device in /sys/class/gpio/
func NewGpio(spec string, output bool) (Gpio, error) {
	line, err := strconv.Atoi(spec)
	if err != nil {
		return nil, err
	}
	if line < 0 {
		return nil, errors.New("GPIO line cannot be negative")
	}
	return &linuxGpio{
		Line: line,
		Output: output,
	}, nil
}

func (g *linuxGpio) Init() error {
	log.Printf("Enabling GPIO line %d as %s", g.Line, map[bool]string{true:"output",false:"input"}[g.Output])
	
	// Write the pin number to /sys/class/gpio/export
	export, err := os.Create("/sys/class/gpio/export")
	if err != nil {
		log.Print(err)
		return err
	}
	export.Write([]byte(strconv.Itoa(g.Line)))
	export.Close()
	
	// Write "in" or "out" to /sys/class/gpio/gpio??/direction
	out, err := os.Create("/sys/class/gpio/gpio" + strconv.Itoa(g.Line) + "/direction")
	if err != nil {
		return err
	}
	defer out.Close()

	if g.Output {
		out.Write([]byte("out"))
	} else {
		out.Write([]byte("in"))
	}
	
	return nil
}

func (g *linuxGpio) Set(value bool) error {
	log.Printf("Setting GPIO line %d to %s", g.Line, map[bool]int{true:1,false:0}[value])

	// Write "1" or "0" to /sys/class/gpio/gpio??/value
	gpio, err := os.Create("/sys/class/gpio/gpio" + strconv.Itoa(g.Line) + "/value")
	if err != nil {
		return err
	}
	defer gpio.Close()

	if value {
		gpio.Write([]byte("1"))
	} else {
		gpio.Write([]byte("0"))
	}
	
	return nil
}

func (g *linuxGpio) Get() (bool, error) {
	log.Printf("Getting value of GPIO line %d", g.Line)

	// Read from /sys/class/gpio/gpio??/value
	gpio, err := os.Open("/sys/class/gpio/gpio" + strconv.Itoa(g.Line) + "/value")
	if err != nil {
		return false, err
	}
	defer gpio.Close()

	value := make([]byte, 1)
	n, err := gpio.Read(value)
	if n > 0 {
		switch value[0] {
			case '0':
				log.Printf("Line is low")
				return false, nil
			case '1':
				log.Printf("Line is high")
				return true, nil
			default:
				return false, errors.New("Invalid state: " + string(value[0]))
		}
	}

	if err == nil {
		return false, errors.New("Can't read from GPIO line: unknown error")
	} else {
		return false, err
	}
}
