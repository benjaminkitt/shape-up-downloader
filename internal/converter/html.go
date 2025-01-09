package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/benjaminkitt/shape-up-downloader/internal/downloader"
	"golang.org/x/net/html"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Shape Up</title>
    <style>{{.CSS}}</style>
</head>
<body>
    <div class="content">
        <h1 class="landing-title landing-title--large">Shape Up</h1>
        <p class="landing-subtitle">Stop Running in Circles<br class="linebreak"> and Ship Work that Matters</p>
        <p class="landing-author"><em>by Ryan Singer</em></p>
        <div id="toc" class="toc">{{.TOC}}</div>
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
	baseConverter
}

func NewHTMLConverter(outputDir string) *HTMLConverter {
	return &HTMLConverter{
		OutputDir: outputDir,
	}
}

func (c *HTMLConverter) Convert(chapters []downloader.Chapter, css string) error {
	if len(chapters) == 0 {
		return fmt.Errorf("no chapters provided for conversion")
	}

	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Extract TOC from first chapter
	doc, err := html.Parse(strings.NewReader(chapters[0].Content))
	if err != nil {
		return fmt.Errorf("failed to parse main page: %w", err)
	}

	tocHTML, err := c.extractTOC(doc)
	if err != nil {
		return fmt.Errorf("failed to extract TOC: %w", err)
	}

	// Process chapters
	for i := range chapters {
		processedContent, err := c.processChapterContent(chapters[i].Content)
		if err != nil {
			return fmt.Errorf("failed to process chapter %s: %w", chapters[i].Title, err)
		}
		doc, err := html.Parse(strings.NewReader(processedContent))
		if err != nil {
			return fmt.Errorf("failed to parse processed content: %w", err)
		}
		c.processLinks(doc)

		// Render the processed document back to string
		var buf strings.Builder
		if err := html.Render(&buf, doc); err != nil {
			return fmt.Errorf("failed to render processed content: %w", err)
		}
		chapters[i].Content = buf.String()
	}

	data := struct {
		CSS   string
		TOC   string
		Parts []Part
	}{
		CSS:   css,
		TOC:   tocHTML,
		Parts: c.organizeParts(chapters),
	}

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

	return tmpl.Execute(f, data)
}

func (c *HTMLConverter) organizeParts(chapters []downloader.Chapter) []Part {
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

func (c *HTMLConverter) processLinks(node *html.Node) {
	links := findBookLinks(node)
	for _, link := range links {
		href := getAttr(link, "href")

		// Strip basecamp prefix if present
		href = strings.TrimPrefix(href, "https://basecamp.com/shapeup/")
		href = strings.TrimPrefix(href, "/shapeup/")

		// If href already contains a #, preserve only the section reference
		if strings.Contains(href, "#") {
			parts := strings.Split(href, "#")
			href = "#" + parts[1]
		} else {
			href = "#" + href
		}

		setAttr(link, "href", href)
	}
}

func (c *HTMLConverter) extractTOC(doc *html.Node) (string, error) {
	tocDiv := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "div" &&
			hasClass(n, "toc")
	})

	if tocDiv == nil {
		return "", fmt.Errorf("could not find TOC element")
	}

	// Process TOC links
	c.processLinks(tocDiv)

	var buf strings.Builder
	if err := html.Render(&buf, tocDiv); err != nil {
		return "", fmt.Errorf("failed to render TOC: %w", err)
	}

	return buf.String(), nil
}
