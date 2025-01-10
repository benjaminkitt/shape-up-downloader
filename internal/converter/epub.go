package converter

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/benjaminkitt/shape-up-downloader/internal/downloader"
	"github.com/go-shiori/go-epub"
	"golang.org/x/net/html"
)

type EPUBConverter struct {
	OutputPath string
	baseConverter
}

func NewEPUBConverter(outputPath string) *EPUBConverter {
	if !strings.HasSuffix(outputPath, ".epub") {
		outputPath = outputPath + ".epub"
	}

	return &EPUBConverter{
		OutputPath: outputPath,
	}
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

	// Add title page as first section
	titlePage, err := e.createTitlePage()
	if err != nil {
		return fmt.Errorf("failed to create title page: %w", err)
	}
	_, err = book.AddSection(titlePage, "Title Page", "", "")
	if err != nil {
		return fmt.Errorf("failed to add title page: %w", err)
	}

	// Extract and add TOC as first chapter
	doc, err := html.Parse(strings.NewReader(chapters[0].Content))
	if err != nil {
		return fmt.Errorf("failed to parse main page: %w", err)
	}

	tocHTML, err := e.extractTOC(doc, chapters)
	if err != nil {
		return fmt.Errorf("failed to extract TOC: %w", err)
	}

	cleanToc, err := cleanHTML(tocHTML)
	if err != nil {
		return fmt.Errorf("failed to clean HTML: %w", err)
	}

	// Add TOC as second section
	_, err = book.AddSection(cleanToc, "Table of Contents", "", "")
	if err != nil {
		return fmt.Errorf("failed to add TOC: %w", err)
	}

	// Process chapters
	for _, chapter := range chapters {
		processedContent, err := e.processChapterContent(chapter.Content)
		if err != nil {
			return fmt.Errorf("failed to process chapter %s: %w", chapter.Title, err)
		}

		cleanContent, err := cleanHTML(processedContent)
		if err != nil {
			return fmt.Errorf("failed to clean HTML: %w", err)
		}

		// Parse content for image processing
		doc, err := html.Parse(strings.NewReader(cleanContent))
		if err != nil {
			return fmt.Errorf("failed to parse chapter content: %w", err)
		}

		// Process links in the chapter
		e.processLinks(doc, chapters)

		// Process images in the chapter
		if err := e.processImages(doc, book); err != nil {
			return fmt.Errorf("failed to process images in chapter %s: %w", chapter.Title, err)
		}

		// Render the processed content
		var buf strings.Builder
		if err := html.Render(&buf, doc); err != nil {
			return fmt.Errorf("failed to render chapter content: %w", err)
		}

		// Add processed chapter to epub
		_, err = book.AddSection(buf.String(), chapter.Title, "", "")
		if err != nil {
			return fmt.Errorf("failed to add chapter %s: %w", chapter.Title, err)
		}
	}

	return book.Write(e.OutputPath)
}

func (e *EPUBConverter) createTitlePage() (string, error) {
	resp, err := http.Get("https://basecamp.com/assets/images/books/shapeup/cover_summary.jpeg")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	b64Data := base64.StdEncoding.EncodeToString(imgData)
	imgSrc := fmt.Sprintf("data:image/jpeg;base64,%s", b64Data)

	return fmt.Sprintf(`
			<div class="content" style="display: flex; flex-direction: column; justify-content: center; align-items: center; min-height: 100vh;">
					<img src="%s" alt="Shape Up Cover" style="max-width: 80%%; margin-bottom: 2em;">
					<div style="width: 80%%; text-align: left;">
							<h1 class="landing-title landing-title--large">Shape Up</h1>
							<p class="landing-subtitle">Stop Running in Circles<br class="linebreak"> and Ship Work that Matters</p>
							<p class="landing-author"><em>by Ryan Singer</em></p>
					</div>
			</div>`, imgSrc), nil
}

func (e *EPUBConverter) processLinks(node *html.Node, chapters []downloader.Chapter) {
	links := findBookLinks(node)
	for _, link := range links {
		href := getAttr(link, "href")
		chapterNum := findChapterNumberByURL(href, chapters)

		if strings.Contains(href, "#") {
			parts := strings.Split(href, "#")
			section := parts[1]
			setAttr(link, "href", fmt.Sprintf("section%04d.xhtml#%s", chapterNum, section))
		} else {
			setAttr(link, "href", fmt.Sprintf("section%04d.xhtml", chapterNum))
		}
	}
}

func findChapterNumberByURL(href string, chapters []downloader.Chapter) int {
	// Strip any section reference and prefixes
	href = strings.Split(href, "#")[0]
	href = strings.TrimPrefix(href, "https://basecamp.com")
	fullURL := "https://basecamp.com" + href

	for _, chapter := range chapters {
		if chapter.URL == fullURL {
			return chapter.Number + 2 // Account for title page and TOC
		}
	}
	return 3 // Default to first section if not found
}

func (e *EPUBConverter) extractTOC(doc *html.Node, chapters []downloader.Chapter) (string, error) {
	tocDiv := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "div" &&
			hasClass(n, "toc")
	})

	if tocDiv == nil {
		return "", fmt.Errorf("could not find TOC element")
	}

	// Process TOC links with chapter information
	e.processLinks(tocDiv, chapters)

	var buf strings.Builder
	if err := html.Render(&buf, tocDiv); err != nil {
		return "", fmt.Errorf("failed to render TOC: %w", err)
	}

	return buf.String(), nil
}

func cleanHTML(content string) (string, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse content: %w", err)
	}

	var clean func(*html.Node)
	clean = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Remove data- attributes
			newAttrs := []html.Attribute{}
			for _, attr := range n.Attr {
				if !strings.HasPrefix(attr.Key, "data-") {
					newAttrs = append(newAttrs, attr)
				}
			}
			n.Attr = newAttrs
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			clean(c)
		}
	}
	clean(doc)

	var buf strings.Builder
	if err := html.Render(&buf, doc); err != nil {
		return "", fmt.Errorf("failed to render document: %w", err)
	}
	return buf.String(), nil
}
