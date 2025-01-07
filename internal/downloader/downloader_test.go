package downloader

import (
	"strings"
	"testing"
)

func TestDownloader_FetchChapter(t *testing.T) {
	d := New()
	chapter, err := d.FetchChapter("https://basecamp.com/shapeup/1.1-chapter-02")

	if err != nil {
		t.Fatalf("Failed to fetch chapter: %v", err)
	}

	if chapter.Title == "" {
		t.Error("Chapter title is empty")
	}

	if chapter.Content == "" {
		t.Error("Chapter content is empty")
	}

	if chapter.CSS == "" {
		t.Error("CSS content is empty")
	}
}

func TestDownloader_FetchTOC(t *testing.T) {
	d := New()
	chapters, err := d.FetchTOC()

	if err != nil {
		t.Fatalf("Failed to fetch TOC: %v", err)
	}

	// Verify we have chapters
	if len(chapters) == 0 {
		t.Error("No chapters found in TOC")
	}

	// Test specific known chapters
	expectedChapters := []struct {
		title   string
		urlPart string
	}{
		{"Foreword by Jason Fried", "0.1-foreword"},
		{"Introduction", "0.3-chapter-01"},
		{"Principles of Shaping", "1.1-chapter-02"},
		{"About the Author", "4.6-appendix-07"},
	}

	for _, expected := range expectedChapters {
		found := false
		for _, chapter := range chapters {
			if strings.Contains(chapter.Title, expected.title) &&
				strings.Contains(chapter.URL, expected.urlPart) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find chapter '%s' with URL containing '%s'",
				expected.title, expected.urlPart)
		}
	}
}
func TestDownloader_FetchDifferentChapterTypes(t *testing.T) {
	testCases := []struct {
		name string
		url  string
	}{
		{"Preamble", "https://basecamp.com/shapeup/0.1-foreword"},
		{"MainChapter", "https://basecamp.com/shapeup/1.1-chapter-02"},
		{"Appendix", "https://basecamp.com/shapeup/4.1-appendix-02"},
	}

	d := New()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chapter, err := d.FetchChapter(tc.url)
			if err != nil {
				t.Fatalf("Failed to fetch %s: %v", tc.name, err)
			}

			if chapter.Title == "" {
				t.Error("Chapter title is empty")
			}

			if chapter.Content == "" {
				t.Error("Chapter content is empty")
			}
		})
	}
}
