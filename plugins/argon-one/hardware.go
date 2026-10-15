package main

// Talking to the hardware through Linux's device files: the I²C bus and the
// GPIO character device (the "v2" interface of linux/gpio.h).

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// fanController is the case's microcontroller on the I²C bus.
type fanController struct{ f *os.File }

// i2cSlave is the I2C_SLAVE ioctl: talk to this address from now on.
const i2cSlave = 0x0703

func openFan() (*fanController, error) {
	f, err := os.OpenFile(i2cDevice, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	if err := ioctl(f.Fd(), i2cSlave, uintptr(fanAddress)); err != nil {
		f.Close()
		return nil, fmt.Errorf("selecting address 0x%x: %w", fanAddress, err)
	}
	return &fanController{f: f}, nil
}

// set writes the fan speed. Newer cases take it in register 0x80; older
// ones take the byte alone.
func (c *fanController) set(percent int) error {
	speed := byte(max(0, min(100, percent)))
	if _, err := c.f.Write([]byte{0x80, speed}); err == nil {
		return nil
	}
	_, err := c.f.Write([]byte{speed})
	return err
}

// GPIO v2 interface (linux/gpio.h).
const (
	gpioLinesMax     = 64
	gpioMaxNameSize  = 32
	gpioNumAttrsMax  = 10
	flagInput        = 1 << 2
	flagEdgeRising   = 1 << 4
	flagEdgeFalling  = 1 << 5
	flagBiasPullDown = 1 << 9
	eventRisingEdge  = 1
	eventFallingEdge = 2
)

type lineAttribute struct {
	ID      uint32
	Padding uint32
	Value   uint64
}

type lineConfigAttribute struct {
	Attr lineAttribute
	Mask uint64
}

type lineConfig struct {
	Flags    uint64
	NumAttrs uint32
	Padding  [5]uint32
	Attrs    [gpioNumAttrsMax]lineConfigAttribute
}

type lineRequest struct {
	Offsets         [gpioLinesMax]uint32
	Consumer        [gpioMaxNameSize]byte
	Config          lineConfig
	NumLines        uint32
	EventBufferSize uint32
	Padding         [5]uint32
	FD              int32
}

type lineEvent struct {
	TimestampNs uint64
	ID          uint32
	Offset      uint32
	Seqno       uint32
	LineSeqno   uint32
	Padding     [6]uint32
}

// getLineIoctl is GPIO_V2_GET_LINE_IOCTL: _IOWR(0xB4, 0x07, struct
// gpio_v2_line_request).
var getLineIoctl = uintptr(3<<30 | uint32(unsafe.Sizeof(lineRequest{}))<<16 | 0xB4<<8 | 0x07)

// watchLine watches one GPIO line for pulses (rising, then falling) and
// calls onPulse with each pulse's length. It returns only on an error.
func watchLine(chip string, line uint32, onPulse func(time.Duration)) error {
	f, err := os.OpenFile(chip, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	var req lineRequest
	req.Offsets[0] = line
	copy(req.Consumer[:], "rasbeacon-argon-one")
	req.Config.Flags = flagInput | flagEdgeRising | flagEdgeFalling | flagBiasPullDown
	req.NumLines = 1
	if err := ioctl(f.Fd(), getLineIoctl, uintptr(unsafe.Pointer(&req))); err != nil {
		return fmt.Errorf("requesting GPIO %d: %w", line, err)
	}
	events := os.NewFile(uintptr(req.FD), "gpio-line")
	defer events.Close()

	var risen uint64
	buf := make([]byte, unsafe.Sizeof(lineEvent{}))
	for {
		if _, err := events.Read(buf); err != nil {
			return err
		}
		var e lineEvent
		if err := binary.Read(bytes.NewReader(buf), binary.NativeEndian, &e); err != nil {
			return err
		}
		switch e.ID {
		case eventRisingEdge:
			risen = e.TimestampNs
		case eventFallingEdge:
			if risen != 0 && e.TimestampNs > risen {
				onPulse(time.Duration(e.TimestampNs - risen))
			}
			risen = 0
		}
	}
}

func ioctl(fd, request, arg uintptr) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, arg); errno != 0 {
		return errno
	}
	return nil
}

// askRasBeacon asks RasBeacon, through the socket it gives this plugin, to
// shut down or restart.
func askRasBeacon(action string) {
	socket := os.Getenv("RASBEACON_SOCKET")
	if socket == "" {
		fmt.Fprintln(os.Stderr, "no RASBEACON_SOCKET: the power-button permission is missing")
		return
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socket)
		}},
	}
	resp, err := client.Post("http://rasbeacon/power", "application/json", strings.NewReader(`{"action":"`+action+`"}`))
	if err != nil {
		fmt.Fprintln(os.Stderr, "asking RasBeacon:", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		fmt.Fprintln(os.Stderr, "RasBeacon answered", resp.Status)
	}
}
