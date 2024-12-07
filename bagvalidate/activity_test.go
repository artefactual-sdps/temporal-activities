package bagvalidate_test

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
)

const (
	dirMode        fs.FileMode = 0o700
	fileMode       fs.FileMode = 0o600
	sha512manifest string      = `946af3bfd3b0b84ea0d99136085dcd66ee7e769371dbcd097ed35fd377116087e25d004afd68dc48e4eb0bcb6a434b04078577b531a7da1452296d1ae98d20b3  data/another.txt
8cbdd4ed5452f7c066509c066d5ea87fc03f30b0c67153624a1bce4d6e14b6709b5e78caf723cdf419d0efad4db96ba1cad3196783c26a7743029459bdd148b0  data/small.txt
`
)

func validTestBag(t *testing.T) string {
	d := tfs.NewDir(t, "temporal-activities-test",
		tfs.WithFile(
			"bag-info.txt",
			`Bag-Software-Agent: bagvalidate.py v1.8.1 <https://github.com/LibraryOfCongress/bagit-python>
Bagging-Date: 2024-07-04
Payload-Oxum: 38.2
`,
			tfs.WithMode(fileMode),
		),
		tfs.WithFile(
			"bagit.txt",
			`BagIt-Version: 0.97
Tag-File-Character-Encoding: UTF-8`,
			tfs.MatchAnyFileContent, tfs.WithMode(fileMode),
		),
		tfs.WithFile("manifest-sha512.txt", sha512manifest, tfs.WithMode(fileMode)),
		tfs.WithFile("tagmanifest-sha512.txt", "", tfs.MatchAnyFileContent, tfs.WithMode(fileMode)),
		tfs.WithDir("data", tfs.WithMode(dirMode),
			tfs.WithFile("another.txt", "I am another file.\n", tfs.WithMode(fileMode)),
			tfs.WithFile("small.txt", "I am a small file.\n", tfs.WithMode(fileMode)),
		),
	)

	return d.Path()
}

func invalidTestBag(t *testing.T) string {
	d := tfs.NewDir(t, "temporal-activities-test",
		tfs.WithFile(
			"bag-info.txt",
			`Bag-Software-Agent: bagvalidate.py v1.8.1 <https://github.com/LibraryOfCongress/bagit-python>
Bagging-Date: 2024-07-04
Payload-Oxum: 38.2
`,
			tfs.WithMode(fileMode),
		),
		tfs.WithFile(
			"bagit.txt",
			`BagIt-Version: 0.97
Tag-File-Character-Encoding: UTF-8`,
			tfs.MatchAnyFileContent, tfs.WithMode(fileMode),
		),
		tfs.WithFile("manifest-sha512.txt", sha512manifest, tfs.WithMode(fileMode)),
		tfs.WithFile("tagmanifest-sha512.txt", "", tfs.MatchAnyFileContent, tfs.WithMode(fileMode)),
		tfs.WithDir("data", tfs.WithMode(dirMode),
			tfs.WithFile("small.txt", "I am a small file.\n", tfs.WithMode(fileMode)),
		),
	)

	return d.Path()
}

func TestActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name   string
		params bagvalidate.Params
		want   bagvalidate.Result
	}
	for _, tt := range []test{
		{
			name: "Validates a bag",
			params: bagvalidate.Params{
				Path: validTestBag(t),
			},
			want: bagvalidate.Result{
				Valid: true,
			},
		},
		{
			name: "Returns a validation error",
			params: bagvalidate.Params{
				Path: invalidTestBag(t),
			},
			want: bagvalidate.Result{
				Valid: false,
				Error: "invalid: payload-oxum validation failed. expected 2 files and 38 bytes but found 1 files and 19 bytes",
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := bagvalidate.NewValidator()

			// Execute activity with test data.
			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				bagvalidate.New(validator).Execute,
				temporalsdk_activity.RegisterOptions{Name: bagvalidate.Name},
			)

			enc, _ := env.ExecuteActivity(bagvalidate.Name, tt.params)

			// Test activity result.
			var result bagvalidate.Result
			fmt.Println(result)
			_ = enc.Get(&result)
			assert.DeepEqual(t, result, tt.want)
		})
	}
}

func TestActivitySystemError(t *testing.T) {
	t.Parallel()

	validator := bagvalidate.NewMockValidator().SetErr(errors.New("transporter accident"))
	ts := &temporalsdk_testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()
	env.RegisterActivityWithOptions(
		bagvalidate.New(validator).Execute,
		temporalsdk_activity.RegisterOptions{Name: bagvalidate.Name},
	)

	_, err := env.ExecuteActivity(bagvalidate.Name, bagvalidate.Params{})
	assert.Error(
		t,
		err,
		"activity error (type: bag-validate, scheduledEventID: 0, startedEventID: 0, identity: ): transporter accident",
	)
}
