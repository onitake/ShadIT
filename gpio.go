package main

import (
	"log"
	"os"
	"strconv"
)

func GpioInit(line int, output bool) error {
	log.Printf("Enabling GPIO line %d as %s", line, map[bool]string{true:"output",false:"input"}[output])
	
	// Write the pin number to /sys/class/gpio/export
	export, err := os.Create("/sys/class/gpio/export")
	if err != nil {
		log.Print(err)
		return err
	}
	export.Write([]byte(strconv.Itoa(line)))
	export.Close()
	
	// Write "in" or "out" to /sys/class/gpio/gpio??/direction
	out, err := os.Create("/sys/class/gpio/gpio" + strconv.Itoa(line) + "/direction")
	if err != nil {
		log.Print(err)
		return err
	}
	if output {
		out.Write([]byte("out"))
	} else {
		out.Write([]byte("in"))
	}
	out.Close()
	
	return nil
}

func GpioSet(line int, value bool) error {
	log.Printf("Setting GPIO line %d to %s", line, map[bool]int{true:1,false:0}[value])

	// Write "1" or "0" to /sys/class/gpio/gpio??/value
	gpio, err := os.Create("/sys/class/gpio/gpio" + strconv.Itoa(line) + "/value")
	if err != nil {
		log.Print(err)
		return err
	}
	if value {
		gpio.Write([]byte("1"))
	} else {
		gpio.Write([]byte("0"))
	}
	gpio.Close()
	
	return nil
}
