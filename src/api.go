package main

import (
	"log"
	"os"
	"time"
	"strings"
	"net/http"
	"encoding/json"
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

type Configuration struct {
	Listen string
	UpTime int
	DownTime int
	FlipTime int
	Shades []struct {
		Name string
		GpioUp int
		GpioDown int
	}
}

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

type ShadeServer struct {
	State *ShadeState
}

func NewShadeServer(state *ShadeState) *ShadeServer {
	return &ShadeServer{
		State: state,
	}
}

func (server *ShadeServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	path := strings.Split(request.URL.Path, "/")
	log.Printf("len(path)=%d path[0]=%s path[1]=%s\n", len(path), path[0], path[1])
	if (len(path) == 0) {
		shades := make(map[string]map[string]interface{})
		for _, shade := range(server.State.Shades) {
			shades[shade.Name] = map[string]interface{}{
				"name": shade.Name,
				"position": shade.Position,
				"angle": shade.Angle,
			}
		}
		writer.Header().Add("Content-Type", "application/json")
		response, error := json.Marshal(map[string]interface{}{
			"shades": shades,
			"error": "none",
		})
		if (error == nil) {
			writer.WriteHeader(http.StatusOK);
			writer.Write(response)
		} else {
			writer.WriteHeader(http.StatusInternalServerError);
			writer.Write([]byte("{'error':'internal'}"))
			log.Print(error)
		}
	} else {
		shade := server.State.Shades[path[0]]
		if (shade != nil) {
			if (len(path) > 1) {
				writer.Header().Add("Content-Type", "application/json")
				writer.WriteHeader(http.StatusNotFound);
				writer.Write([]byte("{'error':'not_implemented'}"))
			} else {
				writer.Header().Add("Content-Type", "application/json")
				response, error := json.Marshal(map[string]interface{}{
					"name": shade.Name,
					"position": shade.Position,
					"angle": shade.Angle,
				})
				if (error == nil) {
					writer.WriteHeader(http.StatusOK);
					writer.Write(response)
				} else {
					writer.WriteHeader(http.StatusInternalServerError);
					writer.Write([]byte("{'error':'internal'}"))
					log.Print(error)
				}
			}
		} else {
			writer.Header().Add("Content-Type", "application/json")
			writer.WriteHeader(http.StatusNotFound);
			writer.Write([]byte("{'error':'invalid_object'}"))
		}
	}
}

func main() {
	var configname string
	if len(os.Args) > 1 {
		configname = os.Args[1]
	} else {
		configname = "server.json"
	}
	
	configfile, err := os.Open(configname)
	if err != nil {
		log.Fatal("Can't read configuration from server.json: ", err)
	}
	decoder := json.NewDecoder(configfile)
	config := Configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error parsing configuration: ", err)
	}
	configfile.Close()

	state := NewShadeState(&config)
	server := NewShadeServer(state)
	log.Fatal(http.ListenAndServe(config.Listen, server))
}
