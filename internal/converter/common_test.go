package converter

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// TestProcessChapterContent verifies the chapter content processing logic
func TestProcessChapterContent(t *testing.T) {
	// Sample chapter HTML with common elements we need to process
	testHTML := `
        <html>
            <body>
                <h1 class="intro__title"><a href="/something">Test Chapter</a></h1>
                <div class="content">
                    <template>Should be removed</template>
                    <nav>Should be removed</nav>
                    <p>Should remain</p>
                    <footer>Should be removed</footer>
                </div>
            </body>
        </html>`

	conv := &baseConverter{}
	result, err := conv.processChapterContent(testHTML)
	if err != nil {
		t.Fatalf("processChapterContent() error = %v", err)
	}

	// Verify elements were properly removed/retained
	if strings.Contains(result, "Should be removed") {
		t.Error("processChapterContent() failed to remove unwanted elements")
	}
	if !strings.Contains(result, "Should remain") {
		t.Error("processChapterContent() incorrectly removed wanted content")
	}
	if !strings.Contains(result, "Test Chapter") {
		t.Error("processChapterContent() failed to retain chapter title")
	}
}

// TestFindAllNodes verifies the node search functionality
func TestFindAllNodes(t *testing.T) {
	testHTML := `
        <div>
            <p class="test">First</p>
            <p class="test">Second</p>
            <span class="test">Third</span>
        </div>`

	doc, _ := html.Parse(strings.NewReader(testHTML))

	tests := []struct {
		name     string
		criteria func(*html.Node) bool
		want     int
	}{
		{
			name: "find all p elements",
			criteria: func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "p"
			},
			want: 2,
		},
		{
			name: "find elements with class 'test'",
			criteria: func(n *html.Node) bool {
				return n.Type == html.ElementNode && hasClass(n, "test")
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := findAllNodes(doc, tt.criteria)
			if len(nodes) != tt.want {
				t.Errorf("findAllNodes() found %v nodes, want %v", len(nodes), tt.want)
			}
		})
	}
}

// TestHasClass verifies class detection in HTML nodes
func TestHasClass(t *testing.T) {
	tests := []struct {
		name      string
		classList string
		checkFor  string
		want      bool
	}{
		{"single class match", "test", "test", true},
		{"one of multiple classes", "test other classes", "other", true},
		{"no matching class", "test other", "missing", false},
		{"empty class list", "", "test", false},
		{"partial class match", "testing", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &html.Node{
				Type: html.ElementNode,
				Attr: []html.Attribute{{Key: "class", Val: tt.classList}},
			}
			if got := hasClass(node, tt.checkFor); got != tt.want {
				t.Errorf("hasClass() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFindBookLinks verifies the book link detection
func TestFindBookLinks(t *testing.T) {
	testHTML := `
        <div>
            <a href="/shapeup/1.1">Valid Link 1</a>
            <a href="https://basecamp.com/shapeup/1.2">Valid Link 2</a>
            <a href="https://other.com/page">Invalid Link</a>
        </div>`

	doc, _ := html.Parse(strings.NewReader(testHTML))
	links := findBookLinks(doc)

	if len(links) != 2 {
		t.Errorf("findBookLinks() found %v links, want 2", len(links))
	}

	// Verify each found link is a valid book link
	for _, link := range links {
		href := getAttr(link, "href")
		if !strings.Contains(href, "/shapeup/") {
			t.Errorf("findBookLinks() found invalid link: %v", href)
		}
	}
}
