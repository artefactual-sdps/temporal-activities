package xsdvalidate

import (
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/fs"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/xsdvalidate"
)

func TestXSDValidateActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		params  xsdvalidate.XSDValidateActivityParams
		want    tfs.Manifest
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Test test test",
			params: xsdvalidate.XSDValidateActivityParams{
				DirectoryPath: "/tmp",
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				xsdvalidate.NewXSDValidateActivity().Execute,
				temporalsdk_activity.RegisterOptions{Name: xsdvalidate.XSDValidateActivityName},
			)

			var res xsdvalidate.XSDValidateResult
			future, err := env.ExecuteActivity(xsdvalidate.XSDValidateActivityName, tt.params)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			future.Get(&res)
			assert.NilError(t, err)
		})
	}
}

func MatchingFilesRelativeToDirectory(t *testing.T) {
	pattern := "^premis*.xml$"

	testDirectoryPath := fs.NewDir(t, "",
		fs.WithDir("somedir",
			fs.WithDir("somesubdir",
				fs.WithFile("stuff.xml", ""),
				fs.WithFile("premis.xml", ""),
			),
		),
		fs.WithDir("anotherdir",
			fs.WithFile("00000001_PREMIS.xml", ""),
			fs.WithFile("cat.jpg", ""),
			fs.WithDir("content",
				fs.WithFile("premis.XML", ""),
				fs.WithFile("goat.xml", ""),
			),
		),
	).Path()

	files := xsdvalidate.MatchingFilesRelativeToDirectory(testDirectoryPath, pattern)

	panic(files)
}
