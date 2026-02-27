package main

import (
	"syscall/js"
)

const (
	width       = 800
	height      = 600
	mouseRadius = 50
)

type RGBA struct {
	R, G, B, A uint8
}

type Screen struct {
	Width  int
	Height int
	Pix    []uint8
	MouseX float64
	MouseY float64
}

type Ball struct {
	X, Y   float64
	VX, VY float64
	Radius float64
	Color  RGBA
}

func main() {
	screen := newBuffer(width, height)

	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "canvas")
	ctx := canvas.Call("getContext", "2d")

	imageData := ctx.Call("createImageData", width, height)
	jsPixels := js.Global().Get("Uint8ClampedArray").New(len(screen.Pix))
	dataArray := imageData.Get("data")

	mouseListener := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		rect := canvas.Call("getBoundingClientRect")

		screen.MouseX = event.Get("clientX").Float() - rect.Get("left").Float()
		screen.MouseY = event.Get("clientY").Float() - rect.Get("top").Float()

		return nil
	})

	canvas.Call("addEventListener", "mousemove", mouseListener)

	balls := []Ball{
		{X: 100, Y: 100, VX: 2, VY: 1.5, Radius: 40, Color: RGBA{255, 0, 0, 255}},
		{X: 400, Y: 300, VX: -1, VY: 2, Radius: 60, Color: RGBA{0, 255, 0, 255}},
		{X: 600, Y: 100, VX: 1.2, VY: -1.2, Radius: 35, Color: RGBA{0, 0, 255, 255}},
	}
	mouseColor := RGBA{255, 255, 0, 255}

	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		for i := range balls {
			balls[i].X += balls[i].VX
			balls[i].Y += balls[i].VY

			if balls[i].X-balls[i].Radius < 0 || balls[i].X+balls[i].Radius > float64(width) {
				balls[i].VX *= -1
			}
			if balls[i].Y-balls[i].Radius < 0 || balls[i].Y+balls[i].Radius > float64(height) {
				balls[i].VY *= -1
			}
		}

		for y := 0; y < screen.Height; y++ {
			for x := 0; x < screen.Width; x++ {
				var totalIntensity, totalWeight float64
				var rSum, gSum, bSum float64

				const maxContribution = 20.0

				for i := -1; i < len(balls); i++ {
					var bx, by, br float64
					var bColor RGBA

					if i == -1 {
						bx, by, br, bColor = screen.MouseX, screen.MouseY, float64(mouseRadius), mouseColor
					} else {
						b := balls[i]
						bx, by, br, bColor = b.X, b.Y, b.Radius, b.Color
					}

					dx := float64(x) - bx
					dy := float64(y) - by
					distSq := dx*dx + dy*dy

					var intensity float64
					if distSq < 1 {
						intensity = 2.0
					} else {
						intensity = (br * br) / distSq
					}
					totalIntensity += intensity

					weight := (br * br) / (distSq + 1)
					rSum += float64(bColor.R) * weight
					gSum += float64(bColor.G) * weight
					bSum += float64(bColor.B) * weight
					totalWeight += weight
				}

				var r, g, b uint8
				if totalIntensity > 0.01 {
					factor := 1.0
					if totalIntensity < 1.0 {
						factor = totalIntensity * totalIntensity
					}

					if totalWeight > 0 {
						r = uint8((rSum / totalWeight) * factor)
						g = uint8((gSum / totalWeight) * factor)
						b = uint8((bSum / totalWeight) * factor)
					}
				}

				idx := (y*screen.Width + x) * 4
				screen.Pix[idx] = r
				screen.Pix[idx+1] = g
				screen.Pix[idx+2] = b
				screen.Pix[idx+3] = 255
			}
		}

		js.CopyBytesToJS(jsPixels, screen.Pix)
		dataArray.Call("set", jsPixels)
		ctx.Call("putImageData", imageData, 0, 0)

		js.Global().Call("requestAnimationFrame", renderFrame)
		return nil
	})

	js.Global().Call("requestAnimationFrame", renderFrame)

	select {}
}

func newBuffer(w, h int) *Screen {
	return &Screen{
		Width:  w,
		Height: h,
		Pix:    make([]uint8, w*h*4),
	}
}

