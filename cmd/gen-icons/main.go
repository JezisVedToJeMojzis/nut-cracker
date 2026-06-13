// Command gen-icons writes the PWA app icons (web/static/icon-192.png and
// icon-512.png): a white peanut silhouette on the brand green. Run with
// `go run ./cmd/gen-icons` whenever the icon design changes.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	for _, size := range []int{192, 512} {
		img := drawIcon(size)
		f, err := os.Create(fmt.Sprintf("web/static/icon-%d.png", size))
		if err != nil {
			return err
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	fmt.Println("wrote web/static/icon-192.png and icon-512.png")
	return nil
}

func drawIcon(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	bg := color.RGBA{0x10, 0xb9, 0x81, 0xff}  // brand green
	fg := color.RGBA{0xff, 0xff, 0xff, 0xff}   // white peanut
	n := float64(size)

	// Peanut = two overlapping circles (a figure-8) centred horizontally.
	r := 0.20 * n
	cy := 0.5 * n
	cx1, cx2 := 0.40*n, 0.60*n

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bg)
			fx, fy := float64(x)+0.5, float64(y)+0.5
			d1 := math.Hypot(fx-cx1, fy-cy)
			d2 := math.Hypot(fx-cx2, fy-cy)
			if d1 <= r || d2 <= r {
				img.Set(x, y, fg)
			}
		}
	}
	return img
}
