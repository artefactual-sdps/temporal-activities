package removepaths_test

import (
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/removepaths"
)

func testDir(t *testing.T) *tfs.Dir {
	t.Helper()

	return tfs.NewDir(t, "temporal-activity-test",
		tfs.WithDir("delete_dir",
			tfs.WithFile("delete.txt", "delete me."),
		),
		tfs.WithFile("keepme", "don't delete me."),
		tfs.WithFile("delete2.txt", "delete me too."),
	)
}

func TestActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name   string
		params removepaths.Params
		wantFs tfs.Manifest
	}
	for _, tt := range []test{
		{
			name: "Deletes the given paths",
			params: removepaths.Params{
				Paths: []string{
					"delete_dir",
					"delete2.txt",
				},
			},
			wantFs: tfs.Expected(t,
				tfs.WithFile("keepme", "don't delete me."),
			),
		},
		{
			name: "No error when a path doesn't exist",
			params: removepaths.Params{
				Paths: []string{
					"delete_dir",
					"delete2.txt",
					"unknown",
				},
			},
			wantFs: tfs.Expected(t,
				tfs.WithFile("keepme", "don't delete me."),
			),
		},
	} {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				removepaths.New().Execute,
				temporalsdk_activity.RegisterOptions{Name: removepaths.Name},
			)

			td := testDir(t)
			for i, p := range tt.params.Paths {
				tt.params.Paths[i] = td.Join(p)
			}

			enc, err := env.ExecuteActivity(removepaths.Name, &tt.params)
			assert.NilError(t, err)

			var result removepaths.Result
			_ = enc.Get(&result)
			assert.Equal(t, result, removepaths.Result{})
			assert.Assert(t, tfs.Equal(td.Path(), tt.wantFs))
		})
	}
}
