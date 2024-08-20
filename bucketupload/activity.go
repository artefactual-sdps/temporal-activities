package bucketupload

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.artefactual.dev/tools/temporal"
	"gocloud.dev/blob"
)

const Name = "bucket-upload"

type (
	Params struct {
		Path       string
		Key        string
		BufferSize int
	}
	Result struct {
		Key string
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

	file, err := os.Open(params.Path)
	if err != nil {
		return nil, fmt.Errorf("bucketupload: open file: %w", err)
	}
	defer file.Close()

	key := filepath.Base(params.Path)
	if params.Key != "" {
		key = params.Key
	}

	opts := &blob.WriterOptions{
		ContentType: "application/octet-stream",
		BufferSize:  params.BufferSize,
	}
	if err = a.bucket.Upload(ctx, key, file, opts); err != nil {
		return nil, fmt.Errorf("bucketupload: upload file: %w", err)
	}

	return &Result{Key: key}, nil
}
