package main

import (
	"image"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/vector"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	vec := vector.NewRasterizer(25, 25)

	// create a rectangle
	vec.LineTo(25, 0)  // top right
	vec.LineTo(25, 25) // bottom right
	vec.LineTo(0, 25)  // bottom left
	vec.LineTo(0, 0)   // top left

	// fill the img's background for visibility
	draw.Draw(img, img.Rect, image.White, image.Point{}, draw.Over)

	// Test cases:
	// Replace (0,0) to one of these:
	// 1. (-10, -10) Panics
	// 2. (35, 35) Panics
	// 3. (-10, 15) Draws those -10x pixels on the right and a pixel higher
	// 4. (35, 15) Draws those +10x pixels on the left and a pixel lower
	addPoint := image.Pt(35, 35)

	// Draw an image.Uniform on an image.RGBA
	vec.Draw(img, vec.Bounds().Add(addPoint), image.Black, image.Point{})

	// Save the image. I avoided the error handling for now.
	f, _ := os.OpenFile("go-issue/test-case.png", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	png.Encode(f, img)
}
