package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/benjaminkitt/shape-up-downloader/internal/converter"
	"github.com/benjaminkitt/shape-up-downloader/internal/downloader"
	"github.com/spf13/cobra"
)

func validateFlags(format string, output string) error {
	// Validate format
	format = strings.ToLower(format)
	if format != "html" && format != "epub" {
		return fmt.Errorf("invalid format: %s (must be 'html' or 'epub')", format)
	}

	// Validate output path
	if format == "epub" {
		if !strings.HasSuffix(output, ".epub") {
			output = output + ".epub"
		}
	}

	// Check if output directory/file exists
	if _, err := os.Stat(output); err == nil {
		return fmt.Errorf("output path already exists: %s", output)
	}

	return nil
}

func main() {
	var outputFormat string
	var outputDir string

	rootCmd := &cobra.Command{
		Use:   "shape-up-downloader",
		Short: "Download the Shape Up book from Basecamp",
		Long: `A CLI tool to download the Shape Up book by Ryan Singer, 
               published by Basecamp, and save it as HTML or EPUB`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateFlags(outputFormat, outputDir); err != nil {
				return err
			}

			// Initialize downloader
			dl := downloader.New()

			// Fetch table of contents
			chapters, err := dl.FetchTOC()
			if err != nil {
				return fmt.Errorf("failed to fetch table of contents: %w", err)
			}

			// Download each chapter
			for i, chapter := range chapters {
				fmt.Printf("Downloading chapter %d/%d: %s\n", i+1, len(chapters), chapter.Title)
				ch, err := dl.FetchChapter(chapter)
				if err != nil {
					return fmt.Errorf("failed to fetch chapter %s: %w", chapter.Title, err)
				}
				chapters[i] = *ch
			}

			// Convert to requested format
			switch outputFormat {
			case "html":
				conv := converter.NewHTMLConverter(outputDir)
				if err := conv.Convert(chapters, chapters[0].CSS); err != nil {
					return fmt.Errorf("failed to convert to HTML: %w", err)
				}
			case "epub":
				conv := converter.NewEPUBConverter(outputDir)
				if err := conv.Convert(chapters, chapters[0].CSS); err != nil {
					return fmt.Errorf("failed to convert to EPUB: %w", err)
				}
			}

			fmt.Printf("Successfully downloaded Shape Up book to %s\n", outputDir)
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "html", "Output format (html or epub)")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "shape-up-book", "Output directory for HTML or filename for EPUB")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
