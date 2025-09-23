package archiveextract

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/safeopen"
	"github.com/mholt/archives"
	"go.artefactual.dev/tools/temporal"
)

const Name = "archive-extract"

var (
	ErrNotAFile       = errors.New("Not a file")
	ErrInvalidArchive = errors.New("Invalid archive")
)

type (
	Params struct {
		// SourcePath is the path of the archive to be extracted.
		SourcePath string

		// DestPath is the path where the ExtractPath should be created. If
		// DestPath is not set then ExtractPath will be in the same directory as
		// SourcePath.
		DestPath string
	}
	Result struct {
		// ExtractPath is the path of the extracted archive contents.
		ExtractPath string
	}
	Activity struct {
		cfg Config
	}
)

func New(cfg Config) *Activity {
	return &Activity{cfg: cfg}
}

// Execute extracts the content of SourcePath to a unique extract directory in
// DestPath then returns the ExtractPath.
//
// If SourcePath is a directory an ErrNotAFile error is returned.
// If SourcePath is a file, but not a valid archive, an ErrInvalidArchive error
// is returned.
func (a *Activity) Execute(ctx context.Context, params *Params) (*Result, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing ExtractActivity",
		"SourcePath", params.SourcePath,
		"DestPath", params.DestPath,
	)

	a.cfg.setDefaults()

	dest, err := a.extract(ctx, params.SourcePath, params.DestPath)
	if err != nil {
		switch err {
		case ErrNotAFile:
			logger.V(2).Info("archiveextract: not a file", "SourcePath", params.SourcePath)
			return nil, err
		case ErrInvalidArchive:
			logger.V(2).Info("archiveextract: not a valid archive", "SourcePath", params.SourcePath)
			return nil, err
		default:
			return nil, fmt.Errorf("archiveextract: %v", err)
		}
	}

	dest, err = skipTopLevelDir(dest)
	if err != nil {
		return nil, fmt.Errorf("archiveextract: skipTopLevelDir: %v", err)
	}

	return &Result{ExtractPath: dest}, nil
}

// extract extracts the contents of src archive to dest.
func (a *Activity) extract(ctx context.Context, src, dest string) (string, error) {
	fi, err := os.Stat(src)
	if err != nil {
		return "", err
	}

	if fi.IsDir() {
		return src, ErrNotAFile
	}

	f, err := os.Open(src) // #nosec G304 -- trusted path.
	if err != nil {
		return "", fmt.Errorf("open: %v", err)
	}
	defer f.Close()

	format, r, err := archives.Identify(ctx, src, f)
	if err != nil {
		if errors.Is(err, archives.NoMatch) {
			return "", ErrInvalidArchive
		}
		return "", fmt.Errorf("identify archive: %v", err)
	}

	if ex, ok := format.(archives.Extractor); ok {
		dest, err = extractPath(src, dest)
		if err != nil {
			return "", fmt.Errorf("get extract path: %v", err)
		}

		if err := ex.Extract(ctx, r, a.writeFileHandler(dest)); err != nil {
			// Attempt to remove extract path.
			_ = os.RemoveAll(dest)
			return "", fmt.Errorf("extract: %v", err)
		}
	} else {
		return "", fmt.Errorf("no extractor found: %q", src)
	}

	return dest, nil
}

// writeFileHandler writes the extracted archive file to dest.
func (a *Activity) writeFileHandler(dest string) archives.FileHandler {
	return func(ctx context.Context, f archives.FileInfo) error {
		path := filepath.Join(dest, f.NameInArchive)

		if f.IsDir() {
			// Make any missing dirs in path, then return.
			if err := os.MkdirAll(path, a.cfg.DirMode); err != nil {
				return fmt.Errorf("make directories: %v", err)
			}
			return nil
		} else {
			// Make any missing parent dirs before creating a file.
			if err := os.MkdirAll(filepath.Dir(path), a.cfg.DirMode); err != nil {
				return fmt.Errorf("make parent directories: %v", err)
			}
		}

		df, err := safeopen.CreateBeneath(dest, f.NameInArchive)
		if err != nil {
			return fmt.Errorf("create file: %v", err)
		}
		defer df.Close()

		if err := df.Chmod(a.cfg.FileMode); err != nil {
			return fmt.Errorf("chmod: %v", err)
		}

		r, err := f.Open()
		if err != nil {
			return fmt.Errorf("open source file: %v", err)
		}
		defer r.Close()

		_, err = io.Copy(df, r)
		if err != nil {
			return fmt.Errorf("copy: %v", err)
		}

		return nil
	}
}

// extractPath creates a unique extract directory. If dest is not empty then the
// extract path will be a subdirectory of dest. If dest is empty then the
// extract path will be a subdirectory of src's parent directory.
func extractPath(src, dest string) (string, error) {
	if dest == "" {
		dest = filepath.Dir(src)
	}

	xp, err := os.MkdirTemp(dest, "extract")
	if err != nil {
		return "", fmt.Errorf("make extract dir: %v", err)
	}

	return xp, nil
}

// skipTopLevelDir will return the path to d if base contains exactly one
// sub-directory named d. If base doesn't contain exactly one sub-directory,
// then base is returned.
func skipTopLevelDir(base string) (string, error) {
	items, err := os.ReadDir(base)
	if err != nil {
		return "", fmt.Errorf("read dir: %v", err)
	}

	if len(items) != 1 || !items[0].IsDir() {
		return base, nil
	}

	return filepath.Join(base, items[0].Name()), nil
}
