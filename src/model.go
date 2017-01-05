package main

import (
	"log"
	"time"
)

type Shade struct {
	Name string
	GpioUp int
	GpioDown int
	Position float32
	Angle float32
	UpTime time.Duration
	DownTime time.Duration
	FlipUpTime time.Duration
	FlipDownTime time.Duration
}

func (shade *Shade) Initialize() {
	log.Printf("Initializing GPIO lines of shade %s\n", shade.Name)
	GpioInit(shade.GpioUp, true)
	GpioSet(shade.GpioUp, false)
	GpioInit(shade.GpioDown, true)
	GpioSet(shade.GpioDown, false)
}

func (shade *Shade) Reset() {
	log.Printf("Moving shade %s to position 0\n", shade.Name)
	GpioSet(shade.GpioDown, false)
	GpioSet(shade.GpioUp, true)
	time.Sleep(shade.UpTime)
	GpioSet(shade.GpioUp, false)
	shade.Position = 0.0
	shade.Angle = 0.0
}

func (shade *Shade) Flip(angle float32) {
	log.Printf("Flipping shade %s to angle %f\n", shade.Name, angle)
	GpioSet(shade.GpioDown, false)
	GpioSet(shade.GpioUp, true)
	time.Sleep(shade.FlipUpTime)
	GpioSet(shade.GpioUp, false)
	GpioSet(shade.GpioDown, true)
	time.Sleep(time.Duration(float32(shade.FlipDownTime) * angle))
	GpioSet(shade.GpioDown, false)
	shade.Angle = angle
}

func (shade *Shade) Move(position float32) {
	if (position > shade.Position) {
		log.Printf("Moving shade %s down to position %f\n", shade.Name, position)
		GpioSet(shade.GpioUp, false)
		GpioSet(shade.GpioDown, true)
		time.Sleep(time.Duration(float32(shade.DownTime) * (position - shade.Position)))
		GpioSet(shade.GpioDown, false)
		shade.Position = position
		shade.Angle = 0.0
	} else if (position < shade.Position) {
		log.Printf("Moving shade %s up to position %f\n", shade.Name, position)
		GpioSet(shade.GpioDown, false)
		GpioSet(shade.GpioUp, true)
		time.Sleep(time.Duration(float32(shade.UpTime) * (shade.Position - position)))
		GpioSet(shade.GpioUp, false)
		shade.Position = position
		shade.Angle = 0.0
	}
	log.Printf("Not moving shade %s\n", shade.Name)
}

type ShadeState struct {
	Shades map[string]*Shade
}

func NewShadeState(config *Configuration) *ShadeState {
	state := &ShadeState{
		Shades: make(map[string]*Shade),
	}
	for _, shade := range config.Shades {
		state.Shades[shade.Name] = &Shade{
			Name: shade.Name,
			GpioUp: shade.GpioUp,
			GpioDown: shade.GpioDown,
			Position: 0.0,
			Angle: 0.0,
			DownTime: time.Duration(config.DownTime) * time.Second,
			UpTime: time.Duration(config.UpTime) * time.Second,
			FlipUpTime: time.Duration(config.FlipTime) * time.Second,
			FlipDownTime: time.Duration(config.FlipTime) * time.Second, 
		}
	}
	return state
}

