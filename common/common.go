package common

import (
	"encoding/json"
	"fmt"
)

const (
	webp = ".webp"
	jpeg = ".jpeg"
	png  = ".png"
	gif  = ".gif"
)

// ImageExtFromMimeType returns file extension by given MIME.
func ImageExtFromMimeType(mime string) (string, error) {
	switch mime {
	case "image/webp":
		return webp, nil
	case "image/jpeg":
		return jpeg, nil
	case "image/png":
		return png, nil
	case "image/gif":
		return gif, nil
	default:
		return "", fmt.Errorf("unknown MIME: %s", mime)
	}
}

// UnescapeJSON decodes JSON escapes left in s by APIs that double-encode
// strings (e.g. `\/` -> `/`, `é` -> `é`). Returns s unchanged if it
// isn't a valid JSON string body.
func UnescapeJSON(s string) string {
	var out string
	if err := json.Unmarshal([]byte(`"`+s+`"`), &out); err != nil {
		return s
	}
	return out
}
