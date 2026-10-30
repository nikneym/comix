package common

import (
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
