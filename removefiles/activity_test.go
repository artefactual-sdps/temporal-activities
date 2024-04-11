package removefiles_test

import (
	"fmt"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/removefiles"
)

func TestActivity(t *testing.T) {
	t.Parallel()

	type Test struct {
		name    string
		params  removefiles.ActivityParams
		want    removefiles.ActivityResult
		wantFs  tfs.Manifest
		wantErr string
	}
	for _, tt := range []Test{
		{
			name: "Deletes all .DS_Store files",
			params: removefiles.ActivityParams{
				Path: tfs.NewDir(t, "remove-files-activity-test",
					tfs.WithFile(".DS_Store", "delete me."),
					tfs.WithFile("keepme", "don't delete me."),
					tfs.WithDir("subdir",
						tfs.WithFile("Thumbs.db", "delete me too."),
						tfs.WithFile("keepme2", "keep me too."),
					),
				).Path(),
				RemoveNames: "Thumbs.db, .DS_Store\n",
			},
			want: removefiles.ActivityResult{Count: 2},
			wantFs: tfs.Expected(t,
				tfs.WithFile("keepme", "don't delete me."),
				tfs.WithDir("subdir",
					tfs.WithFile("keepme2", "keep me too."),
				),
			),
		},
		{
			name: "Deletes a .git directory",
			params: removefiles.ActivityParams{
				Path: tfs.NewDir(t, "remove-files-activity-test",
					tfs.WithFile("keepme", "don't delete me."),
					tfs.WithDir(".git",
						tfs.WithFile(
							"HEAD",
							`ref: refs/heads/dev/issue-2-remove-files-activity
`),
						tfs.WithFile(
							"ORIG_HEAD",
							`e47c6e469a15d78a3fca8512f6fe9b6cec9a1e68
`),
					),
				).Path(),
				RemoveNames: ".git",
			},
			want: removefiles.ActivityResult{Count: 1},
			wantFs: tfs.Expected(t,
				tfs.WithFile("keepme", "don't delete me."),
			),
		},
		{
			name: "Does nothing when no remove names given",
			params: removefiles.ActivityParams{
				Path: tfs.NewDir(t, "remove-files-activity-test",
					tfs.WithFile(".DS_Store", "delete me."),
					tfs.WithFile("keepme", "don't delete me."),
				).Path(),
			},
			want: removefiles.ActivityResult{Count: 0},
			wantFs: tfs.Expected(t,
				tfs.WithFile(".DS_Store", "delete me."),
				tfs.WithFile("keepme", "don't delete me."),
			),
		},
		{
			name: "Errors if Path doesn't exist",
			params: removefiles.ActivityParams{
				Path:        "not_there",
				RemoveNames: "Thumbs.db, .DS_Store",
			},
			wantErr: "activity error (type: remove-files-activity, scheduledEventID: 0, startedEventID: 0, identity: ): RemoveFilesActivity: remove: stat %s: no such file or directory",
		},
		{
			name: "Errors if path is not a directory",
			params: removefiles.ActivityParams{
				Path: tfs.NewDir(t, "remove-files-test",
					tfs.WithFile(".DS_Store", "delete me."),
					tfs.WithFile("keepme", "don't delete me."),
				).Join("keepme"),
				RemoveNames: "Thumbs.db, .DS_Store",
			},
			wantErr: "activity error (type: remove-files-activity, scheduledEventID: 0, startedEventID: 0, identity: ): RemoveFilesActivity: remove: path: %q: not a directory",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				removefiles.NewActivity().Execute,
				temporalsdk_activity.RegisterOptions{
					Name: removefiles.ActivityName,
				},
			)

			enc, err := env.ExecuteActivity(removefiles.ActivityName, &tt.params)

			if tt.wantErr != "" {
				assert.Error(t, err, fmt.Sprintf(tt.wantErr, tt.params.Path))
				return
			}
			assert.NilError(t, err)

			var result removefiles.ActivityResult
			_ = enc.Get(&result)
			assert.Equal(t, result, tt.want)
			assert.Assert(t, tfs.Equal(tt.params.Path, tt.wantFs))
		})
	}
}
