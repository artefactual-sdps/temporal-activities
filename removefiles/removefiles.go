package removefiles

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Remove deletes any files or directories in dir (and sub-directories) whose
// names match one of the values in remove, then returns a count of the items
// deleted.
//
// The remove parameter should be a comma delimited list of file names, e.g.
// "Thumbs.db,.DS_Store".  If remove is empty, then Remove returns without
// deleting anything.
func Remove(dir, remove string) (int, error) {
	var count int

	if remove == "" {
		return 0, nil
	}

	fi, err := os.Stat(dir)
	if err != nil {
		return 0, fmt.Errorf("remove: %v", err)
	}
	if !fi.IsDir() {
		return 0, fmt.Errorf("remove: path: %q: not a directory", dir)
	}

	files := parseFilenames(remove)
	err = filepath.WalkDir(
		dir,
		func(path string, d fs.DirEntry, err error) error {
			if slices.Contains(files, d.Name()) {
				err := os.RemoveAll(path)
				if err != nil {
					return fmt.Errorf("remove file: %v", err)
				}
				count++
			}

			return nil
		},
	)
	if err != nil {
		return count, fmt.Errorf("remove files: %v", err)
	}

	return count, nil
}

func parseFilenames(names string) []string {
	files := strings.Split(names, ",")
	for i, s := range files {
		files[i] = strings.TrimSpace(s)
	}

	return files
}
