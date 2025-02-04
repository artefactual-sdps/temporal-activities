package ffvalidate_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/tonglil/buflogr"
	"go.artefactual.dev/tools/temporal"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_interceptor "go.temporal.io/sdk/interceptor"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	temporalsdk_worker "go.temporal.io/sdk/worker"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/ffvalidate"
)

const pngContent = "\x89PNG\r\n\x1a\n\x00\x00\x00\x0DIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90\x77\x53\xDE\x00\x00\x00\x00IEND\xAE\x42\x60\x82"

func TestValidateFileFormats(t *testing.T) {
	t.Parallel()

	invalidFormatsDir := fs.NewDir(t, "",
		fs.WithDir("invalid_sip",
			fs.WithDir("dir",
				fs.WithFile("file1.png", pngContent),
			),
			fs.WithFile("file2.png", pngContent),
		),
	).Path()

	validFormatsDir := fs.NewDir(t, "",
		fs.WithDir("valid_sip",
			fs.WithFile("file1.txt", "content"),
			fs.WithDir("dir",
				fs.WithFile("file2.txt", "content"),
			),
		),
	).Path()

	emptyList := fs.NewDir(t, "", fs.WithFile("allowlist.csv", "")).Join("allowlist.csv")

	invalidCSVList := fs.NewDir(t, "", fs.WithFile("allowlist.csv", `PRONOM PUID
fmt/95,fmt/96
`)).Join("allowlist.csv")

	noPUIDList := fs.NewDir(t, "", fs.WithFile("allowlist.csv", `pronom id
fmt/95
`)).Join("allowlist.csv")

	weirdButValidList := fs.NewDir(t, "", fs.WithFile("allowlist.csv", `Pronom puid,Format
fmt/95,"PDF/A "

" x-fmt/111 ","text file"
`)).Join("allowlist.csv")

	tests := []struct {
		name    string
		cfg     ffvalidate.Config
		params  ffvalidate.Params
		want    ffvalidate.Result
		wantErr string
		wantLog string
	}{
		{
			name:   "Succeeds with valid formats",
			cfg:    ffvalidate.Config{AllowlistPath: "./testdata/allowed_file_formats.csv"},
			params: ffvalidate.Params{Path: validFormatsDir},
		},
		{
			name:   "Succeeds with weird but valid CSV values",
			cfg:    ffvalidate.Config{AllowlistPath: weirdButValidList},
			params: ffvalidate.Params{Path: validFormatsDir},
		},
		{
			name:   "Fails with invalid formats",
			cfg:    ffvalidate.Config{AllowlistPath: "./testdata/allowed_file_formats.csv"},
			params: ffvalidate.Params{Path: invalidFormatsDir},
			want: ffvalidate.Result{
				Failures: []string{
					fmt.Sprintf(
						`file format %q not allowed: "invalid_sip/dir/file1.png"`,
						"fmt/11",
					),
					fmt.Sprintf(
						`file format %q not allowed: "invalid_sip/file2.png"`,
						"fmt/11",
					),
				},
			},
		},
		{
			name:    "Fails with empty source",
			cfg:     ffvalidate.Config{AllowlistPath: "./testdata/allowed_file_formats.csv"},
			params:  ffvalidate.Params{Path: fs.NewDir(t, "", fs.WithFile("file.txt", "")).Path()},
			wantErr: "validate-file-formats: check allowed formats: identify format: empty source",
		},
		{
			name:    "Does nothing when no allowlist path configured",
			params:  ffvalidate.Params{Path: validFormatsDir},
			wantLog: "V[1] Executing activity. ActivityID 0 ActivityType validate-file-formats\nINFO validate-file-formats: No allowlist path configured, skipping file format validation ActivityID 0 ActivityType validate-file-formats\n",
		},
		{
			name:    "Errors when allowlist path doesn't exist",
			cfg:     ffvalidate.Config{AllowlistPath: filepath.Join("/dev/null/allowlist.csv")},
			params:  ffvalidate.Params{Path: validFormatsDir},
			wantErr: "validate-file-formats: open /dev/null/allowlist.csv: not a directory",
		},
		{
			name:    "Errors when allowlist is empty",
			cfg:     ffvalidate.Config{AllowlistPath: emptyList},
			params:  ffvalidate.Params{Path: validFormatsDir},
			wantErr: "validate-file-formats: load allowed formats: no allowed file formats",
		},
		{
			name:    "Errors when no PRONOM PUID column exists",
			cfg:     ffvalidate.Config{AllowlistPath: noPUIDList},
			params:  ffvalidate.Params{Path: validFormatsDir},
			wantErr: "validate-file-formats: load allowed formats: missing \"PRONOM PUID\" column",
		},
		{
			name:    "Errors when allowlist is not a valid CSV format",
			cfg:     ffvalidate.Config{AllowlistPath: invalidCSVList},
			params:  ffvalidate.Params{Path: validFormatsDir},
			wantErr: "validate-file-formats: load allowed formats: invalid CSV: record on line 2: wrong number of fields",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logbuf bytes.Buffer
			logger := buflogr.NewWithBuffer(&logbuf)

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.SetWorkerOptions(temporalsdk_worker.Options{
				Interceptors: []temporalsdk_interceptor.WorkerInterceptor{
					temporal.NewLoggerInterceptor(logger),
				},
			})
			env.RegisterActivityWithOptions(
				ffvalidate.New(tt.cfg).Execute,
				temporalsdk_activity.RegisterOptions{Name: ffvalidate.Name},
			)

			enc, err := env.ExecuteActivity(ffvalidate.Name, tt.params)
			if tt.wantErr != "" {
				prefix := "activity error (type: validate-file-formats, scheduledEventID: 0, startedEventID: 0, identity: ): "
				if err == nil {
					t.Errorf("error is nil, expecting: %q", tt.wantErr)
				} else {
					assert.Error(t, err, prefix+tt.wantErr)
				}

				return
			}
			assert.NilError(t, err)

			if tt.wantLog != "" {
				assert.Equal(t, logbuf.String(), tt.wantLog)
			}

			var got ffvalidate.Result
			_ = enc.Get(&got)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
