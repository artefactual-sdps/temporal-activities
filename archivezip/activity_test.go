package archivezip_test

import (
	"archive/zip"
	"fmt"
	"testing"

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

	type test struct {
		name    string
		params  archivezip.Params
		want    map[string]int64
		wantErr string
	}
	for _, tc := range []test{
		{
			name:   "Zips a directory",
			params: archivezip.Params{SourceDir: td.Join(transferName)},
			want: map[string]int64{
				"my_transfer/123.txt":        14,
				"my_transfer/subdir/abc.txt": 13,
			},
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
		tc := tc
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
			assert.DeepEqual(t, res, archivezip.Result{Path: td.Join(transferName + ".zip")})

			// Confirm the zip has the expected contents.
			rc, err := zip.OpenReader(td.Join(transferName + ".zip"))
			assert.NilError(t, err)
			t.Cleanup(func() { rc.Close() })

			files := make(map[string]int64, len(rc.File))
			for _, f := range rc.File {
				files[f.Name] = f.FileInfo().Size()
			}
			assert.DeepEqual(t, files, tc.want)
		})
	}
}
