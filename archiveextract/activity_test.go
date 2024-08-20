package archiveextract_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cp "github.com/otiai10/copy"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
)

const smallTxtContent = "I am a small file.\n"

func TestActivity(t *testing.T) {
	t.Parallel()

	// Use a shared dest directory to test that multiple simultaneous extracts
	// don't conflict on write.
	dest := tfs.NewDir(t, "sdps_extract_test").Path()

	type test struct {
		name    string
		cfg     archiveextract.Config
		params  archiveextract.Params
		wantFs  tfs.Manifest
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Extracts a tar gzip archive",
			params: archiveextract.Params{
				SourcePath: filepath.Join("testdata", "transfer.tar.gz"),
				DestPath:   dest,
			},
			wantFs: tfs.Expected(t,
				tfs.WithFile("small.txt", smallTxtContent, tfs.WithMode(0o600)),
			),
		},
		{
			name: "Extracts a tar gzip archive with no DestPath",
			params: archiveextract.Params{
				SourcePath: func() string {
					// Copy transfer.tar.gz to a temporary directory so we don't
					// make a random extract dir in testdata/.
					src := filepath.Join("testdata", "transfer.tar.gz")
					dest := filepath.Join(t.TempDir(), "transfer.tar.gz")
					if err := cp.Copy(src, dest); err != nil {
						t.Fatalf("Error copying %s to %s", src, dest)
					}
					return dest
				}(),
			},
			wantFs: tfs.Expected(t,
				tfs.WithFile("small.txt", smallTxtContent, tfs.WithMode(0o600)),
			),
		},
		{
			name: "Extracts a zip archive with no sub-directories",
			params: archiveextract.Params{
				SourcePath: filepath.Join("testdata", "transfer_no_subdir.zip"),
				DestPath:   dest,
			},
			wantFs: tfs.Expected(t,
				tfs.WithFile("small.txt", smallTxtContent, tfs.WithMode(0o600)),
			),
		},
		{
			name: "Extracts a zip archive with a sub-directory and a file",
			params: archiveextract.Params{
				SourcePath: filepath.Join("testdata", "transfer_subdir+file.zip"),
				DestPath:   dest,
			},
			wantFs: tfs.Expected(t,
				tfs.WithDir("subdir", tfs.WithMode(0o700)),
				tfs.WithFile("small.txt", smallTxtContent, tfs.WithMode(0o600)),
			),
		},
		{
			name: "Extracts a zip archive with explicit file modes",
			cfg:  archiveextract.Config{DirMode: 0o750, FileMode: 0o640},
			params: archiveextract.Params{
				SourcePath: filepath.Join("testdata", "transfer_subdir+file.zip"),
				DestPath:   dest,
			},
			wantFs: tfs.Expected(t,
				tfs.WithDir("subdir", tfs.WithMode(0o750)),
				tfs.WithFile("small.txt", smallTxtContent, tfs.WithMode(0o640)),
			),
		},
		{
			name: "Extracts a 7z archive",
			params: archiveextract.Params{
				SourcePath: filepath.Join("testdata", "transfer.7z"),
				DestPath:   dest,
			},
			wantFs: tfs.Expected(t,
				tfs.WithFile("small.txt", smallTxtContent, tfs.WithMode(0o600)),
			),
		},
		{
			name: "Errors when SourcePath is a dir",
			params: archiveextract.Params{
				SourcePath: "testdata",
			},
			wantErr: fmt.Sprintf(
				"activity error (type: archive-extract, scheduledEventID: 0, startedEventID: 0, identity: ): %s",
				archiveextract.ErrNotAFile,
			),
		},
		{
			name: "Errors when SourcePath is a non-archive file",
			params: archiveextract.Params{
				SourcePath: filepath.Join("testdata", "small.txt"),
			},
			wantErr: fmt.Sprintf(
				"activity error (type: archive-extract, scheduledEventID: 0, startedEventID: 0, identity: ): %s",
				archiveextract.ErrInvalidArchive,
			),
		},
		{
			name: "Errors on corrupt archive",
			params: archiveextract.Params{
				SourcePath: filepath.Join("testdata", "corrupt.zip"),
			},
			wantErr: "activity error (type: archive-extract, scheduledEventID: 0, startedEventID: 0, identity: ): archiveextract: extract: zip: not a valid zip file",
		},
		{
			name: "Errors when source path doesn't exist",
			params: archiveextract.Params{
				SourcePath: filepath.Join("testdata", "missing.zip"),
			},
			wantErr: "activity error (type: archive-extract, scheduledEventID: 0, startedEventID: 0, identity: ): archiveextract: stat testdata/missing.zip: no such file or directory",
		},
	} {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				archiveextract.New(tt.cfg).Execute,
				temporalsdk_activity.RegisterOptions{Name: archiveextract.Name},
			)

			enc, err := env.ExecuteActivity(archiveextract.Name, &tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var result archiveextract.Result
			_ = enc.Get(&result)

			if tt.params.DestPath != "" {
				// result.DestPath must be a subdirectory of params.DestPath.
				assert.Assert(t, strings.HasPrefix(result.ExtractPath, tt.params.DestPath))
			} else {
				os.RemoveAll(tt.params.DestPath)
			}

			assert.Assert(t, tfs.Equal(result.ExtractPath, tt.wantFs))
		})
	}
}
