package bagextract_test

import (
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/bagextract"
)

func TestActivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		params  func(string) bagextract.Params
		result  func(string) bagextract.Result
		wantFS  fs.Manifest
		wantErr string
	}{
		{
			name: "Extracts a bag",
			path: fs.NewDir(t, "enduro-test",
				fs.WithDir("data",
					fs.WithDir("d_0000001",
						fs.WithFile("Prozess_Digitalisierung_PREMIS.xml", ""),
					),
					fs.WithDir("additional"),
				),
				fs.WithFile("bagit.txt", ""),
				fs.WithFile("manifest-md5.txt", ""),
			).Path(),
			params: func(path string) bagextract.Params {
				return bagextract.Params{Path: path}
			},
			result: func(path string) bagextract.Result {
				return bagextract.Result{Path: path}
			},
			wantFS: fs.Expected(t,
				fs.WithDir("d_0000001",
					fs.WithFile("Prozess_Digitalisierung_PREMIS.xml", ""),
				),
				fs.WithDir("additional"),
			),
		},
		{
			name: "Does nothing when path is not a bag",
			path: fs.NewDir(t, "enduro-test",
				fs.WithDir("d_0000001",
					fs.WithFile("Prozess_Digitalisierung_PREMIS.xml", ""),
				),
				fs.WithDir("additional"),
			).Path(),
			params: func(path string) bagextract.Params {
				return bagextract.Params{Path: path}
			},
			result: func(path string) bagextract.Result {
				return bagextract.Result{Path: path}
			},
			wantFS: fs.Expected(t,
				fs.WithDir("d_0000001",
					fs.WithFile("Prozess_Digitalisierung_PREMIS.xml", ""),
				),
				fs.WithDir("additional"),
			),
		},
		{
			name: "Errors when bag is missing data dir",
			path: fs.NewDir(t, "enduro-test",
				fs.WithDir("content",
					fs.WithDir("d_0000001",
						fs.WithFile("Prozess_Digitalisierung_PREMIS.xml", ""),
					),
					fs.WithDir("additional"),
				),
				fs.WithFile("bagit.txt", ""),
			).Path(),
			params: func(path string) bagextract.Params {
				return bagextract.Params{Path: path}
			},
			wantErr: "activity error (type: bagextract, scheduledEventID: 0, startedEventID: 0, identity: ): missing data directory",
		},
		{
			name: "Errors when data is not a directory",
			path: fs.NewDir(t, "enduro-test",
				fs.WithFile("data", ""),
				fs.WithFile("bagit.txt", ""),
			).Path(),
			params: func(path string) bagextract.Params {
				return bagextract.Params{Path: path}
			},
			wantErr: "read data dir:",
		},
		{
			name: "Extracts a bag and keeps metadata directory",
			path: fs.NewDir(t, "enduro-test",
				fs.WithDir("data",
					fs.WithDir("d_0000001",
						fs.WithFile("Prozess_Digitalisierung_PREMIS.xml", ""),
					),
					fs.WithDir("additional"),
				),
				fs.WithDir("metadata"),
				fs.WithFile("bagit.txt", ""),
				fs.WithFile("manifest-md5.txt", ""),
			).Path(),
			params: func(path string) bagextract.Params {
				return bagextract.Params{Path: path, Keep: []string{"metadata"}}
			},
			result: func(path string) bagextract.Result {
				return bagextract.Result{Path: path}
			},
			wantFS: fs.Expected(t,
				fs.WithDir("d_0000001",
					fs.WithFile("Prozess_Digitalisierung_PREMIS.xml", ""),
				),
				fs.WithDir("additional"),
				fs.WithDir("metadata"),
			),
		},
		{
			name: "Extracts a bag and keeps custom file",
			path: fs.NewDir(t, "enduro-test",
				fs.WithDir("data",
					fs.WithDir("d_0000001",
						fs.WithFile("Prozess_Digitalisierung_PREMIS.xml", ""),
					),
					fs.WithDir("additional"),
				),
				fs.WithFile("metadata.txt", ""),
				fs.WithFile("bagit.txt", ""),
				fs.WithFile("manifest-md5.txt", ""),
			).Path(),
			params: func(path string) bagextract.Params {
				return bagextract.Params{Path: path, Keep: []string{"metadata.txt"}}
			},
			result: func(path string) bagextract.Result {
				return bagextract.Result{Path: path}
			},
			wantFS: fs.Expected(t,
				fs.WithDir("d_0000001",
					fs.WithFile("Prozess_Digitalisierung_PREMIS.xml", ""),
				),
				fs.WithDir("additional"),
				fs.WithFile("metadata.txt", ""),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				bagextract.New().Execute,
				temporalsdk_activity.RegisterOptions{Name: bagextract.Name},
			)

			var res bagextract.Result
			future, err := env.ExecuteActivity(bagextract.Name, tt.params(tt.path))
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("error is nil, expecting: %q", tt.wantErr)
				} else {
					assert.ErrorContains(t, err, tt.wantErr)
				}

				return
			}
			assert.NilError(t, err)

			assert.NilError(t, future.Get(&res))
			assert.DeepEqual(t, res, tt.result(tt.path))
			assert.Assert(t, fs.Equal(tt.path, tt.wantFS))
		})
	}
}
