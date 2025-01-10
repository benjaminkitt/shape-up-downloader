package converter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benjaminkitt/shape-up-downloader/internal/downloader"
	"golang.org/x/net/html"
)

// TestHTMLConverter_Convert tests the full HTML conversion process
func TestHTMLConverter_Convert(t *testing.T) {
	// Set up test directory
	testDir := t.TempDir()

	// Create test chapters
	chapters := []downloader.Chapter{
		{
			Title: "Table of Contents",
			Content: `<div class="content">
										<div class="toc">
												<a href="/shapeup/1.1">Chapter 1</a>
										</div>
								 </div>`,
			URL:    "https://basecamp.com/shapeup/toc",
			Number: 0,
		},
		{
			Title:   "Chapter 1",
			Content: "<div class='content'><h1>Test Content</h1><p>Test paragraph</p></div>",
			URL:     "https://basecamp.com/shapeup/1.1",
			Number:  1,
		},
	}

	// Create converter instance
	conv := NewHTMLConverter(testDir)

	// Test conversion
	err := conv.Convert(chapters, "body { color: black; }")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Verify output file exists
	outputPath := filepath.Join(testDir, "index.html")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify content includes expected elements
	htmlContent := string(content)
	expectedElements := []string{
		"<!DOCTYPE html>",
		"<style>body { color: black; }</style>",
		"Test Content",
		"Test paragraph",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(htmlContent, expected) {
			t.Errorf("Convert() output missing expected content: %s", expected)
		}
	}
}

// TestHTMLConverter_OrganizeParts tests the chapter organization logic
func TestHTMLConverter_OrganizeParts(t *testing.T) {
	conv := NewHTMLConverter("test")

	tests := []struct {
		name         string
		inputLength  int
		wantSections int
	}{
		{"small book", 3, 1},
		{"full book", 20, 5}, // Should split into intro, parts 1-3, and appendices
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chapters := make([]downloader.Chapter, tt.inputLength)
			for i := range chapters {
				chapters[i] = downloader.Chapter{
					Title:  "Chapter",
					Number: i + 1,
				}
			}

			parts := conv.organizeParts(chapters)
			if len(parts) != tt.wantSections {
				t.Errorf("organizeParts() got %v sections, want %v", len(parts), tt.wantSections)
			}
		})
	}
}

// TestHTMLConverter_ProcessLinks tests the link processing functionality
func TestHTMLConverter_ProcessLinks(t *testing.T) {
	testHTML := `
        <div>
            <a href="https://basecamp.com/shapeup/1.1">Link 1</a>
            <a href="/shapeup/1.2#section">Link 2</a>
            <a href="https://other.com">External Link</a>
        </div>`

	doc, _ := html.Parse(strings.NewReader(testHTML))

	conv := NewHTMLConverter("test")
	conv.processLinks(doc)

	links := findAllNodes(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a"
	})

	expectedHrefs := map[string]bool{
		"#1.1":              false,
		"#section":          false,
		"https://other.com": false,
	}

	for _, link := range links {
		href := getAttr(link, "href")
		if _, exists := expectedHrefs[href]; exists {
			expectedHrefs[href] = true
		}
	}

	for href, found := range expectedHrefs {
		if !found {
			t.Errorf("processLinks() failed to process href correctly: %s", href)
		}
	}
}
