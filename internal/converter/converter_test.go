package converter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benjaminkitt/shape-up-downloader/internal/downloader"
)

func TestHTMLConverter_Convert(t *testing.T) {
	// Create temp dir for test output
	tmpDir := t.TempDir()

	// Create test data with TOC content
	chapters := []downloader.Chapter{
		{
			Title: "Table of Contents",
			URL:   "toc",
			Content: `<div class="toc">
				<div class="toc__part">
					<h2 class="toc__part-title">Test Part</h2>
					<ul class="toc__chapters">
						<li><a href="/shapeup/test-chapter-1">Chapter 1</a></li>
						<li><a href="/shapeup/test-chapter-2">Chapter 2</a></li>
					</ul>
				</div>
			</div>`,
		},
		{
			Title:   "Test Chapter 1",
			URL:     "test-chapter-1",
			Content: "<h1>Test Chapter 1</h1><p>Test content</p>",
		},
		{
			Title:   "Test Chapter 2",
			URL:     "test-chapter-2",
			Content: "<h1>Test Chapter 2</h1><p>More test content</p>",
		},
	}

	testCSS := "body { font-family: sans-serif; }"

	conv := NewHTMLConverter(tmpDir)
	err := conv.Convert(chapters, testCSS)

	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Verify output file exists
	outputPath := filepath.Join(tmpDir, "index.html")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify content
	htmlContent := string(content)

	expectations := []string{
		testCSS,          // CSS included
		`class="toc"`,    // TOC structure preserved
		"Test Chapter 1", // Chapter titles present
		"Test Chapter 2",
		"<p>Test content</p>", // Chapter content present
		"<p>More test content</p>",
		`href="#test-chapter-1"`, // Updated relative links
	}

	for _, expected := range expectations {
		if !strings.Contains(htmlContent, expected) {
			t.Errorf("Output HTML missing expected content: %s", expected)
		}
	}
}

func TestHTMLConverter_InvalidOutputDir(t *testing.T) {
	// Test conversion with invalid/unwriteable output directory
	conv := NewHTMLConverter("/nonexistent/directory")
	err := conv.Convert([]downloader.Chapter{}, "")
	if err == nil {
		t.Error("Expected error for invalid output directory")
	}
}

func TestHTMLConverter_ComplexContent(t *testing.T) {
	tmpDir := t.TempDir()

	chapters := []downloader.Chapter{
		{
			Title: "Table of Contents",
			URL:   "toc",
			Content: `<div class="toc">
				<div class="toc__part">
					<h2 class="toc__part-title">Complex Part</h2>
					<ul class="toc__chapters">
						<li><a href="/shapeup/complex">Complex Chapter</a></li>
					</ul>
				</div>
			</div>`,
		},
		{
			Title: "Complex Chapter",
			URL:   "complex",
			Content: `
				<h1>Complex Content</h1>
				<img src="data:image/png;base64,TEST" />
				<pre><code>Some code block</code></pre>
				<blockquote>A quote</blockquote>
			`,
		},
	}

	conv := NewHTMLConverter(tmpDir)
	err := conv.Convert(chapters, "")
	if err != nil {
		t.Fatalf("Failed to convert complex content: %v", err)
	}

	// Verify complex elements preserved
	content, _ := os.ReadFile(filepath.Join(tmpDir, "index.html"))
	htmlContent := string(content)

	elements := []string{
		"data:image/png;base64,TEST",
		"<pre><code>",
		"<blockquote>",
		`class="toc"`,
		`href="#complex"`,
	}

	for _, elem := range elements {
		if !strings.Contains(htmlContent, elem) {
			t.Errorf("Output HTML missing complex element: %s", elem)
		}
	}
}

func TestHTMLConverter_EmptyInput(t *testing.T) {
	tmpDir := t.TempDir()
	conv := NewHTMLConverter(tmpDir)

	// Test with empty chapter list
	err := conv.Convert([]downloader.Chapter{}, "")
	if err == nil {
		t.Error("Expected error for empty chapter list")
	}
}

func TestHTMLConverter_CreateOutputDirectory(t *testing.T) {
	// Create a temporary parent directory
	tmpParent := t.TempDir()

	// Specify a subdirectory that doesn't exist yet
	outputDir := filepath.Join(tmpParent, "subdir", "output")

	chapters := []downloader.Chapter{
		{
			Title: "Table of Contents",
			URL:   "toc",
			Content: `<div class="toc">
				<div class="toc__part">
					<h2 class="toc__part-title">Test Part</h2>
					<ul class="toc__chapters">
						<li><a href="/shapeup/test-chapter">Test Chapter</a></li>
					</ul>
				</div>
			</div>`,
		},
		{
			Title:   "Test Chapter",
			URL:     "test-chapter",
			Content: "<h1>Test Content</h1>",
		},
	}

	// Convert should create the directory structure
	conv := NewHTMLConverter(outputDir)
	err := conv.Convert(chapters, "test-css")

	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("Output directory was not created")
	}

	// Verify file exists in the directory
	outputFile := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}
