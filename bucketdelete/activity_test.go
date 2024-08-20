package bucketdelete_test

import (
	"context"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gocloud.dev/blob"
	"gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/bucketdelete"
)

func bucket(t *testing.T, key, contents string) *blob.Bucket {
	t.Helper()

	b := memblob.OpenBucket(nil)
	t.Cleanup(func() { b.Close() })

	if key != "" && contents != "" {
		b.WriteAll(context.Background(), key, []byte(contents), nil)
	}

	return b
}

func TestActivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		bucket      *blob.Bucket
		closeBucket bool
		params      bucketdelete.Params
		wantErr     string
	}{
		{
			name:   "Deletes a file",
			bucket: bucket(t, "file.txt", "content"),
			params: bucketdelete.Params{Key: "file.txt"},
		},
		{
			name:   "Doesn't fail if the file doesn't exist",
			bucket: bucket(t, "", ""),
			params: bucketdelete.Params{Key: "file.txt"},
		},
		{
			name:        "Fails to delete a file",
			bucket:      bucket(t, "", ""),
			closeBucket: true,
			params:      bucketdelete.Params{Key: "file.txt"},
			wantErr:     "bucketdelete: delete key:",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				bucketdelete.New(tt.bucket).Execute,
				temporalsdk_activity.RegisterOptions{Name: bucketdelete.Name},
			)

			if tt.closeBucket {
				tt.bucket.Close()
			}

			enc, err := env.ExecuteActivity(bucketdelete.Name, tt.params)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			_ = enc.Get(nil)
			exists, err := tt.bucket.Exists(context.Background(), tt.params.Key)
			assert.NilError(t, err)
			assert.Assert(t, !exists)
		})
	}
}
