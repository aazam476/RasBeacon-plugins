package main

import (
	"testing"
	"time"
	"unsafe"
)

func TestCurve(t *testing.T) {
	c := Curve{{55, 10}, {60, 55}, {65, 100}}
	cases := []struct {
		temp    float64
		current int
		want    int
	}{
		{40, 0, 0}, {55, 0, 10}, {61, 10, 55}, {70, 55, 100},
		// Slowing down waits for the temperature to drop 3 °C below the step.
		{63, 100, 100}, {61.9, 100, 55}, {57.5, 55, 55}, {56.9, 55, 10}, {50, 10, 0},
	}
	for _, tc := range cases {
		if got := c.SpeedFor(tc.temp, tc.current); got != tc.want {
			t.Errorf("SpeedFor(%v, %d) = %d, want %d", tc.temp, tc.current, got, tc.want)
		}
	}
}

func TestCurveFromEnv(t *testing.T) {
	env := map[string]string{"FAN_TEMP1": "50", "FAN_SPEED3": "90"}
	c, err := curveFromEnv(func(k string) string { return env[k] })
	if err != nil || c[0].TempC != 50 || c[2].Speed != 90 || c[1].Speed != 55 {
		t.Errorf("curve = %v, %v", c, err)
	}
	env = map[string]string{"FAN_TEMP2": "50"} // below step 1's 55
	if _, err := curveFromEnv(func(k string) string { return env[k] }); err == nil {
		t.Error("steps out of order must be refused")
	}
}

func TestClassify(t *testing.T) {
	for width, want := range map[time.Duration]Press{
		5 * time.Millisecond: NoPress, 20 * time.Millisecond: Restart, 30 * time.Millisecond: Restart,
		40 * time.Millisecond: Shutdown, 50 * time.Millisecond: Shutdown, 200 * time.Millisecond: NoPress,
	} {
		if got := Classify(width); got != want {
			t.Errorf("Classify(%v) = %v, want %v", width, got, want)
		}
	}
}

// The structures must match linux/gpio.h exactly, or the kernel refuses them.
func TestGPIOStructSizes(t *testing.T) {
	if s := unsafe.Sizeof(lineRequest{}); s != 592 {
		t.Errorf("gpio_v2_line_request is %d bytes, want 592", s)
	}
	if s := unsafe.Sizeof(lineEvent{}); s != 48 {
		t.Errorf("gpio_v2_line_event is %d bytes, want 48", s)
	}
	if getLineIoctl != 0xC250B407 {
		t.Errorf("GPIO_V2_GET_LINE_IOCTL = %#x", getLineIoctl)
	}
}
