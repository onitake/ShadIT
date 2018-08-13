package main

import (
	"log"
	"time"
)

type Shutter struct {
	Name string
	GpioUp Gpio
	GpioDown Gpio
	Position float32
	Angle float32
	UpTime time.Duration
	DownTime time.Duration
	FlipUpTime time.Duration
	FlipDownTime time.Duration
}

func (shutter *Shutter) Init() {
	log.Printf("Initializing GPIO lines of shutter %s\n", shutter.Name)
	shutter.GpioUp.Init()
	shutter.GpioUp.Set(false)
	shutter.GpioDown.Init()
	shutter.GpioDown.Set(false)
}

func (shutter *Shutter) Reset() {
	log.Printf("Moving shutter %s to position 0\n", shutter.Name)
	shutter.GpioDown.Set(false)
	shutter.GpioUp.Set(true)
	time.Sleep(shutter.UpTime)
	shutter.GpioUp.Set(false)
	shutter.Position = 0.0
	shutter.Angle = 0.0
}

func (shutter *Shutter) Flip(angle float32) {
	log.Printf("Flipping shutter %s to angle %f\n", shutter.Name, angle)
	// TODO adjust the position as well, or use angle directly
	shutter.GpioDown.Set(false)
	shutter.GpioUp.Set(true)
	time.Sleep(shutter.FlipUpTime)
	shutter.GpioUp.Set(false)
	shutter.GpioDown.Set(true)
	time.Sleep(time.Duration(float32(shutter.FlipDownTime) * angle))
	shutter.GpioDown.Set(false)
	shutter.Angle = angle
}

func (shutter *Shutter) Move(position float32) {
	// TODO adjust the angle as well, according to the direction
	if (position > shutter.Position) {
		log.Printf("Moving shutter %s down to position %f\n", shutter.Name, position)
		shutter.GpioUp.Set(false)
		shutter.GpioDown.Set(true)
		time.Sleep(time.Duration(float32(shutter.DownTime) * (position - shutter.Position)))
		shutter.GpioDown.Set(false)
		shutter.Position = position
		shutter.Angle = 0.0
	} else if (position < shutter.Position) {
		log.Printf("Moving shutter %s up to position %f\n", shutter.Name, position)
		shutter.GpioDown.Set(false)
		shutter.GpioUp.Set(true)
		time.Sleep(time.Duration(float32(shutter.UpTime) * (shutter.Position - position)))
		shutter.GpioUp.Set(false)
		shutter.Position = position
		shutter.Angle = 0.0
	}
	log.Printf("Not moving shutter %s\n", shutter.Name)
}

type ShutterState struct {
	Shutters map[string]*Shutter
}

func NewShutterState(config *Configuration) (*ShutterState, error) {
	state := &ShutterState{
		Shutters: make(map[string]*Shutter),
	}
	for _, shutter := range config.Shutters {
		gpioup, err := NewGpio(shutter.GpioUp, true)
		if err != nil {
			return nil, err
		}
		gpiodown, err := NewGpio(shutter.GpioDown, true)
		if err != nil {
			return nil, err
		}
		state.Shutters[shutter.Name] = &Shutter{
			Name: shutter.Name,
			GpioUp: gpioup,
			GpioDown: gpiodown,
			Position: 0.0,
			Angle: 0.0,
			DownTime: time.Duration(config.DownTime) * time.Second,
			UpTime: time.Duration(config.UpTime) * time.Second,
			FlipUpTime: time.Duration(config.FlipTime) * time.Second,
			FlipDownTime: time.Duration(config.FlipTime) * time.Second, 
		}
	}
	return state, nil
}

