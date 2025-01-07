package downloader

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const baseURL = "https://basecamp.com/shapeup"

type Downloader struct {
	client  *http.Client
	mainCSS string
}

func New() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type Chapter struct {
	URL      string
	Title    string
	Content  string
	CSS      string
	Sections []Section
}

type Section struct {
	ID      string
	Title   string
	Content string
}

// findNode recursively searches for a node matching the given criteria
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

// extractText gets all text content from a node
func extractText(n *html.Node) string {
	var text strings.Builder
	var extract func(*html.Node)

	extract = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(n)
	return strings.TrimSpace(text.String())
}

func (d *Downloader) FetchTOC() ([]Chapter, error) {
	resp, err := d.client.Get(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch TOC: %w", err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TOC HTML: %w", err)
	}

	// Fetch the main web-book CSS
	mainCSS, err := d.fetchCSS(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch main CSS: %w", err)
	}

	// Store the main CSS for later use
	d.mainCSS = mainCSS

	var chapters []Chapter

	// Find the main TOC container
	tocDiv := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "div" &&
			hasClass(n, "toc")
	})

	if tocDiv == nil {
		return nil, fmt.Errorf("could not find table of contents")
	}

	// Process each chapter link
	var processNode func(*html.Node)
	processNode = func(n *html.Node) {
		if n.Type == html.ElementNode &&
			n.Data == "a" &&
			!strings.Contains(getAttr(n, "href"), "#") {
			href := getAttr(n, "href")
			chapter := Chapter{
				Title: extractText(n),
				URL:   "https://basecamp.com" + href,
			}
			chapters = append(chapters, chapter)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c)
		}
	}

	processNode(tocDiv)

	if len(chapters) == 0 {
		return nil, fmt.Errorf("no chapters found in table of contents")
	}

	return chapters, nil
}

// Update FetchChapter to combine both CSS sources
func (d *Downloader) FetchChapter(url string) (*Chapter, error) {
	resp, err := d.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chapter: %w", err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chapter HTML: %w", err)
	}

	chapter := &Chapter{
		URL: url,
	}

	// Fetch CSS
	css, err := d.fetchCSS(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CSS: %w", err)
	}
	chapter.CSS = css

	// Find main content
	mainContent := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "main"
	})

	if mainContent == nil {
		return nil, fmt.Errorf("could not find main content")
	}

	// Extract title and content
	titleNode := findNode(mainContent, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "h1"
	})

	if titleNode != nil {
		chapter.Title = extractText(titleNode)
	}

	// Process images before converting to string
	if err := d.processImages(mainContent); err != nil {
		return nil, fmt.Errorf("failed to process images: %w", err)
	}

	// Convert main content to string
	var content strings.Builder
	html.Render(&content, mainContent)
	chapter.Content = content.String()

	return chapter, nil
}

// Helper function to check node classes
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

// Helper function to get node attribute
func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func (d *Downloader) downloadImage(url string) (string, error) {
	// Handle relative URLs by converting to absolute
	if strings.HasPrefix(url, "/") {
		url = "https://basecamp.com" + url
	}

	resp, err := d.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Read image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Determine MIME type
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mime.TypeByExtension(path.Ext(url))
	}

	// Convert to base64
	b64Data := base64.StdEncoding.EncodeToString(imageData)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, b64Data), nil
}
func (d *Downloader) processImages(doc *html.Node) error {
	var processNode func(*html.Node) error
	processNode = func(n *html.Node) error {
		if n.Type == html.ElementNode && n.Data == "img" {
			// Find src attribute
			for i, attr := range n.Attr {
				if attr.Key == "src" {
					// Download and convert image
					b64URL, err := d.downloadImage(attr.Val)
					if err != nil {
						return err
					}
					// Update src attribute
					n.Attr[i].Val = b64URL
					break
				}
			}
		}

		// Process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := processNode(c); err != nil {
				return err
			}
		}
		return nil
	}

	return processNode(doc)
}

func (d *Downloader) fetchCSS(doc *html.Node) (string, error) {
	var cssContent string

	// Find CSS link in head
	cssLink := findNode(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "link" {
			for _, attr := range n.Attr {
				if attr.Key == "rel" && attr.Val == "stylesheet" &&
					!strings.Contains(attr.Val, "typekit") {
					return true
				}
			}
		}
		return false
	})

	if cssLink == nil {
		return "", fmt.Errorf("could not find CSS link")
	}

	// Get href attribute and handle relative URLs
	var cssURL string
	for _, attr := range cssLink.Attr {
		if attr.Key == "href" {
			if strings.HasPrefix(attr.Val, "/") {
				cssURL = "https://basecamp.com" + attr.Val
			} else {
				cssURL = attr.Val
			}
			break
		}
	}

	// Download CSS
	resp, err := d.client.Get(cssURL)
	if err != nil {
		return "", fmt.Errorf("failed to download CSS: %w", err)
	}
	defer resp.Body.Close()

	css, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read CSS: %w", err)
	}

	cssContent = string(css)
	return cssContent, nil
}
