package archive_test

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

	"github.com/artefactual-sdps/temporal-activities/archive"
)

const smallTxtContent = "I am a small file.\n"

func TestExtractActivity(t *testing.T) {
	t.Parallel()

	// Use a shared dest directory to test that multiple simultaneous extracts
	// don't conflict on write.
	dest := tfs.NewDir(t, "sdps_extract_test").Path()

	type test struct {
		name    string
		cfg     archive.Config
		params  archive.ExtractActivityParams
		wantFs  tfs.Manifest
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Extracts a tar gzip archive",
			params: archive.ExtractActivityParams{
				SourcePath: filepath.Join("testdata", "transfer.tar.gz"),
				DestPath:   dest,
			},
			wantFs: tfs.Expected(t,
				tfs.WithFile("small.txt", smallTxtContent, tfs.WithMode(0o600)),
			),
		},
		{
			name: "Extracts a tar gzip archive with no DestPath",
			params: archive.ExtractActivityParams{
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
			params: archive.ExtractActivityParams{
				SourcePath: filepath.Join("testdata", "transfer_no_subdir.zip"),
				DestPath:   dest,
			},
			wantFs: tfs.Expected(t,
				tfs.WithFile("small.txt", smallTxtContent, tfs.WithMode(0o600)),
			),
		},
		{
			name: "Extracts a zip archive with a sub-directory and a file",
			params: archive.ExtractActivityParams{
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
			cfg:  archive.Config{DirMode: 0o750, FileMode: 0o640},
			params: archive.ExtractActivityParams{
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
			params: archive.ExtractActivityParams{
				SourcePath: filepath.Join("testdata", "transfer.7z"),
				DestPath:   dest,
			},
			wantFs: tfs.Expected(t,
				tfs.WithFile("small.txt", smallTxtContent, tfs.WithMode(0o600)),
			),
		},
		{
			name: "Errors when SourcePath is a dir",
			params: archive.ExtractActivityParams{
				SourcePath: "testdata",
			},
			wantErr: fmt.Sprintf(
				"activity error (type: extract-archive-activity, scheduledEventID: 0, startedEventID: 0, identity: ): %s",
				archive.ErrNotAFile,
			),
		},
		{
			name: "Errors when SourcePath is a non-archive file",
			params: archive.ExtractActivityParams{
				SourcePath: filepath.Join("testdata", "small.txt"),
			},
			wantErr: fmt.Sprintf(
				"activity error (type: extract-archive-activity, scheduledEventID: 0, startedEventID: 0, identity: ): %s",
				archive.ErrInvalidArchive,
			),
		},
		{
			name: "Errors on corrupt archive",
			params: archive.ExtractActivityParams{
				SourcePath: filepath.Join("testdata", "corrupt.zip"),
			},
			wantErr: "activity error (type: extract-archive-activity, scheduledEventID: 0, startedEventID: 0, identity: ): extract archive: extract: zip: not a valid zip file",
		},
		{
			name: "Errors when source path doesn't exist",
			params: archive.ExtractActivityParams{
				SourcePath: filepath.Join("testdata", "missing.zip"),
			},
			wantErr: "activity error (type: extract-archive-activity, scheduledEventID: 0, startedEventID: 0, identity: ): extract archive: stat testdata/missing.zip: no such file or directory",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				archive.NewExtractActivity(tt.cfg).Execute,
				temporalsdk_activity.RegisterOptions{Name: archive.ExtractActivityName},
			)

			enc, err := env.ExecuteActivity(archive.ExtractActivityName, &tt.params)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var result archive.ExtractActivityResult
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
