package removefiles

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/go-logr/logr"
)

const RemoveFilesActivityName = "remove-files"

type RemoveFilesActivity struct {
	logger logr.Logger
}

func NewRemoveFilesActivity() *RemoveFilesActivity {
	return &RemoveFilesActivity{}
}

type RemoveFilesActivityParams struct {
	// Filelist is the list of filenames that should be removed from Path.
	Filelist []string

	// Path is the directory from which files should be removed.
	Path string
}

type RemoveFilesActivityResult struct {
	// Count is the number of files removed from Path.
	Count int
}

func (a *RemoveFilesActivity) Execute(
	ctx context.Context,
	params *RemoveFilesActivityParams,
) (*RemoveFilesActivityResult, error) {
	a.logger.V(2).Info("Executing RemoveFilesActivity",
		"Filelist", params.Filelist,
		"Path", params.Path,
	)

	count, err := removeFiles(params.Path, params.Filelist)
	if err != nil {
		return nil, fmt.Errorf("RemoveFilesActivity: %v", err)
	}

	return &RemoveFilesActivityResult{Count: count}, nil
}

func removeFiles(dir string, files []string) (int, error) {
	var count int

	fi, err := os.Stat(dir)
	if err != nil {
		return 0, fmt.Errorf("file info: %v", err)
	}
	if !fi.IsDir() {
		return 0, fmt.Errorf("%q is not a directory", dir)
	}

	err = filepath.WalkDir(
		dir,
		func(path string, d fs.DirEntry, err error) error {
			if slices.Contains(files, d.Name()) {
				err := os.Remove(path)
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
