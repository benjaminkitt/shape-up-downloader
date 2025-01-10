package converter

import (
	"archive/zip"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benjaminkitt/shape-up-downloader/internal/downloader"
	"golang.org/x/net/html"
)

// TestEPUBConverter_Convert verifies the complete EPUB generation process
// including file structure, chapter organization, and required EPUB components
func TestEPUBConverter_Convert(t *testing.T) {
	// Create temporary directory for test output
	testFile := filepath.Join(t.TempDir(), "test.epub")

	// Set up test chapters including TOC and content
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

	// Initialize converter and perform conversion
	conv := NewEPUBConverter(testFile)
	err := conv.Convert(chapters, "body { color: black; }")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Open and verify EPUB file structure
	reader, err := zip.OpenReader(testFile)
	if err != nil {
		t.Fatalf("Failed to open EPUB file: %v", err)
	}
	defer reader.Close()

	// Define required files according to EPUB 3.0 specification
	requiredFiles := []string{
		"META-INF/container.xml",       // EPUB container descriptor
		"EPUB/toc.ncx",                 // Table of Contents
		"EPUB/nav.xhtml",               // Navigation document
		"EPUB/package.opf",             // Package information
		"EPUB/xhtml/section0001.xhtml", // Print Table of Contents
		"EPUB/xhtml/section0002.xhtml", // Chapter 1
	}

	// Track found files for verification
	foundFiles := make(map[string]bool)
	for _, f := range reader.File {
		foundFiles[f.Name] = true
	}

	// Verify all required files are present
	for _, required := range requiredFiles {
		if !foundFiles[required] {
			t.Errorf("Missing required EPUB file: %s", required)
		}
	}
}

// TestEPUBConverter_CreateTitlePage verifies the generation of the EPUB title page
// including all required metadata and formatting
func TestEPUBConverter_CreateTitlePage(t *testing.T) {
	conv := NewEPUBConverter("test.epub")
	titlePage, err := conv.createTitlePage()
	if err != nil {
		t.Fatalf("createTitlePage() error = %v", err)
	}

	// Define required elements for the title page
	expectedElements := []string{
		"Shape Up",
		"Stop Running in Circles",
		"Ship Work that Matters",
		"Ryan Singer",
	}

	// Verify all required elements are present
	for _, expected := range expectedElements {
		if !strings.Contains(titlePage, expected) {
			t.Errorf("createTitlePage() missing expected content: %s", expected)
		}
	}
}

// TestEPUBConverter_ProcessLinks verifies the conversion of internal links
// to the correct EPUB chapter reference format
func TestEPUBConverter_ProcessLinks(t *testing.T) {
	// Set up test HTML with internal links
	testHTML := `
        <div>
            <a href="/shapeup/1.1">Chapter 1</a>
            <a href="/shapeup/1.2#section">Chapter 1.2</a>
        </div>`

	doc, _ := html.Parse(strings.NewReader(testHTML))

	// Define test chapters for link resolution
	chapters := []downloader.Chapter{
		{URL: "https://basecamp.com/shapeup/1.1", Number: 1},
		{URL: "https://basecamp.com/shapeup/1.2", Number: 2},
	}

	// Process links in the document
	conv := NewEPUBConverter("test.epub")
	conv.processLinks(doc, chapters)

	// Find all processed links
	links := findAllNodes(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a"
	})

	// Define expected EPUB link formats
	expectedFormats := []string{
		"section0003.xhtml",
		"section0004.xhtml#section",
	}

	// Verify link transformations
	for i, link := range links {
		href := getAttr(link, "href")
		if href != expectedFormats[i] {
			t.Errorf("processLinks() got href = %s, want %s", href, expectedFormats[i])
		}
	}
}
