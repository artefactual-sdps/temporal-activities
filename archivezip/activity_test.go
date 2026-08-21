package archivezip_test

import (
	"archive/zip"
	"fmt"
	"os"
	"testing"
	"time"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/archivezip"
)

func TestActivity(t *testing.T) {
	t.Parallel()

	transferName := "my_transfer"
	contents := tfs.WithDir(transferName,
		tfs.WithDir("subdir",
			tfs.WithFile("abc.txt", "Testing A-B-C"),
		),
		tfs.WithFile("123.txt", "Testing 1-2-3!"),
	)
	td := tfs.NewDir(t, "enduro-zip-test", contents)
	restrictedDir := tfs.NewDir(t, "enduro-zip-restricted", tfs.WithMode(0o555))
	wantModTime := time.Date(2001, time.February, 3, 4, 5, 6, 0, time.UTC)
	assert.NilError(t, os.Chtimes(td.Join(transferName), wantModTime, wantModTime))
	assert.NilError(t, os.Chtimes(td.Join(transferName, "subdir"), wantModTime, wantModTime))
	assert.NilError(t, os.Chtimes(td.Join(transferName, "123.txt"), wantModTime, wantModTime))
	assert.NilError(t, os.Chtimes(td.Join(transferName, "subdir", "abc.txt"), wantModTime, wantModTime))

	type zipFile struct {
		Size    int64
		Method  uint16
		ModTime time.Time
	}
	type test struct {
		name     string
		params   archivezip.Params
		want     map[string]zipFile
		wantPath string
		wantErr  string
	}
	for _, tc := range []test{
		{
			name:   "Zips a directory",
			params: archivezip.Params{SourceDir: td.Join(transferName)},
			want: map[string]zipFile{
				"my_transfer/": {
					Method:  zip.Store,
					ModTime: wantModTime,
				},
				"my_transfer/123.txt": {
					Size:    14,
					Method:  zip.Deflate,
					ModTime: wantModTime,
				},
				"my_transfer/subdir/": {
					Method:  zip.Store,
					ModTime: wantModTime,
				},
				"my_transfer/subdir/abc.txt": {
					Size:    13,
					Method:  zip.Deflate,
					ModTime: wantModTime,
				},
			},
			wantPath: td.Join(transferName + ".zip"),
		},
		{
			name:    "Errors when SourceDir is missing",
			wantErr: "archivezip: missing source dir",
		},
		{
			name: "Errors when dest is not writable",
			params: archivezip.Params{
				SourceDir: td.Join(transferName),
				DestPath:  restrictedDir.Join(transferName + ".zip"),
			},
			wantErr: fmt.Sprintf("archivezip: create destination: open %s: permission denied", restrictedDir.Join(transferName+".zip")),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				archivezip.New().Execute,
				temporalsdk_activity.RegisterOptions{
					Name: archivezip.Name,
				},
			)

			fut, err := env.ExecuteActivity(archivezip.Name, tc.params)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)

			var res archivezip.Result
			_ = fut.Get(&res)
			assert.DeepEqual(t, res, archivezip.Result{Path: tc.wantPath})

			// Confirm the zip has the expected contents.
			rc, err := zip.OpenReader(res.Path)
			assert.NilError(t, err)
			t.Cleanup(func() { rc.Close() })

			files := make(map[string]zipFile, len(rc.File))
			for _, f := range rc.File {
				files[f.Name] = zipFile{
					Size:    f.FileInfo().Size(),
					Method:  f.Method,
					ModTime: f.Modified.UTC(),
				}
			}
			assert.DeepEqual(t, files, tc.want)
		})
	}
}
