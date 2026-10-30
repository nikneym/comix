package batcave

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/chromedp/chromedp"
)

type ChapterResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ChapterID int      `json:"chapter_id"`
		Title     string   `json:"title"`
		Images    []string `json:"images"`
	} `json:"data"`
}

// jsEval fetches chapter data from inside the browser. It's a format
// string for news_id and chapter_id, so literal % must be written as %%.
const jsEval = `(async () => {
  const res = await fetch(
    "https://batcave.biz/engine/ajax/controller.php?mod=api&action=reader%%2FgetChapterData",
    {
      credentials: "omit",
      headers: {
        "User-Agent":
          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:157.0) Gecko/20100101 Firefox/157.0",
        Accept: "*/*",
        "Accept-Language": "en-US,en;q=0.9",
        "Content-Type": "application/x-www-form-urlencoded",
        "Upgrade-Insecure-Requests": "1",
        "Sec-Fetch-Dest": "document",
        "Sec-Fetch-Mode": "navigate",
        "Sec-Fetch-Site": "cross-site",
        "Sec-Fetch-User": "?1",
        Priority: "u=4",
        Pragma: "no-cache",
        "Cache-Control": "no-cache",
      },
      body: "news_id=%s&chapter_id=%s",
      method: "POST",
      mode: "cors",
    },
  );
  return res.json();
})();`

// chapterData returns title and assets for a chapter in the same order.
// If error is not nil, title is empty and assets are nil.
func (g *Grabber) chapterData(u string) (string, []string, error) {
	url, err := url.Parse(u)
	if err != nil {
		return "", nil, err
	}
	if len(url.Path) < 8 {
		return "", nil, errors.New("invalid url")
	}

	p := url.Path[8:]
	sep := strings.IndexByte(p, '/')
	if sep == -1 || sep == len(url.Path)-1 {
		return "", nil, errors.New("invalid url")
	}
	newsId, chapterId := p[:sep], p[sep+1:]

	g.browserMu.Lock()
	defer g.browserMu.Unlock()
	cd, err := chromedp.Run(g.browserCtx, chromedp.Evaluate[ChapterResponse](fmt.Sprintf(jsEval, newsId, chapterId), chromedp.EvalAwaitPromise))
	if err != nil {
		return "", nil, err
	}
	if !cd.Success {
		return "", nil, errors.New("unsuccessful")
	}

	// Title may include the name twice (separated by '/'), here we sanitize it.
	_, after, found := strings.Cut(cd.Data.Title, "/")
	if !found {
		return cd.Data.Title, cd.Data.Images, nil
	}
	title := strings.TrimSpace(after)
	if len(title) == 0 {
		return "", nil, errors.New("unexpected title format")
	}

	return title, cd.Data.Images, nil
}
