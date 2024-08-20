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
		DirPath  string
		DirPerm  os.FileMode
		FileName string
		FilePerm os.FileMode
		Key      string
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

	dirPerm := fs.FileMode(0o700)
	if params.DirPerm != 0 {
		dirPerm = params.DirPerm
	}

	if err := os.MkdirAll(params.DirPath, dirPerm); err != nil {
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

	filePath := filepath.Join(params.DirPath, fileName)
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
