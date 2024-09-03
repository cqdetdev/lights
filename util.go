package main

import (
	"encoding/hex"
	"fmt"
	"image"
	"math"
)

func toHex(value int) string {
	return fmt.Sprintf("%02X", value)
}

func rgb(red int, green int, blue int, white int, brightness int) []byte {
	prefix := "5A00"
	isRGB := "01"
	rgbHex := toHex(red) + toHex(green) + toHex(blue)
	whiteHex := toHex(white)
	brightnessHex := toHex(brightness)
	speed := "00"
	suffix := "A5"

	bytes, err := hex.DecodeString(prefix + isRGB + rgbHex + whiteHex + brightnessHex + speed + suffix)
	if err != nil {
		fmt.Println("Failed to decode hex string:", err)
	}

	return bytes
}

func averageImageColor(img image.Image) []int {
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

// AI generated function, need to find a better way to correct for oranges/dark-browns
func adjustColorForOrange(rgb []int) []int {
	red := rgb[0]
	green := rgb[1]
	blue := rgb[2]

	if red < 200 && green < 200 && blue < 200 {
		return []int{red, green, blue}
	}

	const yellowThreshold = 200
	const brownThreshold = 100
	const minOrangeRed = 255
	const minOrangeGreen = 100
	const maxBlue = 50

	average := (red + green + blue) / 3

	if average < brownThreshold {
		red = minOrangeRed
		green = int(math.Max(float64(green), minOrangeGreen))
		blue = int(math.Min(float64(blue), maxBlue))
	} else if average > yellowThreshold {
		red = minOrangeRed
		green = int(math.Max(float64(green), minOrangeGreen))
		blue = 0
	}

	return []int{red, green, blue}
}
