package main

import (
	"fmt"
	"time"

	"github.com/kbinani/screenshot"
	"tinygo.org/x/bluetooth"
)

// Can be updated, this is set to around my monitor which is 240hz.
const UPDATE_INTERVAL = 15 * time.Millisecond

// Should be different for everybody, configure this accordingly.
const DEVICE_NAME = "KS03~8a0035"

const FADE_STEPS = 2

func main() {
	adapter := bluetooth.DefaultAdapter
	err := adapter.Enable()
	if err != nil {
		fmt.Println("Failed to enable the adapter:", err)
		return
	}

	var addr bluetooth.Address
	adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		// Specific to me, could me different for other users.
		if device.LocalName() == DEVICE_NAME {
			addr = device.Address
			adapter.StopScan()
		}
	})

	// Delay for .Scan()
	time.Sleep(1 * time.Second)

	if addr == (bluetooth.Address{}) {
		fmt.Println("Device not found")
		return
	}

	device, err := adapter.Connect(addr, bluetooth.ConnectionParams{})
	if err != nil {
		fmt.Println("Failed to connect to device:", err)
		return
	}
	defer device.Disconnect()

	services, err := device.DiscoverServices(nil)
	if err != nil {
		fmt.Println("Failed to discover services:", err)
		return
	}

	var color *bluetooth.DeviceCharacteristic
	for _, service := range services {
		chars, err := service.DiscoverCharacteristics(nil)
		if err != nil {
			fmt.Println("Failed to discover characteristics:", err)
			return
		}

		for _, char := range chars {
			if char.UUID().String() == CHARACTERISTIC_WRITE_UUID {
				color = &char
				break
			}
		}
	}

	if color == nil {
		fmt.Println("Failed to discover characteristics:", err)
		return
	}

	var lastColor []int
	var fading bool

	for {
		if fading {
			continue
		}
		bounds := screenshot.GetDisplayBounds(0)
		// TODO: Make this configurable (i.e multiple lights, top bottom resolution)
		img, _ := screenshot.CaptureRect(bounds)
		cc := (averageImageColor(img))

		if len(lastColor) == 0 {
			lastColor = cc
		}

		if lastColor[0] == cc[0] && lastColor[1] == cc[1] && lastColor[2] == cc[2] {
			color.WriteWithoutResponse([]byte(rgb(cc[0], cc[1], cc[2], 0, 50)))
			continue
		}

		for i := 0; i <= FADE_STEPS; i++ {
			intermediateColor := []int{
				lastColor[0] + (cc[0]-lastColor[0])*i/FADE_STEPS,
				lastColor[1] + (cc[1]-lastColor[1])*i/FADE_STEPS,
				lastColor[2] + (cc[2]-lastColor[2])*i/FADE_STEPS,
			}

			color.WriteWithoutResponse([]byte(rgb(intermediateColor[0], intermediateColor[1], intermediateColor[2], 0, 50)))

			time.Sleep(UPDATE_INTERVAL)
		}

		fading = false

		lastColor = cc

		//fmt.Printf("Color updated to RGB(%d, %d, %d)\n", cc[0], cc[1], cc[2])

		time.Sleep(UPDATE_INTERVAL)
	}
}
