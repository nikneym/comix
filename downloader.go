package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type Page struct {
	number  int
	err     error
	content []byte
}

type Downloader struct {
	client        http.Client
	browserCtx    context.Context
	browserCancel context.CancelFunc
}

func InitDownloader() Downloader {
	return Downloader{client: http.Client{}}
}

// grabPage writes page content to a buffer.
func (d *Downloader) grabPage(ctx context.Context, url string) ([]byte, error) {
	// Prepare & fire a fetch.
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://batcave.biz")
	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Pull the content.
	var buf []byte
	if res.ContentLength == -1 {
		// Fallback to `io.ReadAll`.
		buf, err = io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
	} else {
		buf = make([]byte, res.ContentLength)
		_, err = io.ReadAtLeast(res.Body, buf, len(buf))
		if err != nil {
			return nil, err
		}
	}

	return buf, nil
}

// workerLoop runs for a single worker. Picks the next page to
// download, fetches it and queues it.
func (d *Downloader) workerLoop(ctx context.Context, pages chan<- Page, assets []string, index *atomic.Int32) {
	for {
		n := int(index.Add(1) - 1)
		// No images left.
		if n >= len(assets) {
			return
		}

		buf, err := d.grabPage(ctx, assets[n])

		select {
		case pages <- Page{number: n, err: err, content: buf}:
		case <-ctx.Done():
			return
		}
	}
}

// DownloadChapter downloads a single chapter.
func (d *Downloader) DownloadChapter(ctx context.Context, assets []string) {
	// Pages are received through this.
	pages := make(chan Page, 32)
	// Workers read assets through this.
	var index atomic.Int32

	// Fire up workers.
	for range 8 {
		go d.workerLoop(ctx, pages, assets, &index)
	}

	out, err := os.Create("out.zip")
	if err != nil {
		// TODO
		panic(err)
	}
	defer out.Close()

	cbzWriter := zip.NewWriter(out)
	defer cbzWriter.Close()

	now := time.Now()
loop:
	for range len(assets) {
		select {
		// Got a new page.
		case page := <-pages:
			if page.err != nil {
				break loop
			}

			w, err := cbzWriter.CreateHeader(&zip.FileHeader{
				Name:     fmt.Sprintf("%04d.jpg", page.number),
				Method:   zip.Store,
				Modified: now,
			})
			if err != nil {
				break loop
			}

			_, err = w.Write(page.content)
			if err != nil {
				break loop
			}
		// Exit.
		case <-ctx.Done():
			break loop
		}
	}
}
