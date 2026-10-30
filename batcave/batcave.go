package batcave

import (
	"archive/zip"
	"comix/common"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/chromedp"
)

type Grabber struct {
	browserMu     sync.Mutex
	browserCancel context.CancelFunc
	browserCtx    context.Context
	client        http.Client
}

func Init(allocatorCtx context.Context) (*Grabber, error) {
	browserCtx, browserCancel := chromedp.NewContext(allocatorCtx)
	// Initial navigation.
	err := chromedp.Do(browserCtx, chromedp.Navigate("https://batcave.biz"))
	if err != nil {
		browserCancel()
		return nil, err
	}

	return &Grabber{
		browserMu:     sync.Mutex{},
		browserCancel: browserCancel,
		browserCtx:    browserCtx,
		client:        http.Client{},
	}, nil
}

func (g *Grabber) Close() {
	g.browserCancel()
}

// pullContent writes page content and it's extension to a buffer.
func (g *Grabber) pullContent(ctx context.Context, url string) ([]byte, string, error) {
	// Prepare & fire a fetch.
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Referer", "https://batcave.biz")
	res, err := g.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, "", errors.New("unexpected status code")
	}

	// Pull the content.
	var buf []byte
	if res.ContentLength <= 0 {
		// Fallback to `io.ReadAll`.
		buf, err = io.ReadAll(res.Body)
		if err != nil {
			return nil, "", err
		}
	} else {
		buf = make([]byte, res.ContentLength)
		_, err = io.ReadAtLeast(res.Body, buf, len(buf))
		if err != nil {
			return nil, "", err
		}
	}

	// Get extension.
	ext, err := common.ImageExtFromMimeType(http.DetectContentType(buf))
	if err != nil {
		return nil, ext, err
	}

	return buf, ext, nil
}

// page represents an operation that's done for each asset.
// if err is not nil; content and ext are empty.
type page struct {
	number  int
	err     error
	ext     string
	content []byte
}

// workerLoop runs for a single worker. Picks the next page to
// download, fetches it and queues it.
func (g *Grabber) workerLoop(ctx context.Context, pages chan<- page, assets []string, index *atomic.Int32) {
	for {
		n := int(index.Add(1) - 1)
		// No images left.
		if n >= len(assets) {
			return
		}

		buf, ext, err := g.pullContent(ctx, assets[n])

		select {
		case pages <- page{number: n, err: err, ext: ext, content: buf}:
		case <-ctx.Done():
			return
		}
	}
}

// writeChapterCBZ writes a chapter to given writer in CBZ format.
// It is up to caller to close the writer if necessary. On error, the
// CBZ written so far is left incomplete.
func (g *Grabber) writeChapterCBZ(ctx context.Context, w io.Writer, assets []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cbzWriter := zip.NewWriter(w)
	// Will be used as modified time of assets.
	now := time.Now()

	// Pages are received through this.
	pages := make(chan page, 32)
	// Workers read assets through this.
	var index atomic.Int32

	// Fire up workers.
	// TODO: Make worker count configurable.
	for range 8 {
		go g.workerLoop(ctx, pages, assets, &index)
	}

	// Write to CBZ file.
	for range len(assets) {
		select {
		// Got a new page.
		case page := <-pages:
			if page.err != nil {
				return page.err
			}

			w, err := cbzWriter.CreateHeader(&zip.FileHeader{
				Name:     fmt.Sprintf("%04d%s", page.number, page.ext),
				Method:   zip.Store,
				Modified: now,
			})
			if err != nil {
				return err
			}

			_, err = w.Write(page.content)
			if err != nil {
				return err
			}
		// Exit.
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Writes the central directory; the archive is unreadable without it.
	return cbzWriter.Close()
}

func (g *Grabber) DownloadChapter(ctx context.Context, u string) error {
	title, assets, err := g.chapterData(u)
	if err != nil {
		return err
	}

	filename := title + ".cbz"
	// Write to a temporary file and rename it on success, so a failed
	// download neither leaves a partial CBZ behind nor replaces an existing one.
	tmpname := filename + ".part"
	out, err := os.Create(tmpname)
	if err != nil {
		return err
	}

	err = g.writeChapterCBZ(ctx, out, assets)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmpname, filename)
	}
	if err != nil {
		_ = os.Remove(tmpname)
	}
	return err
}
