package main

import (
	"github.com/teamortix/golang-wasm/wasm"
)

func add(x int, y int) (int, error) {
	return x + y, nil
}

func main() {
	wasm.Expose("add", add)
	wasm.Ready()
	<-make(chan struct{}, 0)
}
