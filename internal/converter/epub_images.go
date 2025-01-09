package converter

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-shiori/go-epub"
	"golang.org/x/net/html"
)

func (e *EPUBConverter) processImages(doc *html.Node, book *epub.Epub) error {
	images := findAllNodes(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "img"
	})

	for _, img := range images {
		src := getAttr(img, "src")
		if src == "" {
			continue
		}

		// Handle data URLs directly
		if strings.HasPrefix(src, "data:") {
			imgPath, err := book.AddImage(src, fmt.Sprintf("image_%d.jpg", time.Now().UnixNano()))
			if err != nil {
				continue
			}
			setAttr(img, "src", imgPath)
			continue
		}

		// Download and process external images
		imgPath, err := e.downloadAndAddImage(src, book)
		if err != nil {
			continue
		}
		setAttr(img, "src", imgPath)
	}

	return nil
}

func (e *EPUBConverter) downloadAndAddImage(src string, book *epub.Epub) (string, error) {
	resp, err := http.Get(src)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mime.TypeByExtension(path.Ext(src))
	}

	b64Data := base64.StdEncoding.EncodeToString(imgData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, b64Data)

	return book.AddImage(dataURL, fmt.Sprintf("image_%d.jpg", time.Now().UnixNano()))
}
