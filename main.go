package main

import "syscall/js"

const (
	width = 800
	height = 600
)

func main() {
	pixels := make([]uint8, width*height*4)

	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "canvas")
	ctx := canvas.Call("getContext", "2d")

	imageData := ctx.Call("createImageData", width, height)
	jsPixels := js.Global().Get("Uint8ClampedArray").New(len(pixels))

	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		js.CopyBytesToJS(jsPixels, pixels)

		imageData.Get("data").Call("set", jsPixels)
		ctx.Call("putImageData", imageData, 0, 0)

		js.Global().Call("requestAnimationFrame", renderFrame)
		return nil
	})

	js.Global().Call("requestAnimationFrame", renderFrame)

	select {}
}

