package epub

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benjaminkitt/shape-up-downloader/internal/downloader"
)

func TestEPUBConverter_Convert(t *testing.T) {
	// Create temp dir for test output
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "shape-up.epub")

	// Create test data
	chapters := []downloader.Chapter{
		{
			Title:   "Test Chapter 1",
			Content: "<h1>Test Chapter 1</h1><p>Test content</p>",
		},
		{
			Title:   "Test Chapter 2",
			Content: "<h1>Test Chapter 2</h1><p>More test content</p>",
		},
	}

	testCSS := "body { font-family: sans-serif; }"

	// Run conversion
	conv := NewEPUBConverter(outputPath)
	err := conv.Convert(chapters, testCSS)

	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Verify EPUB file exists and has content
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Output EPUB file is empty")
	}
}

func TestEPUBConverter_EmptyInput(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "empty.epub")

	conv := NewEPUBConverter(outputPath)
	err := conv.Convert([]downloader.Chapter{}, "")

	if err == nil || !strings.Contains(err.Error(), "no chapters provided") {
		t.Error("Expected error about no chapters provided")
	}
}

func TestEPUBConverter_ComplexContent(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "complex.epub")

	chapters := []downloader.Chapter{
		{
			Title: "Complex Chapter",
			Content: `
                <h1>Complex Content</h1>
                <img src="data:image/png;base64,TEST" />
                <pre><code>Some code block</code></pre>
                <blockquote>A quote</blockquote>
            `,
		},
	}

	conv := NewEPUBConverter(outputPath)
	err := conv.Convert(chapters, "")

	if err != nil {
		t.Fatalf("Failed to convert complex content: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}
