package handler

import (
	"fmt"
	"os/exec"
	"strings"
)

// extractPDFText uses pdftotext (from poppler-utils) to extract text from a PDF.
// Install on Mac: brew install poppler
// Install on Ubuntu: sudo apt-get install poppler-utils
func extractPDFText(filePath string) (string, error) {
	// First try pdftotext (most reliable)
	cmd := exec.Command("pdftotext", "-layout", filePath, "-")
	output, err := cmd.Output()
	if err == nil {
		text := strings.TrimSpace(string(output))
		if text != "" {
			return text, nil
		}
	}

	return "", fmt.Errorf("pdftotext not available or failed: %w (install with: brew install poppler)", err)
}