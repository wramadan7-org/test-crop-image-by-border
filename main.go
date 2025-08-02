package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

type ValidType int

const (
	Row ValidType = iota
	Column
)

func isBlack(c color.Color) bool {
	r, g, b, a := c.RGBA()
	if a == 0 {
		return false
	}
	// A pixel is considered black if `R=0`, `G=0`, and `B=0`
	return r == 0 && g == 0 && b == 0
}

func isValidBorder(validType ValidType, img image.Image, coordinate, initialStartCoordinate, initialEndCoordinate int) bool {
	switch validType {
	case Row:
		for x := initialStartCoordinate; x <= initialEndCoordinate; x++ {
			if !isBlack(img.At(x, coordinate)) {
				return false
			}
		}
		return true
	case Column:
		for y := initialStartCoordinate; y <= initialEndCoordinate; y++ {
			if !isBlack(img.At(coordinate, y)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func main() {
	logFile, err := os.OpenFile("crop-image.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Cannot open log file:", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	log.Default().Println("Starting image cropping...")
	file, err := os.Open("./image.png")
	if err != nil {
		log.Fatal("Error opening image file:", err)
	}
	defer file.Close()

	log.Default().Println("Decoding image...")
	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal("Error decoding image file:", err)
	}

	log.Default().Println("Define full area coordinate of image...")
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	log.Default().Println("Scanning image to find black pixel coordinates for cropping...")
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			crop := img.At(x, y)

			if isBlack(crop) {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	// Check if all x filled black color in the first y (top row)
	for y := minY; y <= maxY; y++ {
		if isValidBorder(Row, img, y, minX, maxX) {
			minY = y
			break
		}
	}

	// Check if all x filled black color in the last y (bottom row)
	for y := maxY; y >= minY; y-- {
		if isValidBorder(Row, img, y, minX, maxX) {
			maxY = y
			break
		}
	}

	// Check if all y filled black color in the first x (left column)
	for x := minX; x <= maxX; x++ {
		if isValidBorder(Column, img, x, minY, maxY) {
			minX = x
			break
		}
	}

	// Check if all y filled black color in the last x (right column)
	for x := maxX; x >= minX; x-- {
		if isValidBorder(Column, img, x, minY, maxY) {
			maxX = x
			break
		}
	}

	log.Default().Printf(
		"Create cropping image area. Top-left=(%d,%d), Bottom-right=(%d,%d)",
		minX, minY, maxX+1, maxY+1,
	)
	cropRect := image.Rect(minX, minY, maxX+1, maxY+1)

	log.Default().Printf(
		"Cropping area defined. Creating SubImage from rectangle: top-left=(%d,%d), bottom-right=(%d,%d)",
		minX, minY, maxX+1, maxY+1,
	)
	subImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	log.Default().Println("Creating cropped image file...")
	outFile, err := os.Create("output.png")
	if err != nil {
		log.Fatal("Error creating cropped image file:", err)
	}
	defer outFile.Close()

	log.Default().Println("Encoding cropped image file...")
	if err := png.Encode(outFile, subImg); err != nil {
		log.Fatal("Error encoding cropped image file:", err)
	}

	log.Println("Image cropping completed successfully.")
}
