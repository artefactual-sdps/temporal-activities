package bagit_test

import (
	"testing"

	bagit_gython "github.com/artefactual-labs/bagit-gython"
	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/bagit"
)

func validTestBag(t *testing.T) string {
	d := tfs.NewDir(t, "temporal-activities-test",
		tfs.WithFile(
			"bag-info.txt",
			`Bag-Software-Agent: bagit.py v1.8.1 <https://github.com/LibraryOfCongress/bagit-python>
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
			`Bag-Software-Agent: bagit.py v1.8.1 <https://github.com/LibraryOfCongress/bagit-python>
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

func TestValidateActivity(t *testing.T) {
	t.Parallel()

	// bagit_gython.NewBagIt() is expensive, so only call it once.
	validator, err := bagit_gython.NewBagIt()
	assert.NilError(t, err)
	defer t.Cleanup(func() { validator.Cleanup() })

	type test struct {
		name    string
		cfg     bagit.Config
		params  bagit.ValidateActivityParams
		want    bagit.ValidateActivityResult
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Validates a bag",
			params: bagit.ValidateActivityParams{
				Path: validTestBag(t),
			},
			want: bagit.ValidateActivityResult{
				Valid: true,
			},
		},
		{
			name: "Returns a validation error",
			params: bagit.ValidateActivityParams{
				Path: invalidTestBag(t),
			},
			want: bagit.ValidateActivityResult{
				Valid: false,
				Error: "invalid: Payload-Oxum validation failed. Expected 2 files and 38 bytes but found 1 files and 19 bytes",
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				bagit.NewValidateActivity(tt.cfg, validator).Execute,
				temporalsdk_activity.RegisterOptions{Name: bagit.ValidateActivityName},
			)

			enc, err := env.ExecuteActivity(bagit.ValidateActivityName, tt.params)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			var result bagit.ValidateActivityResult
			_ = enc.Get(&result)
			assert.DeepEqual(t, result, tt.want)
		})
	}
}