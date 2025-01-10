package converter

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-shiori/go-epub"
	"golang.org/x/net/html"
)

// TestEPUBConverter_ProcessImages verifies image processing in EPUB content
func TestEPUBConverter_ProcessImages(t *testing.T) {
	// Set up test server for image downloads
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, err := w.Write([]byte("fake-image-data"))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// Create test HTML with various image scenarios
	testHTML := `
        <div>
            <img src="` + server.URL + `/test.jpg" alt="External Image">
            <img src="data:image/jpeg;base64,` + base64.StdEncoding.EncodeToString([]byte("test-data")) + `" alt="Data URL Image">
        </div>`

	doc, _ := html.Parse(strings.NewReader(testHTML))

	// Create EPUB book for testing
	book, err := epub.NewEpub("Test Book")
	if err != nil {
		t.Fatalf("Failed to create EPUB: %v", err)
	}

	// Process images in the document
	conv := NewEPUBConverter("test.epub")
	err = conv.processImages(doc, book)
	if err != nil {
		t.Fatalf("processImages() error = %v", err)
	}

	// Verify image processing results
	images := findAllNodes(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "img"
	})

	for _, img := range images {
		src := getAttr(img, "src")
		if !strings.HasPrefix(src, "../images/") {
			t.Errorf("Image src not properly converted: %s", src)
		}
	}
}

// TestEPUBConverter_DownloadAndAddImage verifies image download and EPUB embedding
func TestEPUBConverter_DownloadAndAddImage(t *testing.T) {
	// Set up test server with image response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, err := w.Write([]byte("fake-image-data"))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// Create EPUB book for testing
	book, err := epub.NewEpub("Test Book")
	if err != nil {
		t.Fatalf("Failed to create EPUB: %v", err)
	}

	conv := NewEPUBConverter("test.epub")
	imagePath, err := conv.downloadAndAddImage(server.URL+"/test.jpg", book)

	if err != nil {
		t.Fatalf("downloadAndAddImage() error = %v", err)
	}

	// Verify image path format
	if !strings.HasPrefix(imagePath, "../images/") || !strings.HasSuffix(imagePath, ".jpg") {
		t.Errorf("Invalid image path format: %s", imagePath)
	}
}
