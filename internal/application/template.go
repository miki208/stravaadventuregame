package application

import (
	"io/fs"
	"path/filepath"
)

func getTemplateFileNames(dir string) []string {
	var templates []string

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		templates = append(templates, path)

		return nil
	})

	return templates
}
