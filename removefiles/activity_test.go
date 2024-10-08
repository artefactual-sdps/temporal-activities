package removefiles_test

import (
	"fmt"
	"regexp"
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
		name            string
		params          removefiles.Params
		want            removefiles.Result
		wantFs          tfs.Manifest
		wantErr         string
		wantPathInError bool
	}
	for _, tt := range []Test{
		{
			name: "Deletes all .DS_Store files",
			params: removefiles.Params{
				Path: tfs.NewDir(t, "remove-files-activity-test",
					tfs.WithFile(".DS_Store", "delete me."),
					tfs.WithFile("keepme", "don't delete me."),
					tfs.WithDir("subdir",
						tfs.WithFile("Thumbs.db", "delete me too."),
						tfs.WithFile("keepme2", "keep me too."),
					),
				).Path(),
				RemoveNames: []string{"Thumbs.db", ".DS_Store"},
			},
			want: removefiles.Result{Count: 2},
			wantFs: tfs.Expected(t,
				tfs.WithFile("keepme", "don't delete me."),
				tfs.WithDir("subdir",
					tfs.WithFile("keepme2", "keep me too."),
				),
			),
		},
		{
			name: "Deletes a .git directory",
			params: removefiles.Params{
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
				RemoveNames: []string{".git"},
			},
			want: removefiles.Result{Count: 1},
			wantFs: tfs.Expected(t,
				tfs.WithFile("keepme", "don't delete me."),
			),
		},
		{
			name: "Does nothing when no remove names/patterns given",
			params: removefiles.Params{
				Path: tfs.NewDir(t, "remove-files-activity-test",
					tfs.WithFile(".DS_Store", "delete me."),
					tfs.WithFile("keepme", "don't delete me."),
				).Path(),
			},
			want: removefiles.Result{Count: 0},
			wantFs: tfs.Expected(t,
				tfs.WithFile(".DS_Store", "delete me."),
				tfs.WithFile("keepme", "don't delete me."),
			),
		},
		{
			name: "Errors if Path doesn't exist",
			params: removefiles.Params{
				Path:        "not_there",
				RemoveNames: []string{"Thumbs.db", ".DS_Store"},
			},
			wantErr:         "activity error (type: remove-files, scheduledEventID: 0, startedEventID: 0, identity: ): removefiles: remove: stat %s: no such file or directory",
			wantPathInError: true,
		},
		{
			name: "Errors if path is not a directory",
			params: removefiles.Params{
				Path: tfs.NewDir(t, "remove-files-test",
					tfs.WithFile(".DS_Store", "delete me."),
					tfs.WithFile("keepme", "don't delete me."),
				).Join("keepme"),
				RemoveNames: []string{"Thumbs.db", ".DS_Store"},
			},
			wantErr:         "activity error (type: remove-files, scheduledEventID: 0, startedEventID: 0, identity: ): removefiles: remove: path: %q: not a directory",
			wantPathInError: true,
		},
		{
			name: "Deletes based on regular expressions",
			params: removefiles.Params{
				Path: tfs.NewDir(t, "remove-files-activity-test",
					tfs.WithFile("001_premis.xml", "delete me."),
					tfs.WithFile("keepme", "don't delete me."),
					tfs.WithDir("subdir",
						tfs.WithFile("METS.xml", "delete me too."),
						tfs.WithFile("keepme2", "keep me too."),
					),
				).Path(),
				RemovePatterns: []*regexp.Regexp{
					regexp.MustCompile("premis.xml$"),
					regexp.MustCompile("(?i)mets.xml$"),
				},
			},
			want: removefiles.Result{Count: 2},
			wantFs: tfs.Expected(t,
				tfs.WithFile("keepme", "don't delete me."),
				tfs.WithDir("subdir",
					tfs.WithFile("keepme2", "keep me too."),
				),
			),
		},
	} {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				removefiles.New().Execute,
				temporalsdk_activity.RegisterOptions{Name: removefiles.Name},
			)

			enc, err := env.ExecuteActivity(removefiles.Name, &tt.params)

			if tt.wantErr != "" {
				if tt.wantPathInError {
					tt.wantErr = fmt.Sprintf(tt.wantErr, tt.params.Path)
				}
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var result removefiles.Result
			_ = enc.Get(&result)
			assert.Equal(t, result, tt.want)
			assert.Assert(t, tfs.Equal(tt.params.Path, tt.wantFs))
		})
	}
}
