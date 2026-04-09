package archivezip

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"go.artefactual.dev/tools/temporal"
)

const Name = "archive-zip"

type (
	Params struct {
		SourceDir string
		DestPath  string
	}
	Result struct {
		Path string
	}
	Activity struct{}
)

func New() *Activity {
	return &Activity{}
}

// Execute creates a Zip archive at params.DestPath from the contents of
// params.SourceDir. If params.DestPath is not specified then params.SourceDir
// + ".zip" will be used.
func (a *Activity) Execute(ctx context.Context, params *Params) (*Result, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing ZipActivity",
		"SourceDir", params.SourceDir,
		"DestPath", params.DestPath,
	)

	if params.SourceDir == "" {
		return &Result{}, fmt.Errorf("archivezip: missing source dir")
	}

	dest := params.DestPath
	if params.DestPath == "" {
		dest = params.SourceDir + ".zip"
		logger.V(1).Info("archivezip: dest changed", "dest", dest)
	}

	w, err := os.Create(dest) // #nosec G304 -- trusted path
	if err != nil {
		return &Result{}, fmt.Errorf("archivezip: create destination: %v", err)
	}
	defer w.Close()

	z := zip.NewWriter(w)
	defer z.Close()

	root, err := os.OpenRoot(params.SourceDir)
	if err != nil {
		return &Result{}, fmt.Errorf("archivezip: open root: %v", err)
	}
	defer root.Close()

	err = fs.WalkDir(root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Include the SourceDir name in the zip paths.
		zipPath := filepath.Join(filepath.Base(params.SourceDir), path)

		f, err := z.Create(zipPath)
		if err != nil {
			return err
		}

		r, err := root.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()

		if _, err := io.Copy(f, r); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return &Result{}, fmt.Errorf("archivezip: add files: %v", err)
	}

	return &Result{Path: dest}, nil
}
