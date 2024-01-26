package main

import (
	"github.com/Acedyn/zorro-core/pkg/manager"

	"github.com/teamortix/golang-wasm/wasm"
)

func add(x int, y int) (int, error) {
	return x + y, nil
}

func main() {
	wasm.Expose("invokeAction", manager.InvokeAction)
	wasm.Expose("listActions", manager.InvokeAction)
	wasm.Ready()
	<-make(chan struct{}, 0)
}
