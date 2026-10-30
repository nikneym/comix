package main

import (
	"comix/batcave"
	"context"
	_ "embed"

	"github.com/chromedp/chromedp/remote"
)

func main() {
	allocCtx, allocCancel := remote.NewAllocator(context.Background(), "ws://127.0.0.1:9222/")
	defer allocCancel()

	g, err := batcave.Init(allocCtx)
	if err != nil {
		panic(err)
	}
	defer g.Close()
}
