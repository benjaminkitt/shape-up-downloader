package converter

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// Implementation of baseConverter methods
func (b *baseConverter) processChapterContent(content string) (string, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse chapter content: %w", err)
	}

	// Find title
	titleNode := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "h1" &&
			hasClass(n, "intro__title")
	})

	// Find content
	contentDiv := findNode(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode &&
			n.Data == "div" &&
			hasClass(n, "content")
	})

	if contentDiv != nil {
		// Remove templates
		b.removeElements(contentDiv, "template")

		// Remove nav
		b.removeElements(contentDiv, "nav")

		// Remove footer
		b.removeElements(contentDiv, "footer")

		// Handle title insertion
		if titleNode != nil {
			// Update the title link to point to TOC
			titleLink := findNode(titleNode, func(n *html.Node) bool {
				return n.Type == html.ElementNode && n.Data == "a"
			})
			if titleLink != nil {
				setAttr(titleLink, "href", "#toc")
			}

			// Move title to content
			titleNode.Parent.RemoveChild(titleNode)
			contentDiv.InsertBefore(titleNode, contentDiv.FirstChild)
		}
	}

	var buf strings.Builder
	if err := html.Render(&buf, contentDiv); err != nil {
		return "", fmt.Errorf("failed to render content: %w", err)
	}

	return buf.String(), nil
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

// Helper methods for baseConverter
func (b *baseConverter) removeElements(node *html.Node, tagName string) {
	elements := findAllNodes(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tagName
	})
	for _, element := range elements {
		element.Parent.RemoveChild(element)
	}
}

func findBookLinks(node *html.Node) []*html.Node {
	return findAllNodes(node, func(n *html.Node) bool {
		if n.Type != html.ElementNode || n.Data != "a" {
			return false
		}
		href := getAttr(n, "href")
		return strings.HasPrefix(href, "/shapeup/") ||
			strings.HasPrefix(href, "https://basecamp.com/shapeup/")
	})
}
