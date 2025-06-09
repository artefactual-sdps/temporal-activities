package bucketcopy

import (
	"context"
	"fmt"

	"go.artefactual.dev/tools/temporal"
	"gocloud.dev/blob"
)

const Name = "bucket-copy"

type (
	Params struct {
		// Source object key.
		SourceKey string

		// Destination object key.
		DestKey string
	}
	Result   struct{}
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

	if err := a.bucket.Copy(ctx, params.DestKey, params.SourceKey, nil); err != nil {
		return nil, fmt.Errorf("bucketcopy: copy blob: %w", err)
	}

	return &Result{}, nil
}
