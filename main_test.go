package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateFlags(t *testing.T) {
	// Create temp dir for testing output paths
	testDir := t.TempDir()

	tests := []struct {
		name       string
		format     string
		output     string
		wantError  bool
		errorMatch string
	}{
		{
			name:      "valid html format",
			format:    "html",
			output:    filepath.Join(testDir, "test-output"),
			wantError: false,
		},
		{
			name:       "invalid format",
			format:     "pdf",
			output:     filepath.Join(testDir, "test-output"),
			wantError:  true,
			errorMatch: "invalid format",
		},
		{
			name:      "epub format without extension",
			format:    "epub",
			output:    filepath.Join(testDir, "test-output"),
			wantError: false,
		},
		{
			name:      "epub format with extension",
			format:    "epub",
			output:    filepath.Join(testDir, "test-output.epub"),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFlags(tt.format, tt.output)

			if tt.wantError && err == nil {
				t.Error("validateFlags() expected error but got none")
			}

			if !tt.wantError && err != nil {
				t.Errorf("validateFlags() unexpected error: %v", err)
			}

			if tt.errorMatch != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errorMatch) {
					t.Errorf("validateFlags() error = %v, want error containing %v", err, tt.errorMatch)
				}
			}
		})
	}
}

func TestOutputPathHandling(t *testing.T) {
	testDir := t.TempDir()
	existingPath := filepath.Join(testDir, "existing")

	// Create existing directory/file
	err := os.MkdirAll(existingPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name      string
		format    string
		output    string
		wantError bool
	}{
		{
			name:      "existing path",
			format:    "html",
			output:    existingPath,
			wantError: true,
		},
		{
			name:      "new path",
			format:    "html",
			output:    filepath.Join(testDir, "new-path"),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFlags(tt.format, tt.output)
			if (err != nil) != tt.wantError {
				t.Errorf("validateFlags() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
