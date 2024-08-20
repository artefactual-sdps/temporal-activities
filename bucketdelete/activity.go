package bucketdelete

import (
	"context"
	"fmt"

	"go.artefactual.dev/tools/temporal"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"
)

const Name = "bucket-delete"

type (
	Params struct {
		Key string
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

	err := a.bucket.Delete(ctx, params.Key)
	if err != nil && gcerrors.Code(err) != gcerrors.NotFound {
		return nil, fmt.Errorf("bucketdelete: delete key: %w", err)
	}

	return &Result{}, nil
}
