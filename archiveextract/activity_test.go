package archiveextract_test

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cp "github.com/otiai10/copy"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/archiveextract"
)

const smallTxtContent = "I am a small file.\n"

func zipWithDirectoryModTime(t *testing.T, modTime time.Time) string {
	t.Helper()

	archivePath := filepath.Join(t.TempDir(), "transfer.zip")
	f, err := os.Create(archivePath)
	assert.NilError(t, err)

	zw := zip.NewWriter(f)
	_, err = zw.CreateHeader(&zip.FileHeader{Name: "transfer/subdir/", Modified: modTime})
	assert.NilError(t, err)
	w, err := zw.CreateHeader(&zip.FileHeader{
		Name:     "transfer/subdir/small.txt",
		Method:   zip.Deflate,
		Modified: modTime,
	})
	assert.NilError(t, err)
	_, err = w.Write([]byte(smallTxtContent))
	assert.NilError(t, err)
	assert.NilError(t, zw.Close())
	assert.NilError(t, f.Close())

	return archivePath
}

func TestActivity(t *testing.T) {
	t.Parallel()

	// Use a shared dest directory to test that multiple simultaneous extracts
	// don't conflict on write.
	dest := tfs.NewDir(t, "sdps_extract_test").Path()
	fileModTime := time.Date(2023, time.June, 6, 22, 57, 31, 0, time.UTC)
	dirModTime := time.Date(2001, time.February, 3, 4, 5, 6, 0, time.UTC)
	dirModTimeArchive := zipWithDirectoryModTime(t, dirModTime)

	type test struct {
		name         string
		cfg          archiveextract.Config
		params       archiveextract.Params
		wantFs       tfs.Manifest
		wantModTimes map[string]time.Time
		wantErr      string
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
			wantModTimes: map[string]time.Time{"small.txt": fileModTime},
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
			wantModTimes: map[string]time.Time{"small.txt": fileModTime},
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
			wantModTimes: map[string]time.Time{"small.txt": fileModTime},
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
			name: "Restores directory modification time after contents",
			params: archiveextract.Params{
				SourcePath: dirModTimeArchive,
				DestPath:   dest,
			},
			wantFs: tfs.Expected(t,
				tfs.WithDir("subdir", tfs.WithMode(0o700),
					tfs.WithFile("small.txt", smallTxtContent, tfs.WithMode(0o600)),
				),
			),
			wantModTimes: map[string]time.Time{
				"subdir":           dirModTime,
				"subdir/small.txt": dirModTime,
			},
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
			wantModTimes: map[string]time.Time{
				".":         time.Date(2024, time.April, 17, 16, 40, 46, 499459000, time.UTC),
				"small.txt": fileModTime,
			},
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
			for path, want := range tt.wantModTimes {
				info, err := os.Stat(filepath.Join(result.ExtractPath, path))
				assert.NilError(t, err)
				assert.Equal(t, info.ModTime().UTC(), want)
			}
		})
	}
}
