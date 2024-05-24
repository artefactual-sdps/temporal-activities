package removefiles

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"go.artefactual.dev/tools/temporal"
)

const ActivityName = "remove-files-activity"

type Activity struct{}

func NewActivity() *Activity {
	return &Activity{}
}

type ActivityParams struct {
	// Path is the directory from which files should be removed.
	Path string

	// RemoveNames is a slice of filename strings that should be
	// removed from the directory.
	RemoveNames []string

	// RemovePatterns is a slice of regular expressions that
	// should be removed from the directory.
	RemovePatterns []*regexp.Regexp
}

type ActivityResult struct {
	// Count is the number of files removed from Path.
	Count int
}

// Execute deletes any file or directory in params.Path (and sub-directories)
// whose name matches one of the values in params.RemoveNames and params.RemovePatterns,
// and returns a count of deleted items. A deleted directory is counted as one deleted
// item no matter how many items the directory contains.
//
// If params.RemoveNames and params.RemovePatterns are empty, then Execute returns without
// deleting anything.
func (a *Activity) Execute(ctx context.Context, params *ActivityParams) (*ActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info(
		"Executing RemoveFilesActivity",
		"Path", params.Path,
		"RemoveNames", params.RemoveNames,
		"RemovePatterns", params.RemovePatterns,
	)

	count, err := remove(params.Path, params.RemoveNames, params.RemovePatterns)
	if err != nil {
		return nil, fmt.Errorf("RemoveFilesActivity: %v", err)
	}

	return &ActivityResult{Count: count}, nil
}

// remove deletes any file or directory in dir (and sub-directories) whose
// name matches one of the values in names and/or patterns, then returns a
// count of deleted items.
func remove(dir string, names []string, patterns []*regexp.Regexp) (int, error) {
	var count int

	if len(names) == 0 && len(patterns) == 0 {
		return 0, nil
	}

	fi, err := os.Stat(dir)
	if err != nil {
		return 0, fmt.Errorf("remove: %v", err)
	}
	if !fi.IsDir() {
		return 0, fmt.Errorf("remove: path: %q: not a directory", dir)
	}

	err = filepath.WalkDir(
		dir,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			matches := false
			if slices.Contains(names, d.Name()) {
				matches = true
			} else {
				for _, re := range patterns {
					matches = re.MatchString(d.Name())
					if matches {
						break
					}
				}
			}

			if matches {
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
