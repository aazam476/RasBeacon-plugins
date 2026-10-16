// Command argon-one runs the Argon ONE V2 case for RasBeacon: it sets the
// case's fan speed from the processor's temperature, and turns presses of
// the case's power button into a restart or shutdown of RasBeacon.
//
// The case's microcontroller sits on the I²C bus at address 0x1a; a byte
// written to it is the fan speed in percent. When the power button is
// pressed, the case sends a pulse on GPIO 4: about 20-30 ms for a double
// press (restart), about 40-50 ms for a long press (shut down).
//
// It runs as a RasBeacon plugin with the i2c, gpio, power-button and
// temperature permissions (see manifest.json), and answers
// GET /status on port 8080 with the fan speed and temperature, for the
// numbers on its page in RasBeacon.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Where things are on a Raspberry Pi 4.
const (
	i2cDevice   = "/dev/i2c-1"
	fanAddress  = 0x1a
	gpioChip    = "/dev/gpiochip0"
	buttonLine  = 4
	thermalFile = "/sys/class/thermal/thermal_zone0/temp"
	readEvery   = 5 * time.Second
	// hysteresis keeps the fan from switching back and forth at a step:
	// it slows down only once the temperature is this far below it.
	hysteresis = 3.0
)

// Step is one point of the fan curve: from TempC up, the fan runs at Speed.
type Step struct {
	TempC float64
	Speed int
}

// Curve is the fan curve, coolest step first.
type Curve []Step

// SpeedFor returns the fan speed for a temperature. current is the speed
// now: slowing down waits until the temperature is hysteresis below the
// step that set it.
func (c Curve) SpeedFor(tempC float64, current int) int {
	speed := 0
	for _, step := range c {
		if tempC >= step.TempC {
			speed = step.Speed
		}
	}
	if speed < current {
		// Keep the current speed while still near the step that set it.
		for _, step := range c {
			if step.Speed == current && tempC > step.TempC-hysteresis {
				return current
			}
		}
	}
	return speed
}

// curveFromEnv reads the curve from the plugin's settings.
func curveFromEnv(getenv func(string) string) (Curve, error) {
	number := func(name string, fallback float64) (float64, error) {
		text := strings.TrimSpace(getenv(name))
		if text == "" {
			return fallback, nil
		}
		n, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return 0, fmt.Errorf("%s=%q is not a number", name, text)
		}
		return n, nil
	}
	var c Curve
	defaults := []Step{{55, 10}, {60, 55}, {65, 100}}
	for i, d := range defaults {
		t, err := number(fmt.Sprintf("FAN_TEMP%d", i+1), d.TempC)
		if err != nil {
			return nil, err
		}
		s, err := number(fmt.Sprintf("FAN_SPEED%d", i+1), float64(d.Speed))
		if err != nil {
			return nil, err
		}
		if s < 0 || s > 100 {
			return nil, fmt.Errorf("FAN_SPEED%d must be from 0 to 100", i+1)
		}
		c = append(c, Step{TempC: t, Speed: int(s)})
	}
	for i := 1; i < len(c); i++ {
		if c[i].TempC <= c[i-1].TempC {
			return nil, errors.New("each fan step's temperature must be higher than the one before")
		}
	}
	return c, nil
}

// Press is what a power button pulse means.
type Press int

const (
	NoPress Press = iota
	Restart
	Shutdown
)

// Classify turns the length of a pulse from the case into a press.
func Classify(width time.Duration) Press {
	switch {
	case width >= 15*time.Millisecond && width < 35*time.Millisecond:
		return Restart
	case width >= 35*time.Millisecond && width < 70*time.Millisecond:
		return Shutdown
	}
	return NoPress
}

// status is the latest reading, for GET /status. The fan goroutine writes
// it and the web server reads it, so a mutex makes them take turns.
type status struct {
	mu          sync.Mutex
	Fan         int // percent; -1 until the fan is first set
	Temperature float64
}

func main() {
	log.SetFlags(0)
	curve, err := curveFromEnv(os.Getenv)
	if err != nil {
		log.Fatalf("settings: %v", err)
	}
	log.Printf("fan curve: %v", curve)

	st := &status{Fan: -1}
	go serveStatus(st)
	go watchButton()
	runFan(curve, st)
}

// runFan reads the temperature every few seconds and sets the fan.
func runFan(curve Curve, st *status) {
	fan, err := openFan()
	if err != nil {
		log.Fatalf("the case's fan controller can't be reached: %v", err)
	}
	current := -1
	for {
		temp, err := readTemperature()
		if err != nil {
			log.Printf("reading the temperature: %v", err)
		} else {
			speed := curve.SpeedFor(temp, max(current, 0))
			if speed != current {
				if err := fan.set(speed); err != nil {
					log.Printf("setting the fan to %d%%: %v", speed, err)
				} else {
					log.Printf("%.1f °C: fan %d%%", temp, speed)
					current = speed
				}
			}
			st.mu.Lock()
			st.Fan, st.Temperature = current, temp
			st.mu.Unlock()
		}
		time.Sleep(readEvery)
	}
}

// readTemperature reads the processor's temperature in °C (the kernel
// gives it in thousandths of a degree).
func readTemperature() (float64, error) {
	data, err := os.ReadFile(thermalFile)
	if err != nil {
		return 0, err
	}
	milli, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	return float64(milli) / 1000, nil
}

// serveStatus answers GET /status with the latest reading, for the
// numbers on the plugin's page in RasBeacon.
func serveStatus(st *status) {
	http.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		st.mu.Lock()
		defer st.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Fan         int     `json:"fan"`
			Temperature float64 `json:"temperature"`
		}{st.Fan, st.Temperature})
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// watchButton turns power button pulses into requests to RasBeacon.
func watchButton() {
	for {
		err := watchLine(gpioChip, buttonLine, func(width time.Duration) {
			switch Classify(width) {
			case Restart:
				log.Printf("power button pressed twice (%v): restarting", width)
				askRasBeacon("reboot")
			case Shutdown:
				log.Printf("power button held (%v): shutting down", width)
				askRasBeacon("shutdown")
			}
		})
		log.Printf("watching the power button: %v; trying again in a minute", err)
		time.Sleep(time.Minute)
	}
}
