package bucketupload_test

import (
	"context"
	"path/filepath"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gocloud.dev/blob"
	"gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/bucketupload"
)

func bucket(t *testing.T) *blob.Bucket {
	t.Helper()

	b := memblob.OpenBucket(nil)
	t.Cleanup(func() { b.Close() })

	return b
}

func TestActivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		bucket      *blob.Bucket
		closeBucket bool
		params      bucketupload.Params
		wantRes     bucketupload.Result
		wantErr     string
	}{
		{
			name:   "Uploads a file",
			bucket: bucket(t),
			params: bucketupload.Params{
				Path: filepath.Join(
					fs.NewDir(t, "bucketupload_test", fs.WithFile("file.txt", "content")).Path(),
					"file.txt",
				),
				Key: "changed.txt",
			},
			wantRes: bucketupload.Result{Key: "changed.txt"},
		},
		{
			name:   "Uploads a file using name as key",
			bucket: bucket(t),
			params: bucketupload.Params{
				Path: filepath.Join(
					fs.NewDir(t, "bucketupload_test", fs.WithFile("file.txt", "content")).Path(),
					"file.txt",
				),
			},
			wantRes: bucketupload.Result{Key: "file.txt"},
		},
		{
			name:   "Fails to upload a missing file",
			bucket: bucket(t),
			params: bucketupload.Params{
				Path: filepath.Join(fs.NewDir(t, "bucketupload_test").Path(), "file.txt"),
			},
			wantErr: "bucketupload: open file:",
		},
		{
			name:        "Fails to upload to a closed bucket",
			bucket:      bucket(t),
			closeBucket: true,
			params: bucketupload.Params{
				Path: filepath.Join(
					fs.NewDir(t, "bucketupload_test", fs.WithFile("file.txt", "content")).Path(),
					"file.txt",
				),
			},
			wantErr: "bucketupload: upload file:",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				bucketupload.New(tt.bucket).Execute,
				temporalsdk_activity.RegisterOptions{Name: bucketupload.Name},
			)

			if tt.closeBucket {
				tt.bucket.Close()
			}

			enc, err := env.ExecuteActivity(bucketupload.Name, tt.params)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var result bucketupload.Result
			_ = enc.Get(&result)
			assert.DeepEqual(t, result, tt.wantRes)

			exists, err := tt.bucket.Exists(context.Background(), tt.wantRes.Key)
			assert.NilError(t, err)
			assert.Assert(t, exists)
		})
	}
}
