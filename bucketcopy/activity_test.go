package bucketcopy_test

import (
	"context"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gocloud.dev/blob"
	"gocloud.dev/blob/memblob"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/bucketcopy"
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
		name    string
		bucket  *blob.Bucket
		params  bucketcopy.Params
		wantErr string
	}{
		{
			name:   "Copies a blob",
			bucket: bucket(t, "source.txt", "content"),
			params: bucketcopy.Params{
				SourceKey: "source.txt",
				DestKey:   "dest.txt",
			},
		},
		{
			name:   "Fails copying a missing blob",
			bucket: bucket(t, "", ""),
			params: bucketcopy.Params{
				SourceKey: "missing.txt",
				DestKey:   "dest.txt",
			},
			wantErr: "bucketcopy: copy blob:",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				bucketcopy.New(tt.bucket).Execute,
				temporalsdk_activity.RegisterOptions{Name: bucketcopy.Name},
			)

			enc, err := env.ExecuteActivity(bucketcopy.Name, tt.params)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			// Verify contents are the same.
			_ = enc.Get(nil)
			sourceData, err := tt.bucket.ReadAll(context.Background(), tt.params.SourceKey)
			assert.NilError(t, err)
			destData, err := tt.bucket.ReadAll(context.Background(), tt.params.DestKey)
			assert.NilError(t, err)
			assert.DeepEqual(t, sourceData, destData)
		})
	}
}
