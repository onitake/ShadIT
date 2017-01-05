package main

import (
	"log"
)

func GpioInit(line int, output bool) {
	log.Printf("Enabling GPIO line %d as %s", line, map[bool]string{true:"output",false:"input"}[output])
	//Write the pin number to /sys/class/gpio/export
	//Write "in" or "out" to /sys/class/gpio/gpio??/direction
}

func GpioSet(line int, value bool) {
	log.Printf("Setting GPIO line %d to %s", line, map[bool]int{true:1,false:0}[value])
	//Write "1" or "0" to /sys/class/gpio/gpio??/value
}
