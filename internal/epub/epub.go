package epub

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/benjaminkitt/shape-up-downloader/internal/downloader"
	"github.com/go-shiori/go-epub"
	"golang.org/x/net/html"
)

type EPUBConverter struct {
	OutputPath string
}

func NewEPUBConverter(outputPath string) *EPUBConverter {
	// Add .epub extension if not present
	if !strings.HasSuffix(outputPath, ".epub") {
		outputPath = outputPath + ".epub"
	}

	return &EPUBConverter{
		OutputPath: outputPath,
	}
}

func (e *EPUBConverter) processChapterContent(content string) (string, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse chapter content: %w", err)
	}

	// Find the main content div
	contentDiv := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "div" &&
			hasClass(n, "content")
	})

	if contentDiv == nil {
		return content, nil // Return original if not found
	}

	// Find and remove template elements
	templates := findAllNodes(contentDiv, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "template"
	})
	for _, t := range templates {
		t.Parent.RemoveChild(t)
	}

	// Find and remove nav elements
	navs := findAllNodes(contentDiv, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "nav"
	})
	for _, nav := range navs {
		nav.Parent.RemoveChild(nav)
	}

	// Find and remove footer elements
	footers := findAllNodes(contentDiv, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "footer"
	})
	for _, footer := range footers {
		footer.Parent.RemoveChild(footer)
	}

	// Render just the content div
	var buf strings.Builder
	if err := html.Render(&buf, contentDiv); err != nil {
		return "", fmt.Errorf("failed to render content: %w", err)
	}

	return buf.String(), nil
}

func (e *EPUBConverter) Convert(chapters []downloader.Chapter, css string) error {
	book, err := epub.NewEpub("Shape Up")
	if err != nil {
		return fmt.Errorf("failed to create new epub: %w", err)
	}

	// Set metadata
	book.SetAuthor("Ryan Singer")
	book.SetDescription("Stop Running in Circles and Ship Work that Matters")
	book.SetLang("en")

	// Add CSS
	encodedCSS := url.QueryEscape(css)
	cssDataURL := "data:text/css," + encodedCSS
	cssPath, err := book.AddCSS(cssDataURL, "styles.css")
	if err != nil {
		return fmt.Errorf("failed to add CSS: %w", err)
	}

	// Add each chapter
	for _, chapter := range chapters {
		// Process content
		cleanContent, err := e.processChapterContent(chapter.Content)
		if err != nil {
			return fmt.Errorf("failed to process chapter %s: %w", chapter.Title, err)
		}

		// Handle images
		doc, err := html.Parse(strings.NewReader(cleanContent))
		if err != nil {
			return fmt.Errorf("failed to parse chapter content: %w", err)
		}

		// Find all images
		images := findAllNodes(doc, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "img"
		})

		// Process each image
		for _, img := range images {
			src := getAttr(img, "src")
			if src == "" {
				continue
			}

			// If already a data URL, use it directly
			if strings.HasPrefix(src, "data:") {
				imgPath, err := book.AddImage(src, fmt.Sprintf("image_%d.jpg", time.Now().UnixNano()))
				if err != nil {
					continue
				}
				setAttr(img, "src", imgPath)
				continue
			}

			// Download and convert image to data URL
			resp, err := http.Get(src)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			// Read image data
			imgData, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			// Determine MIME type
			mimeType := resp.Header.Get("Content-Type")
			if mimeType == "" {
				mimeType = mime.TypeByExtension(path.Ext(src))
			}

			// Convert to base64 data URL
			b64Data := base64.StdEncoding.EncodeToString(imgData)
			dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, b64Data)

			// Add image to epub using data URL
			imgPath, err := book.AddImage(dataURL, fmt.Sprintf("image_%d.jpg", time.Now().UnixNano()))
			if err != nil {
				continue
			}

			// Update image src to reference the added image
			setAttr(img, "src", imgPath)
		}

		// Render updated content
		var buf strings.Builder
		if err := html.Render(&buf, doc); err != nil {
			return fmt.Errorf("failed to render chapter content: %w", err)
		}

		// Add chapter with processed content
		_, err = book.AddSection(buf.String(), chapter.Title, "", cssPath)
		if err != nil {
			return fmt.Errorf("failed to add chapter %s: %w", chapter.Title, err)
		}
	}

	return book.Write(e.OutputPath)
}

// Add these helper functions to the epub package

func findNode(n *html.Node, criteria func(*html.Node) bool) *html.Node {
	if criteria(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findNode(c, criteria); result != nil {
			return result
		}
	}
	return nil
}

func findAllNodes(n *html.Node, criteria func(*html.Node) bool) []*html.Node {
	var nodes []*html.Node
	var finder func(*html.Node)

	finder = func(n *html.Node) {
		if criteria(n) {
			nodes = append(nodes, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			finder(c)
		}
	}

	finder(n)
	return nodes
}

func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			for _, c := range classes {
				if c == class {
					return true
				}
			}
		}
	}
	return false
}

func setAttr(n *html.Node, key, value string) {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			n.Attr[i].Val = value
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: value})
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
