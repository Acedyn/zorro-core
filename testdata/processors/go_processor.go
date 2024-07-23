package main

import (
	"syscall/js"
)

func main() {
	js.Global().Get("self").Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		js.Global().Get("self").Call("postMessage", "pong")
		return nil
	}))

	<-make(chan struct{}, 0)
}
