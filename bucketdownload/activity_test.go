package bucketdownload_test

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

	"github.com/artefactual-sdps/temporal-activities/bucketdownload"
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

	type setup struct {
		params  bucketdownload.Params
		wantRes bucketdownload.Result
	}

	tests := []struct {
		name    string
		bucket  *blob.Bucket
		setUp   func(t *testing.T) setup
		wantFs  fs.Manifest
		wantErr string
	}{
		{
			name:   "Downloads creating directory and file (default permissions)",
			bucket: bucket(t, "file.txt", "content"),
			setUp: func(t *testing.T) setup {
				dir := fs.NewDir(t, "bucketdownload_test").Path()
				return setup{
					params: bucketdownload.Params{
						DirPath: filepath.Join(dir, "new"),
						Key:     "file.txt",
					},
					wantRes: bucketdownload.Result{
						FilePath: filepath.Join(dir, "new", "file.txt"),
					},
				}
			},
			wantFs: fs.Expected(
				t,
				fs.WithMode(0o700),
				fs.WithFile("file.txt", "content", fs.WithMode(0o600)),
			),
		},
		{
			name:   "Downloads creating directory and file (given permissions)",
			bucket: bucket(t, "file.txt", "content"),
			setUp: func(t *testing.T) setup {
				dir := fs.NewDir(t, "bucketdownload_test").Path()
				return setup{
					params: bucketdownload.Params{
						DirPath:  filepath.Join(dir, "new"),
						DirPerm:  0o755,
						Key:      "file.txt",
						FileName: "changed.txt",
						FilePerm: 0o644,
					},
					wantRes: bucketdownload.Result{
						FilePath: filepath.Join(dir, "new", "changed.txt"),
					},
				}
			},
			wantFs: fs.Expected(
				t,
				fs.WithMode(0o755),
				fs.WithFile("changed.txt", "content", fs.WithMode(0o644)),
			),
		},
		{
			name:   "Downloads creating file",
			bucket: bucket(t, "file.txt", "content"),
			setUp: func(t *testing.T) setup {
				dir := fs.NewDir(t, "bucketdownload_test", fs.WithDir("empty", fs.WithMode(0o755))).Path()
				return setup{
					params: bucketdownload.Params{
						DirPath: filepath.Join(dir, "empty"),
						Key:     "file.txt",
					},
					wantRes: bucketdownload.Result{
						FilePath: filepath.Join(dir, "empty", "file.txt"),
					},
				}
			},
			wantFs: fs.Expected(
				t,
				fs.WithMode(0o755),
				fs.WithFile("file.txt", "content", fs.WithMode(0o600)),
			),
		},
		{
			name:   "Downloads overwriting file",
			bucket: bucket(t, "file.txt", "content"),
			setUp: func(t *testing.T) setup {
				dir := fs.NewDir(
					t,
					"bucketdownload_test",
					fs.WithDir(
						"withfile",
						fs.WithMode(0o755),
						fs.WithFile("file.txt", "old content", fs.WithMode(0o644)),
					),
				).Path()
				return setup{
					params: bucketdownload.Params{
						DirPath: filepath.Join(dir, "withfile"),
						Key:     "file.txt",
					},
					wantRes: bucketdownload.Result{
						FilePath: filepath.Join(dir, "withfile", "file.txt"),
					},
				}
			},
			wantFs: fs.Expected(
				t,
				fs.WithMode(0o755),
				fs.WithFile("file.txt", "content", fs.WithMode(0o644)),
			),
		},
		{
			name:   "Fails downloading a missing file",
			bucket: bucket(t, "", ""),
			setUp: func(t *testing.T) setup {
				dir := fs.NewDir(t, "bucketdownload_test").Path()
				return setup{
					params: bucketdownload.Params{
						DirPath: filepath.Join(dir, "new"),
						Key:     "file.txt",
					},
				}
			},
			wantErr: "bucketdownload: download file:",
		},
		{
			name:   "Fails creating the directory",
			bucket: bucket(t, "", ""),
			setUp: func(t *testing.T) setup {
				dir := fs.NewDir(t, "bucketdownload_test", fs.WithMode(0o000)).Path()
				return setup{
					params: bucketdownload.Params{
						DirPath: filepath.Join(dir, "new"),
						Key:     "file.txt",
					},
				}
			},
			wantErr: "bucketdownload: create directory:",
		},
		{
			name:   "Fails creating the file",
			bucket: bucket(t, "", ""),
			setUp: func(t *testing.T) setup {
				dir := fs.NewDir(t, "bucketdownload_test", fs.WithDir("empty", fs.WithMode(0o000))).Path()
				return setup{
					params: bucketdownload.Params{
						DirPath: filepath.Join(dir, "empty"),
						Key:     "file.txt",
					},
				}
			},
			wantErr: "bucketdownload: create file:",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				bucketdownload.New(tt.bucket).Execute,
				temporalsdk_activity.RegisterOptions{Name: bucketdownload.Name},
			)

			setup := tt.setUp(t)
			enc, err := env.ExecuteActivity(bucketdownload.Name, setup.params)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var result bucketdownload.Result
			_ = enc.Get(&result)
			assert.DeepEqual(t, result, setup.wantRes)
			assert.Assert(t, fs.Equal(setup.params.DirPath, tt.wantFs))
		})
	}
}
