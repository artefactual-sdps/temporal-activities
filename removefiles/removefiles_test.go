package removefiles_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/removefiles"
)

func TestRemove(t *testing.T) {
	t.Parallel()

	type args struct {
		path   func(td *tfs.Dir) string
		remove string
	}
	type Test struct {
		name    string
		params  args
		want    int
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
			want:   1,
			wantFs: tfs.Expected(t, tfs.WithFile("keepme", "don't delete me.")),
		},
		{
			name: "Does nothing when no remove filenames given",
			params: args{
				path: func(td *tfs.Dir) string {
					return td.Path()
				},
			},
			want: 0,
			wantFs: tfs.Expected(t,
				tfs.WithFile(".DS_Store", "delete me."),
				tfs.WithFile("keepme", "don't delete me."),
			),
		},
		{
			name: "Errors if path doesn't exist",
			params: args{
				path: func(td *tfs.Dir) string {
					return td.Join("not_here")
				},
				remove: "Thumbs.db, .DS_Store",
			},
			wantErr: "remove: stat %s: no such file or directory",
		},
		{
			name: "Errors if path is not a directory",
			params: args{
				path: func(td *tfs.Dir) string {
					return td.Join("keepme")
				},
				remove: "Thumbs.db, .DS_Store",
			},
			wantErr: "remove: path: %q: not a directory",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			td := tfs.NewDir(t, "remove-files-test",
				tfs.WithFile(".DS_Store", "delete me."),
				tfs.WithFile("keepme", "don't delete me."),
			)

			got, err := removefiles.Remove(
				tt.params.path(td),
				tt.params.remove,
			)
			if tt.wantErr != "" {
				assert.Error(t, err, fmt.Sprintf(
					tt.wantErr,
					tt.params.path(td),
				))
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, got, tt.want)
			assert.Assert(t, tfs.Equal(td.Path(), tt.wantFs))
		})
	}
}
