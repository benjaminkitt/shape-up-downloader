package downloader

import (
	"strings"
	"testing"
)

func TestDownloader_FetchChapter(t *testing.T) {
	d := New()
	testChapter := Chapter{
		URL:    "https://basecamp.com/shapeup/1.1-chapter-02",
		Number: 5, // Example chapter number
	}

	chapter, err := d.FetchChapter(testChapter)

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

	if chapter.Number != 5 {
		t.Errorf("Chapter number not preserved, got %d want %d", chapter.Number, 5)
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

	// Test specific known chapters and their numbers
	expectedChapters := []struct {
		title   string
		urlPart string
		number  int
	}{
		{"Foreword by Jason Fried", "0.1-foreword", 1},
		{"Introduction", "0.3-chapter-01", 3},
		{"Principles of Shaping", "1.1-chapter-02", 4},
		{"About the Author", "4.6-appendix-07", len(chapters)},
	}

	for _, expected := range expectedChapters {
		found := false
		for _, chapter := range chapters {
			if strings.Contains(chapter.Title, expected.title) &&
				strings.Contains(chapter.URL, expected.urlPart) {
				found = true
				if chapter.Number == 0 {
					t.Errorf("Chapter %s has no number assigned", chapter.Title)
				}
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
		name    string
		chapter Chapter
	}{
		{"Preamble", Chapter{URL: "https://basecamp.com/shapeup/0.1-foreword", Number: 1}},
		{"MainChapter", Chapter{URL: "https://basecamp.com/shapeup/1.1-chapter-02", Number: 4}},
		{"Appendix", Chapter{URL: "https://basecamp.com/shapeup/4.1-appendix-02", Number: 20}},
	}

	d := New()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chapter, err := d.FetchChapter(tc.chapter)
			if err != nil {
				t.Fatalf("Failed to fetch %s: %v", tc.name, err)
			}

			if chapter.Title == "" {
				t.Error("Chapter title is empty")
			}

			if chapter.Content == "" {
				t.Error("Chapter content is empty")
			}

			if chapter.Number != tc.chapter.Number {
				t.Errorf("Chapter number not preserved, got %d want %d",
					chapter.Number, tc.chapter.Number)
			}
		})
	}
}

func TestDownloader_FetchChapter_Errors(t *testing.T) {
	d := New()
	testCases := []struct {
		name    string
		chapter Chapter
	}{
		{"Invalid URL", Chapter{URL: "https://invalid.url/chapter"}},
		{"Non-existent chapter", Chapter{URL: "https://basecamp.com/shapeup/not-a-chapter"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := d.FetchChapter(tc.chapter)
			if err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}
