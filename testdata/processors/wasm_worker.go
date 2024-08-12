//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	fmt.Println("hello")
	js.Global().Get("self").Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		js.Global().Call("postMessage", "pong")
		return nil
	}))
}
