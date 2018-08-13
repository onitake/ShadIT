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

