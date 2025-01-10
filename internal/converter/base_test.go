package converter

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// TestFindNode verifies the findNode function can correctly locate HTML nodes
// based on different search criteria
func TestFindNode(t *testing.T) {
	// Create a simple HTML structure for testing
	testHTML := `<div><p class="test">Hello</p><span>World</span></div>`
	doc, _ := html.Parse(strings.NewReader(testHTML))

	// Define test cases with different search criteria
	tests := []struct {
		name     string
		criteria func(*html.Node) bool
		want     string
	}{
		{
			name: "find p element",
			criteria: func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "p"
			},
			want: "p",
		},
		{
			name: "find element with class",
			criteria: func(n *html.Node) bool {
				return n.Type == html.ElementNode && getAttr(n, "class") == "test"
			},
			want: "p",
		},
		{
			name: "find non-existent element",
			criteria: func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "notfound"
			},
			want: "",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findNode(doc, tt.criteria)
			// Verify the found node matches expected data
			if got != nil && got.Data != tt.want {
				t.Errorf("findNode() = %v, want %v", got.Data, tt.want)
			}
			// Verify nil results when expected
			if got == nil && tt.want != "" {
				t.Errorf("findNode() = nil, want %v", tt.want)
			}
		})
	}
}

// TestGetAttr verifies the getAttr function correctly retrieves
// attribute values from HTML nodes
func TestGetAttr(t *testing.T) {
	// Create a test node with multiple attributes
	node := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{Key: "class", Val: "test"},
			{Key: "id", Val: "myid"},
		},
	}

	// Define test cases for attribute retrieval
	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing attribute", "class", "test"},
		{"another existing attribute", "id", "myid"},
		{"non-existent attribute", "style", ""},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAttr(node, tt.key); got != tt.want {
				t.Errorf("getAttr() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSetAttr verifies the setAttr function correctly sets and updates
// attribute values on HTML nodes
func TestSetAttr(t *testing.T) {
	// Define test cases for setting attributes
	tests := []struct {
		name      string
		initial   []html.Attribute
		key       string
		value     string
		wantValue string
	}{
		{
			name:      "set new attribute",
			initial:   []html.Attribute{},
			key:       "class",
			value:     "test",
			wantValue: "test",
		},
		{
			name:      "update existing attribute",
			initial:   []html.Attribute{{Key: "class", Val: "old"}},
			key:       "class",
			value:     "new",
			wantValue: "new",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test node with initial attributes
			node := &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: tt.initial,
			}

			// Set the attribute and verify the result
			setAttr(node, tt.key, tt.value)
			got := getAttr(node, tt.key)
			if got != tt.wantValue {
				t.Errorf("setAttr() resulted in value = %v, want %v", got, tt.wantValue)
			}
		})
	}
}
