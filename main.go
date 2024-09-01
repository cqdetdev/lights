package main

import (
	"encoding/hex"
	"fmt"
	"image"
	"math"
	"time"

	"github.com/kbinani/screenshot"
	"tinygo.org/x/bluetooth"
)

func main() {
	adapter := bluetooth.DefaultAdapter
	err := adapter.Enable()
	if err != nil {
		fmt.Println("Failed to enable the adapter:", err)
		return
	}

	var deviceAddress bluetooth.Address
	adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if device.LocalName() == "KS03~8a0035" {
			deviceAddress = device.Address
			adapter.StopScan()
		}
	})

	// Wait for the scan to complete.
	time.Sleep(1 * time.Second)

	if deviceAddress == (bluetooth.Address{}) {
		fmt.Println("Device not found")
		return
	}

	device, err := adapter.Connect(deviceAddress, bluetooth.ConnectionParams{})
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
		fmt.Printf("Service: %s\n", service.UUID().String())

		// Discover characteristics within the service.
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

	for {
		bounds := screenshot.GetDisplayBounds(0)
		img, _ := screenshot.CaptureRect(bounds)
		c := averageColor(img)
		cc := adjustColorForOrange(c[0], c[1], c[2])

		go color.WriteWithoutResponse([]byte(rgbNew(cc[0], cc[1], cc[2], 0, 50)))
		fmt.Printf("Color updated to RGB(%d, %d, %d)\n", cc[0], cc[1], cc[2])

		// Wait for half a second before updating the color again
		time.Sleep(150 * time.Millisecond)
	}
	if err != nil {
		fmt.Println("Failed to write to characteristic:", err)
	} else {
		fmt.Printf("Color changed successfully! (%d)\n")
	}
}

func hexByte(value int) string {
	return fmt.Sprintf("%02X", value)
}

func rgbNew(red int, green int, blue int, white int, brightness int) []byte {
	prefix := "5A00"
	isRGB := "01"
	rgbHex := hexByte(red) + hexByte(green) + hexByte(blue)
	whiteHex := hexByte(white)
	brightnessHex := hexByte(brightness)
	speed := "00"
	suffix := "A5"

	bytes, err := hex.DecodeString(prefix + isRGB + rgbHex + whiteHex + brightnessHex + speed + suffix)
	if err != nil {
		fmt.Println("Failed to decode hex string:", err)
	}

	return bytes
}

func averageColor(img image.Image) []int {
	bounds := img.Bounds()
	totalR, totalG, totalB, count := 0, 0, 0, 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			totalR += int(r >> 8)
			totalG += int(g >> 8)
			totalB += int(b >> 8)
			count++
		}
	}

	avgR := totalR / count
	avgG := totalG / count
	avgB := totalB / count

	return []int{avgR, avgG, avgB}
}

// Adjust the RGB values to ensure the color stays in the orange spectrum
func adjustColorForOrange(red, green, blue int) []int {
	// Define thresholds for approaching yellow and brown

	if red < 200 && green < 200 && blue < 200 {
		return []int{red, green, blue}
	}

	const yellowThreshold = 200
	const brownThreshold = 100
	const minOrangeRed = 255
	const minOrangeGreen = 100
	const maxBlue = 50

	// Calculate the average of the RGB values
	average := (red + green + blue) / 3

	if average < brownThreshold {
		// If close to brown, ensure color is a darker orange
		red = minOrangeRed
		green = int(math.Max(float64(green), minOrangeGreen))
		blue = int(math.Min(float64(blue), maxBlue))
	} else if average > yellowThreshold {
		// If approaching yellow, adjust to prevent yellow hues
		red = minOrangeRed
		green = int(math.Max(float64(green), minOrangeGreen))
		blue = 0
	}

	return []int{red, green, blue}
}
