package removefiles

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.artefactual.dev/tools/temporal"
)

const ActivityName = "remove-files-activity"

type Activity struct{}

func NewActivity() *Activity {
	return &Activity{}
}

type ActivityParams struct {
	// RemoveNames is the comma separated list of filenames that should be
	// removed from Path.
	RemoveNames string

	// Path is the directory from which files should be removed.
	Path string
}

type ActivityResult struct {
	// Count is the number of files removed from Path.
	Count int
}

// Execute deletes any file or directory in params.Path (and sub-directories)
// whose name matches one of the values in params.RemoveNames, and returns a
// count of deleted items. A deleted directory is counted as one deleted item
// no matter how many items the directory contains.
//
// RemoveNames should be a comma delimited list of file names, e.g.
// "Thumbs.db,.DS_Store". Any Unicode whitespace characters are trimmed from
// each name in the list, so "Thumbs.db, .DS_Store\n" is equivalent to previous
// example.
//
// If RemoveNames is an empty string, then Execute returns without deleting
// anything.
func (a *Activity) Execute(ctx context.Context, params *ActivityParams) (*ActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(2).Info("Executing RemoveFilesActivity",
		"RemoveNames", params.RemoveNames,
		"Path", params.Path,
	)

	count, err := removeByNames(params.Path, params.RemoveNames)
	if err != nil {
		return nil, fmt.Errorf("RemoveFilesActivity: %v", err)
	}

	return &ActivityResult{Count: count}, nil
}

// removeByNames deletes any file or directory in dir (and sub-directories)
// whose name matches one of the values in remove, then returns a count of
// deleted items.
func removeByNames(dir, names string) (int, error) {
	var count int

	if names == "" {
		return 0, nil
	}

	fi, err := os.Stat(dir)
	if err != nil {
		return 0, fmt.Errorf("remove: %v", err)
	}
	if !fi.IsDir() {
		return 0, fmt.Errorf("remove: path: %q: not a directory", dir)
	}

	files := parseNames(names)
	err = filepath.WalkDir(
		dir,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if slices.Contains(files, d.Name()) {
				err := os.RemoveAll(path)
				if err != nil {
					return fmt.Errorf("remove file: %v", err)
				}
				count++

				if d.IsDir() {
					return fs.SkipDir
				}
			}

			return nil
		},
	)
	if err != nil {
		return count, fmt.Errorf("remove files: %v", err)
	}

	return count, nil
}

// parseNames splits names using a comma delimiter and trims any leading or
// trailing whitespace from each item in the resulting slice.
func parseNames(names string) []string {
	files := strings.Split(names, ",")
	for i, s := range files {
		files[i] = strings.TrimSpace(s)
	}

	return files
}
