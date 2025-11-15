package docx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// ExtractTextFromDocx saves the uploaded DOCX to a temp file,
// calls pandoc, and returns plain UTF-8 text.
func ExtractTextFromDocx(r io.Reader) (string, error) {
	path, err := saveTempDocx(r)
	if err != nil {
		return "", fmt.Errorf("saveTempDocx: %w", err)
	}
	defer os.Remove(path)

	text, err := extractWithPandoc(path)
	if err != nil {
		return "", fmt.Errorf("pandoc: %w", err)
	}

	return text, nil
}

func saveTempDocx(r io.Reader) (string, error) {
	f, err := os.CreateTemp("", "tce-*.docx")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func extractWithPandoc(path string) (string, error) {
	cmd := exec.Command("pandoc", "-f", "docx", "-t", "plain", path)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%v (stderr: %s)", err, stderr.String())
	}

	return out.String(), nil
}
