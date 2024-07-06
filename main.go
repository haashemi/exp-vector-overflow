package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/vector"
)

// ImageWrapper is a struct to embed 'image.RGBA' and 'image.Alpha' so vector's
// Draw method can't find out what type of image it actually is.
type ImageWrapper struct{ draw.Image }

func main() {
	// A basic color as an example, doesn't really matter.
	uniformImage := image.NewUniform(color.RGBA{127, 127, 127, 255})
	// Size of the images we're going to experiment
	imageRect := image.Rect(0, 0, 50, 50)
	// Position of the vector over the image.
	// Try these values:
	// 	1. Basic: (0, 0) Should work fine
	//  2. Negative: (-10, -10) Should panic on RGBA and Alpha
	//  2. Negative Y: (0, -1) Should panic on RGBA and Alpha
	//  3. Negative X: (-10, 2) Should *also* draw on right-side of the RGBA and Alpha images
	//  4. Overflow: (35, 35) Should panic on RGBA and Alpha
	//  4. Overflow Y: (0, 30) Should panic on RGBA and Alpha
	//  5. Overflow X: (35, 0) Should *also* draw on left-side of the RGBA and Alpha images
	//
	// Feel free to try out more values.
	addPoint := image.Pt(-10, 2)

	// draw a basic rectangle just fow visualizing the issue.
	createVector := func() *vector.Rasterizer {
		vec := vector.NewRasterizer(25, 25)

		vec.LineTo(float32(imageRect.Dx()), 0)
		vec.LineTo(float32(imageRect.Dx()), float32(imageRect.Dy()))
		vec.LineTo(0, float32(imageRect.Dy()))
		vec.LineTo(0, 0)

		return vec
	}

	// The goal of the issue is image.RGBA and image.Alpha, feel free to comment/uncomment any of these.
	images := map[string]draw.Image{
		// "Alpha": image.NewAlpha(imageRect), // PANICS!
		"RGBA": image.NewRGBA(imageRect), // PANICS!
		// "WrappedAlpha": ImageWrapper{image.NewAlpha(imageRect)}, // Works totally fine. But it's a hack.
		// "WrappedRGBA":  ImageWrapper{image.NewRGBA(imageRect)},  // Works totally fine. But it's a hack.
		// "Alpha16":      image.NewAlpha16(imageRect),             // Works totally fine.
		// "CMYK":         image.NewCMYK(imageRect),                // Works totally fine.
		// "Gray":         image.NewGray(imageRect),                // Works totally fine.
		// "Gray16":       image.NewGray16(imageRect),              // Works totally fine.
		// "NRGBA":        image.NewNRGBA(imageRect),               // Works totally fine.
		// "NRGBA64":      image.NewNRGBA64(imageRect),             // Works totally fine.
		// "RGBA64":       image.NewRGBA64(imageRect),              // Works totally fine.
	}

	for imgType, img := range images {
		// Just draw a background for experimental purposes.
		// draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Over)

		vec := createVector()

		// If 'r' be greater than the vector's bounds, it'll panic.
		// Thus we'll use its bounds for now to avoid unnecessary crashes.
		vec.Draw(img, vec.Bounds().Add(addPoint), uniformImage, image.Point{})

		f, err := os.OpenFile(fmt.Sprintf("out-%s.png", imgType), os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			fmt.Printf("[ERR] [IMG: %s] Failed to open a file > %v\n", imgType, err)
			continue
		}

		if err = png.Encode(f, img); err != nil {
			fmt.Printf("[ERR] [IMG: %s] Failed to encode the image > %v\n", imgType, err)
			continue
		}

		fmt.Printf("[INF] [IMG: %s] Drawn and saved successfully\n", imgType)
	}
}
