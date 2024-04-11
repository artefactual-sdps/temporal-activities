package removefiles_test

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/removefiles"
)

func TestActivity(t *testing.T) {
	t.Parallel()

	type args struct {
		path   func(td *tfs.Dir) string
		remove string
	}
	type Test struct {
		name    string
		params  args
		want    removefiles.ActivityResult
		wantFs  tfs.Manifest
		wantErr string
	}
	for _, tt := range []Test{
		{
			name: "Deletes the .DS_Store file",
			params: args{
				path: func(td *tfs.Dir) string {
					return td.Path()
				},
				remove: "Thumbs.db, .DS_Store",
			},
			want:   removefiles.ActivityResult{Count: 1},
			wantFs: tfs.Expected(t, tfs.WithFile("keepme", "don't delete me.")),
		},
		{
			name: "Errors if Path doesn't exist",
			params: args{
				path: func(td *tfs.Dir) string {
					return td.Join("not_here")
				},
				remove: "Thumbs.db, .DS_Store",
			},
			wantErr: "activity error (type: remove-files-activity, scheduledEventID: 0, startedEventID: 0, identity: ): RemoveFilesActivity: remove: stat %s: no such file or directory",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			td := tfs.NewDir(t, "remove-files-activity-test",
				tfs.WithFile(".DS_Store", "delete me!"),
				tfs.WithFile("keepme", "keep me!"),
			)

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				removefiles.NewActivity(logr.Discard()).Execute,
				temporalsdk_activity.RegisterOptions{
					Name: removefiles.ActivityName,
				},
			)

			enc, err := env.ExecuteActivity(
				removefiles.ActivityName,
				&removefiles.ActivityParams{
					Path:        tt.params.path(td),
					RemoveNames: tt.params.remove,
				},
			)

			if tt.wantErr != "" {
				assert.Error(t, err, fmt.Sprintf(tt.wantErr, tt.params.path(td)))
				return
			}
			assert.NilError(t, err)

			var result removefiles.ActivityResult
			_ = enc.Get(&result)
			assert.Equal(t, result, tt.want)
		})
	}
}
