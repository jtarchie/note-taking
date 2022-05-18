package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type location struct {
	directory string
}

func NewLocation(
	directory string,
	layout func([]byte) string,
) (*location, error) {
	absDirectory, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("could not evaluate directory %q: %w", directory, err)
	}

	r := &location{
		directory: absDirectory,
	}

	return r, nil
}

func (l *location) GetDoc(httpPath string) (*Doc, error) {
	markdownFile, _ := filepath.Abs(filepath.Join(l.directory, httpPath))
	if !strings.HasPrefix(markdownFile, l.directory) {
		return nil, fmt.Errorf("could not evaluate file %s", httpPath)
	}
	if !strings.HasSuffix(markdownFile, ".md") {
		markdownFile += ".md"
	}

	return NewDoc(markdownFile)
}