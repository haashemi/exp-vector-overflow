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
	// Contains three type of positions:
	//  1. Common: Nothing special
	//  2. Negative: Move the vector to top-left / top / left
	//  3. Overflow: Move the vector to bottom-right / bottom / right
	offsets := []struct {
		Type  string
		Point image.Point
	}{
		{"Common", image.Pt(0, 0)},       // Everything's fine
		{"Negative", image.Pt(-10, -10)}, // Panic with image.RGBA and image.Alpha
		{"NegativeY", image.Pt(0, -1)},   // Panic with image.RGBA and image.Alpha
		{"NegativeX", image.Pt(-10, 2)},  // Draws  from the other-side of the RGBA and Alpha images
		{"Overflow", image.Pt(35, 35)},   // Panic with image.RGBA and image.Alpha
		{"OverflowY", image.Pt(0, 30)},   // Panic with image.RGBA and image.Alpha
		{"OverflowX", image.Pt(35, 15)},  // Draws  from the other-side of the RGBA and Alpha images
	}

	// Generate the sample image with all offsets and images.
	for _, offset := range offsets {

		// This experiment's targets are image.RGBA and image.Alpha.
		// But feel free to try out any of them.
		images := []struct {
			Type  string
			Image draw.Image
		}{
			// They both Panics on Negative, NegativeY, Overflow, and OverflowY.
			// They also draw the skipped part of the vector higher/lower (depends)
			// from the other side of the image on NegativeX and OverflowX.
			{"Alpha", image.NewAlpha(imageRect)},
			{"RGBA", image.NewRGBA(imageRect)},

			// They work totally fine. But it's just for avoiding the real issue.
			// Feel free to uncomment and test them out.
			// {"WrappedAlpha", ImageWrapper{image.NewAlpha(imageRect)}},
			// {"WrappedRGBA", ImageWrapper{image.NewRGBA(imageRect)}},

			// They all work totally fine.
			{"NRGBA", image.NewNRGBA(imageRect)},
			// {"NRGBA64", image.NewNRGBA64(imageRect)},
			// {"RGBA64", image.NewRGBA64(imageRect)},
			// {"Alpha16", image.NewAlpha16(imageRect)},
			// {"CMYK", image.NewCMYK(imageRect)},
			// {"Gray", image.NewGray(imageRect)},
			// {"Gray16", image.NewGray16(imageRect)},
		}

		for _, img := range images {
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("[FTL] [IMG: %s] [OFF: %s] Panic > %v \n", img.Type, offset.Type, r)
					}
				}()

				vec := createVector(25, 25)

				// If 'r' be greater than the vector's bounds, it'll panic.
				// Thus we'll use its bounds for now to avoid unnecessary crashes.
				//
				// THIS IS WHERE EVERYTHING HAPPENS.
				vec.Draw(img.Image, vec.Bounds().Add(offset.Point), uniformImage, image.Point{})

				err := saveImage(img.Image, fmt.Sprintf("assets/%s-%s-current.png", offset.Type, img.Type))
				if err != nil {
					fmt.Printf("[ERR] [IMG: %s] [OFF: %s] Failed to save the image > %s\n", img.Type, offset.Type, err.Error())
					return
				}
				fmt.Printf("[INF] [IMG: %s] [OFF: %s] Drawn and saved successfully\n", img.Type, offset.Type)
			}()
		}
	}
}

// createVector create a basic rectangle vector just for visualizing the issue.
//
// It's NOT related to the issue.
func createVector(w, h int) *vector.Rasterizer {
	x, y := float32(w), float32(h)

	vec := vector.NewRasterizer(w, h)
	vec.LineTo(x, 0) // top right
	vec.LineTo(x, y) // bottom right
	vec.LineTo(0, y) // bottom left
	vec.LineTo(0, 0) // top left
	return vec
}

// saveImage encodes and write the img to path.
//
// It's NOT related to the issue.
func saveImage(img image.Image, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	if err = png.Encode(f, img); err != nil {
		return err
	}

	return nil
}
