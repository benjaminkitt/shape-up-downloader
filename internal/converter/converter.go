package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gihub.com/benjaminkitt/shape-up-downloader/internal/downloader"
	"golang.org/x/net/html"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Shape Up</title>
    <style>
        {{.CSS}}
    </style>
</head>
<body>
    <div class="content">
        <h1 class="landing-title landing-title--large">Shape Up</h1>

        <p class="landing-subtitle">Stop Running in Circles<br class="linebreak"> and Ship Work that Matters
        </p><p class="landing-author"><em>by Ryan Singer</em></p>
        <div id="toc" class="toc">
            {{.TOC}}
        </div>
    </div>
    <main>
        {{range .Parts}}
            {{range .Chapters}}
            <article id="{{trimPrefix .URL "https://basecamp.com/shapeup/"}}">
                {{.Content}}
            </article>
            {{end}}
        {{end}}
    </main>
</body>
</html>`

type Part struct {
	Title    string
	Chapters []downloader.Chapter
}

type HTMLConverter struct {
	OutputDir string
}

func NewHTMLConverter(outputDir string) *HTMLConverter {
	return &HTMLConverter{
		OutputDir: outputDir,
	}
}

// Helper functions for DOM manipulation
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

func hasLink(n *html.Node, href string) bool {
	link := findNode(n, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a"
	})
	return link != nil && getAttr(link, "href") == href
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
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

// Update chapter content before conversion
func (c *HTMLConverter) processChapterContent(content string) (string, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse chapter content: %w", err)
	}

	updateShapeUpButton(doc)
	updateChapterLinks(doc)

	var buf strings.Builder
	if err := html.Render(&buf, doc); err != nil {
		return "", fmt.Errorf("failed to render processed content: %w", err)
	}

	return buf.String(), nil
}

func (c *HTMLConverter) Convert(chapters []downloader.Chapter, css string) error {
	if len(chapters) == 0 {
		return fmt.Errorf("no chapters provided for conversion")
	}

	// Create output directory
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Process each chapter's content
	for i := range chapters {
		processedContent, err := c.processChapterContent(chapters[i].Content)
		if err != nil {
			return fmt.Errorf("failed to process chapter %s: %w", chapters[i].Title, err)
		}
		chapters[i].Content = processedContent
	}

	// Extract TOC from first chapter (which should be the main page)
	doc, err := html.Parse(strings.NewReader(chapters[0].Content))
	if err != nil {
		return fmt.Errorf("failed to parse main page: %w", err)
	}

	tocHTML, err := extractTOC(doc)
	if err != nil {
		return fmt.Errorf("failed to extract TOC: %w", err)
	}

	data := struct {
		CSS   string
		TOC   string
		Parts []Part
	}{
		CSS:   css,
		TOC:   tocHTML,
		Parts: organizeParts(chapters),
	}

	// Execute template with updated data
	tmpl, err := template.New("book").Funcs(template.FuncMap{
		"trimPrefix": strings.TrimPrefix,
	}).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	outputPath := filepath.Join(c.OutputDir, "index.html")
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	return nil
}

func updateShapeUpButton(doc *html.Node) {
	// Find the button with the specific class and data-action
	button := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "button" &&
			hasClass(n, "intro__book-title") &&
			getAttr(n, "data-action") == "click->sidebar#open"
	})

	if button != nil {
		// Convert button to anchor while preserving SVG and text content
		a := &html.Node{
			Type: html.ElementNode,
			Data: "a",
			Attr: []html.Attribute{
				{Key: "href", Val: "#toc"},
				{Key: "class", Val: "intro__book-title button hidden-print"},
				{Key: "aria-label", Val: "Shape Up Table Of Contents"},
			},
		}

		// Move all child nodes (SVG and text span) to the new anchor
		for c := button.FirstChild; c != nil; {
			next := c.NextSibling
			button.RemoveChild(c)
			a.AppendChild(c)
			c = next
		}

		// Replace button with anchor
		button.Parent.InsertBefore(a, button)
		button.Parent.RemoveChild(button)
	}
}
func updateChapterLinks(doc *html.Node) {
	links := findAllNodes(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "a" &&
			strings.HasPrefix(getAttr(n, "href"), "/shapeup/")
	})

	for _, link := range links {
		href := getAttr(link, "href")

		// Handle section links (contains #)
		if strings.Contains(href, "#") {
			parts := strings.Split(href, "#")
			setAttr(link, "href", "#"+parts[1])
			continue
		}

		// Handle chapter links
		chapterID := strings.TrimPrefix(href, "/shapeup/")
		setAttr(link, "href", "#"+chapterID)
	}
}

func extractTOC(doc *html.Node) (string, error) {
	// Find the main TOC container
	tocDiv := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "div" &&
			hasClass(n, "toc")
	})

	if tocDiv == nil {
		return "", fmt.Errorf("could not find TOC element")
	}

	// Update all links to be relative
	links := findAllNodes(tocDiv, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "a" &&
			strings.HasPrefix(getAttr(n, "href"), "/shapeup/")
	})

	for _, link := range links {
		href := getAttr(link, "href")
		chapterID := strings.TrimPrefix(href, "/shapeup/")
		setAttr(link, "href", "#"+chapterID)
	}

	// Render the TOC HTML
	var buf strings.Builder
	if err := html.Render(&buf, tocDiv); err != nil {
		return "", fmt.Errorf("failed to render TOC: %w", err)
	}

	return buf.String(), nil
}

func organizeParts(chapters []downloader.Chapter) []Part {
	if len(chapters) < 5 {
		return []Part{{
			Title:    "Contents",
			Chapters: chapters,
		}}
	}

	return []Part{
		{
			Title:    "Introduction",
			Chapters: chapters[0:3],
		},
		{
			Title:    "Part 1: Shaping",
			Chapters: chapters[3:8],
		},
		{
			Title:    "Part 2: Betting",
			Chapters: chapters[8:11],
		},
		{
			Title:    "Part 3: Building",
			Chapters: chapters[11:18],
		},
		{
			Title:    "Appendices",
			Chapters: chapters[18:],
		},
	}
}
