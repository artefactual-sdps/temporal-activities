package premis_objects_test

import (
	pseudorand "math/rand"
	"os"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	premis_objects "github.com/artefactual-sdps/temporal-activities/premis/objects"
)

const expectedPREMISWithFile = `<?xml version="1.0" encoding="UTF-8"?>
<premis:premis xmlns:premis="http://www.loc.gov/premis/v3" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.loc.gov/premis/v3 https://www.loc.gov/standards/premis/premis.xsd" version="3.0">
  <premis:object xsi:type="premis:file">
    <premis:objectIdentifier>
      <premis:objectIdentifierType>UUID</premis:objectIdentifierType>
      <premis:objectIdentifierValue>52fdfc07-2182-454f-963f-5f0f9a621d72</premis:objectIdentifierValue>
    </premis:objectIdentifier>
    <premis:objectCharacteristics>
      <premis:format>
        <premis:formatDesignation>
          <premis:formatName></premis:formatName>
        </premis:formatDesignation>
      </premis:format>
    </premis:objectCharacteristics>
    <premis:originalName>somefile.txt</premis:originalName>
  </premis:object>
</premis:premis>
`

const expectedPREMISNoFiles = `<?xml version="1.0" encoding="UTF-8"?>
<premis:premis xmlns:premis="http://www.loc.gov/premis/v3" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.loc.gov/premis/v3 https://www.loc.gov/standards/premis/premis.xsd" version="3.0"></premis:premis>
`

func TestAddPREMISObjects(t *testing.T) {
	t.Parallel()

	// Test transfer with one file.
	transferOneFile := fs.NewDir(t, "",
		fs.WithFile("somefile.txt", "somestuff"),
	)

	// Test transfer with no files.
	transferNoFiles := fs.NewDir(t, "")

	tests := []struct {
		name       string
		params     premis_objects.Params
		result     premis_objects.Result
		wantPREMIS string
		wantErr    string
	}{
		{
			name: "Add PREMIS objects for transfer with one file",
			params: premis_objects.Params{
				SIPPath:        transferOneFile.Path(),
				PREMISFilePath: transferOneFile.Join("metadata", "premis.xml"),
			},
			result:     premis_objects.Result{},
			wantPREMIS: expectedPREMISWithFile,
		},
		{
			name: "Add PREMIS objects for empty transfer",
			params: premis_objects.Params{
				SIPPath:        transferNoFiles.Path(),
				PREMISFilePath: transferNoFiles.Join("metadata", "premis.xml"),
			},
			result:     premis_objects.Result{},
			wantPREMIS: expectedPREMISNoFiles,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			rng := pseudorand.New(pseudorand.NewSource(1)) // #nosec G404
			env.RegisterActivityWithOptions(
				premis_objects.New(rng).Execute,
				temporalsdk_activity.RegisterOptions{Name: premis_objects.Name},
			)

			var res premis_objects.Result
			future, err := env.ExecuteActivity(premis_objects.Name, tt.params)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("error is nil, expecting: %q", tt.wantErr)
				} else {
					assert.ErrorContains(t, err, tt.wantErr)
				}

				return
			}
			assert.NilError(t, err)

			future.Get(&res)
			assert.DeepEqual(t, res, tt.result)

			b, err := os.ReadFile(tt.params.PREMISFilePath)
			assert.NilError(t, err)
			assert.Equal(t, string(b), tt.wantPREMIS)
		})
	}
}
