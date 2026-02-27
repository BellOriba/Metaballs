package main

import (
	"strconv"
	"syscall/js"
)

const (
	width       = 400
	height      = 300
)

var (
	mouseRadius = 25.0
	tailIntensity = 0.0
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

	radiusChange := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		val := args[0].Get("target").Get("value").String()
		newRadius, _ := strconv.ParseFloat(val, 64)
		mouseRadius = newRadius
		return nil
	})
	doc.Call("getElementById", "radiusInput").Call("addEventListener", "input", radiusChange)

	tailChange := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		val := args[0].Get("target").Get("value").String()
		newTail, _ := strconv.ParseFloat(val, 64)
		tailIntensity = newTail
		return nil
	})
	doc.Call("getElementById", "tailInput").Call("addEventListener", "input", tailChange)

	mouseListener := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		rect := canvas.Call("getBoundingClientRect")

		screen.MouseX = (event.Get("clientX").Float() - rect.Get("left").Float()) / 2.0
		screen.MouseY = (event.Get("clientY").Float() - rect.Get("top").Float()) / 2.0

		return nil
	})

	canvas.Call("addEventListener", "mousemove", mouseListener)

	balls := []Ball{
		{X: 100, Y: 100, VX: 1, VY: 0.5, Radius: 20, Color: RGBA{250, 2, 12, 255}},
		{X: 300, Y: 200, VX: -1, VY: 1, Radius: 30, Color: RGBA{42, 232, 0, 255}},
		{X: 50, Y: 100, VX: 0.8, VY: -0.2, Radius: 13, Color: RGBA{0, 255, 255, 255}},
	}
	mouseColor := RGBA{243, 222, 0, 255}

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

		for i := 0; i < len(screen.Pix); i += 4 {
			screen.Pix[i] = uint8(float64(screen.Pix[i]) * tailIntensity)
			screen.Pix[i+1] = uint8(float64(screen.Pix[i+1]) * tailIntensity)
			screen.Pix[i+2] = uint8(float64(screen.Pix[i+2]) * tailIntensity)
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

					weight := br / (distSq + 5)
					rSum += float64(bColor.R) * weight
					gSum += float64(bColor.G) * weight
					bSum += float64(bColor.B) * weight
					totalWeight += weight
				}

				var r, g, b uint8
				if totalIntensity > 0.5 {
					factor := (totalIntensity - 0.5) * 20.0
					if factor > 1.0 { factor = 1.0 }

					if totalWeight > 0 {
						avgR := rSum / totalWeight
						avgG := gSum / totalWeight
						avgB := bSum / totalWeight

						r = uint8(avgR * factor)
						g = uint8(avgG * factor)
						b = uint8(avgB * factor)
					}
				}

				idx := (y*screen.Width + x) * 4

				if r > screen.Pix[idx] { screen.Pix[idx] = r }
				if g > screen.Pix[idx+1] { screen.Pix[idx+1] = g }
				if b > screen.Pix[idx+2] { screen.Pix[idx+2] = b }

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

