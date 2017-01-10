package main

import (
	"log"
	"time"
)

type Shutter struct {
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

func (shutter *Shutter) Initialize() {
	log.Printf("Initializing GPIO lines of shutter %s\n", shutter.Name)
	GpioInit(shutter.GpioUp, true)
	GpioSet(shutter.GpioUp, false)
	GpioInit(shutter.GpioDown, true)
	GpioSet(shutter.GpioDown, false)
}

func (shutter *Shutter) Reset() {
	log.Printf("Moving shutter %s to position 0\n", shutter.Name)
	GpioSet(shutter.GpioDown, false)
	GpioSet(shutter.GpioUp, true)
	time.Sleep(shutter.UpTime)
	GpioSet(shutter.GpioUp, false)
	shutter.Position = 0.0
	shutter.Angle = 0.0
}

func (shutter *Shutter) Flip(angle float32) {
	log.Printf("Flipping shutter %s to angle %f\n", shutter.Name, angle)
	GpioSet(shutter.GpioDown, false)
	GpioSet(shutter.GpioUp, true)
	time.Sleep(shutter.FlipUpTime)
	GpioSet(shutter.GpioUp, false)
	GpioSet(shutter.GpioDown, true)
	time.Sleep(time.Duration(float32(shutter.FlipDownTime) * angle))
	GpioSet(shutter.GpioDown, false)
	shutter.Angle = angle
}

func (shutter *Shutter) Move(position float32) {
	if (position > shutter.Position) {
		log.Printf("Moving shutter %s down to position %f\n", shutter.Name, position)
		GpioSet(shutter.GpioUp, false)
		GpioSet(shutter.GpioDown, true)
		time.Sleep(time.Duration(float32(shutter.DownTime) * (position - shutter.Position)))
		GpioSet(shutter.GpioDown, false)
		shutter.Position = position
		shutter.Angle = 0.0
	} else if (position < shutter.Position) {
		log.Printf("Moving shutter %s up to position %f\n", shutter.Name, position)
		GpioSet(shutter.GpioDown, false)
		GpioSet(shutter.GpioUp, true)
		time.Sleep(time.Duration(float32(shutter.UpTime) * (shutter.Position - position)))
		GpioSet(shutter.GpioUp, false)
		shutter.Position = position
		shutter.Angle = 0.0
	}
	log.Printf("Not moving shutter %s\n", shutter.Name)
}

type ShutterState struct {
	Shutters map[string]*Shutter
}

func NewShutterState(config *Configuration) *ShutterState {
	state := &ShutterState{
		Shutters: make(map[string]*Shutter),
	}
	for _, shutter := range config.Shutters {
		state.Shutters[shutter.Name] = &Shutter{
			Name: shutter.Name,
			GpioUp: shutter.GpioUp,
			GpioDown: shutter.GpioDown,
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

