package main

import (
	"context"
	_ "embed"
	"log"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/remote"
)

//go:embed fetch.js
var jsFetch string

type ChapterResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ChapterID int      `json:"chapter_id"`
		Title     string   `json:"title"`
		Images    []string `json:"images"`
	} `json:"data"`
}

func main() {
	// Local LP instance must be started somewhere.
	allocCtx, allocCancel := remote.NewAllocator(context.Background(), "ws://127.0.0.1:9222/")
	defer allocCancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	err := chromedp.Do(ctx, chromedp.Navigate("https://batcave.biz"))
	if err != nil {
		log.Fatal(err)
	}

	v, err := chromedp.Run(ctx, chromedp.Evaluate[ChapterResponse](jsFetch, chromedp.EvalAwaitPromise))
	if err != nil {
		log.Fatal(err)
	}

	d := InitDownloader()
	d.DownloadChapter(context.Background(), v.Data.Images)
}
