package premis_event_test

import (
	"os"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/premis"
	premis_event "github.com/artefactual-sdps/temporal-activities/premis/event"
)

const expectedPREMISWithSuccessfulEvent = `<?xml version="1.0" encoding="UTF-8"?>
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
    <premis:linkingEventIdentifier>
      <premis:linkingEventIdentifierType></premis:linkingEventIdentifierType>
      <premis:linkingEventIdentifierValue></premis:linkingEventIdentifierValue>
    </premis:linkingEventIdentifier>
  </premis:object>
  <premis:event>
    <premis:eventIdentifier>
      <premis:eventIdentifierType></premis:eventIdentifierType>
      <premis:eventIdentifierValue></premis:eventIdentifierValue>
    </premis:eventIdentifier>
    <premis:eventType>someActivity</premis:eventType>
    <premis:eventDateTime></premis:eventDateTime>
    <premis:eventDetailInformation>
      <premis:eventDetail></premis:eventDetail>
    </premis:eventDetailInformation>
    <premis:eventOutcomeInformation>
      <premis:eventOutcome>valid</premis:eventOutcome>
    </premis:eventOutcomeInformation>
    <premis:linkingAgentIdentifier>
      <premis:linkingAgentIdentifierType valueURI="http://id.loc.gov/vocabulary/identifiers/local">url</premis:linkingAgentIdentifierType>
      <premis:linkingAgentIdentifierValue>https://github.com/artefactual-sdps/preprocessing-sfa</premis:linkingAgentIdentifierValue>
    </premis:linkingAgentIdentifier>
  </premis:event>
</premis:premis>
`

const expectedPREMISWithUnsuccessfulEvent = `<?xml version="1.0" encoding="UTF-8"?>
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
    <premis:linkingEventIdentifier>
      <premis:linkingEventIdentifierType></premis:linkingEventIdentifierType>
      <premis:linkingEventIdentifierValue></premis:linkingEventIdentifierValue>
    </premis:linkingEventIdentifier>
  </premis:object>
  <premis:event>
    <premis:eventIdentifier>
      <premis:eventIdentifierType></premis:eventIdentifierType>
      <premis:eventIdentifierValue></premis:eventIdentifierValue>
    </premis:eventIdentifier>
    <premis:eventType>someActivity</premis:eventType>
    <premis:eventDateTime></premis:eventDateTime>
    <premis:eventDetailInformation>
      <premis:eventDetail></premis:eventDetail>
    </premis:eventDetailInformation>
    <premis:eventOutcomeInformation>
      <premis:eventOutcome>invalid</premis:eventOutcome>
    </premis:eventOutcomeInformation>
    <premis:linkingAgentIdentifier>
      <premis:linkingAgentIdentifierType valueURI="http://id.loc.gov/vocabulary/identifiers/local">url</premis:linkingAgentIdentifierType>
      <premis:linkingAgentIdentifierValue>https://github.com/artefactual-sdps/preprocessing-sfa</premis:linkingAgentIdentifierValue>
    </premis:linkingAgentIdentifier>
  </premis:event>
</premis:premis>
`

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

func TestAddPREMISEvent(t *testing.T) {
	t.Parallel()

	// Creation of PREMIS file in non-existing directory (for execution expected to fail).
	transferDeleted := fs.NewDir(t, "",
		fs.WithDir("metadata"),
	)

	PREMISFilePathNonExistent := transferDeleted.Join("metadata", "premis.xml")

	transferDeleted.Remove()

	tests := []struct {
		name       string
		params     premis_event.Params
		result     premis_event.Result
		wantErr    string
		wantPREMIS string
	}{
		{
			name: "Add PREMIS event for normal content with no failures",
			params: premis_event.Params{
				PREMISFilePath: fs.NewFile(t, "premis.xml",
					fs.WithContent(expectedPREMISWithFile),
				).Path(),
				Agent: premis.AgentDefault(),
				Summary: premis.EventSummary{
					Type:    "someActivity",
					Outcome: "valid",
				},
			},
			result:     premis_event.Result{},
			wantPREMIS: expectedPREMISWithSuccessfulEvent,
		},
		{
			name: "Add PREMIS event for normal content with failures",
			params: premis_event.Params{
				PREMISFilePath: fs.NewFile(t, "premis.xml",
					fs.WithContent(expectedPREMISWithFile),
				).Path(),
				Agent: premis.AgentDefault(),
				Summary: premis.EventSummary{
					Type:    "someActivity",
					Outcome: "invalid",
				},
			},
			result:     premis_event.Result{},
			wantPREMIS: expectedPREMISWithUnsuccessfulEvent,
		},
		{
			name: "Add PREMIS event for no content",
			params: premis_event.Params{
				PREMISFilePath: fs.NewDir(t, "",
					fs.WithDir("metadata"),
				).Join("metadata", "premis.xml"),
				Agent: premis.AgentDefault(),
				Summary: premis.EventSummary{
					Type:    "someActivity",
					Outcome: "valid",
				},
			},
			result:     premis_event.Result{},
			wantPREMIS: premis.EmptyXML,
		},
		{
			name: "Add PREMIS event for bad path",
			params: premis_event.Params{
				PREMISFilePath: PREMISFilePathNonExistent,
				Agent:          premis.AgentDefault(),
				Summary: premis.EventSummary{
					Type:    "someActivity",
					Outcome: "valid",
				},
			},
			result:  premis_event.Result{},
			wantErr: "no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				premis_event.New().Execute,
				temporalsdk_activity.RegisterOptions{Name: premis_event.Name},
			)

			var res premis_event.Result
			future, err := env.ExecuteActivity(premis_event.Name, tt.params)

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

			// Compare PREMIS output to what's expected.
			if tt.wantPREMIS != "" {
				b, err := os.ReadFile(tt.params.PREMISFilePath)
				assert.NilError(t, err)
				assert.Equal(t, string(b), tt.wantPREMIS)
			}
		})
	}
}
