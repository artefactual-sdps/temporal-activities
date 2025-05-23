package bucketdownload

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.artefactual.dev/tools/temporal"
	"gocloud.dev/blob"
)

const Name = "bucket-download"

type (
	Params struct {
		// Target directory, if missing it will create a new
		// one in the default directory for temporary files.
		DirPath string

		// Target directory permissions, default: 0o700.
		DirPerm os.FileMode

		// Target filename, if missing it will use the Key value.
		FileName string

		// Target file permissions, default: 0o600.
		FilePerm os.FileMode

		// Key from the object storage.
		Key string
	}
	Result struct {
		FilePath string
	}
	Activity struct {
		bucket *blob.Bucket
	}
)

func New(bucket *blob.Bucket) *Activity {
	return &Activity{bucket: bucket}
}

func (a *Activity) Execute(ctx context.Context, params *Params) (*Result, error) {
	h := temporal.StartAutoHeartbeat(ctx)
	defer h.Stop()

	var err error
	dirPath := params.DirPath
	if dirPath != "" {
		dirPerm := fs.FileMode(0o700)
		if params.DirPerm != 0 {
			dirPerm = params.DirPerm
		}
		err = os.MkdirAll(dirPath, dirPerm)
	} else {
		dirPath, err = os.MkdirTemp("", "bucketdownload")
	}
	if err != nil {
		return nil, fmt.Errorf("bucketdownload: create directory: %w", err)
	}

	fileName := params.Key
	if params.FileName != "" {
		fileName = params.FileName
	}
	filePerm := fs.FileMode(0o600)
	if params.FilePerm != 0 {
		filePerm = params.FilePerm
	}

	filePath := filepath.Join(dirPath, fileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePerm) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("bucketdownload: create file: %w", err)
	}
	defer file.Close()

	if err := a.bucket.Download(ctx, params.Key, file, &blob.ReaderOptions{}); err != nil {
		return nil, fmt.Errorf("bucketdownload: download file: %w", err)
	}

	return &Result{FilePath: filePath}, nil
}
