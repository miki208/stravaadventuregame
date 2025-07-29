package application

import (
	"io/fs"
	"log/slog"
	"path/filepath"
)

func getTemplateFileNames(dir string) []string {
	var templates []string

	slog.Info("Searching for template files in directory", "dir", dir)

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		templates = append(templates, path)

		return nil
	})

	slog.Info("Template files found", "count", len(templates))

	return templates
}
