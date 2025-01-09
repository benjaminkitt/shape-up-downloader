package converter

import (
	"github.com/benjaminkitt/shape-up-downloader/internal/downloader"
	"golang.org/x/net/html"
)

type Converter interface {
	Convert(chapters []downloader.Chapter, css string) error
}

type baseConverter struct{}

// Shared DOM utilities
func findNode(n *html.Node, criteria func(*html.Node) bool) *html.Node {
	if criteria(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findNode(c, criteria); result != nil {
			return result
		}
	}
	return nil
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func setAttr(n *html.Node, key, value string) {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			n.Attr[i].Val = value
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: value})
}
