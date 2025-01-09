package bagcreate

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	gobagit "github.com/nyudlts/go-bagit"
	cp "github.com/otiai10/copy"
	"go.artefactual.dev/tools/fsutil"
	"go.artefactual.dev/tools/temporal"
)

const (
	Name                 = "bag-create"
	dirMode  fs.FileMode = 0o700
	fileMode fs.FileMode = 0o600
)

type (
	Params struct {
		// SourcePath is the path of the files to be added to the created Bag.
		SourcePath string

		// BagPath is the path where the Bag should be created. If BagPath is empty,
		// then the Bag will be created at SourcePath, replacing the original
		// directory contents.
		BagPath string
	}
	Result struct {
		// BagPath of the path to the created Bag.
		BagPath string
	}
	Activity struct {
		cfg *Config
	}
)

func New(cfg Config) *Activity {
	cfg.setDefaults()

	return &Activity{cfg: &cfg}
}

// Execute creates a BagIt Bag containing the files at SourcePath.
//
// If BagPath is set in the parameters, then the Bag will be created at BagPath.
// If BagPath is empty, then the Bag will be created at SourcePath, replacing
// the original directory contents. In either case the path of the Bag is
// returned.
func (a *Activity) Execute(ctx context.Context, params *Params) (*Result, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing bag-create activity", "SourcePath", params.SourcePath)

	// Check if directory is already a Bag
	if _, err := os.Stat(filepath.Join(params.SourcePath, "bagit.txt")); err == nil {
		return &Result{BagPath: params.SourcePath}, nil
	}

	dest, err := a.create(params.SourcePath, params.BagPath)
	if err != nil {
		return nil, fmt.Errorf("bagcreate: %v", err)
	}

	return &Result{BagPath: dest}, nil
}

// create creates a BagIt Bag at dest from the files at src. If dest is empty,
// the BagIt Bag is created in-place at src.
func (a *Activity) create(src, dest string) (string, error) {
	if dest == "" {
		dest = src
	} else {
		if err := cp.Copy(src, dest); err != nil {
			return "", fmt.Errorf("copy source dir to bag path: %v", err)
		}
	}

	// CreateBag currently only runs a single process to create a bag. If this
	// changes in the future we should add a config value to set numProcesses.
	_, err := gobagit.CreateBag(dest, a.cfg.ChecksumAlgorithm, 1)
	if err != nil {
		return "", fmt.Errorf("create bag: %v", err)
	}

	if err := fsutil.SetFileModes(dest, dirMode, fileMode); err != nil {
		return "", fmt.Errorf("set file modes: %v", err)
	}

	return dest, nil
}
