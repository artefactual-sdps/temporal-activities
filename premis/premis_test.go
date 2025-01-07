package premis_test

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/premis"
)

const premisObjectAddContent = `<?xml version="1.0" encoding="UTF-8"?>
<premis:premis xmlns:premis="http://www.loc.gov/premis/v3" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.loc.gov/premis/v3 https://www.loc.gov/standards/premis/premis.xsd" version="3.0">
  <premis:object xsi:type="premis:file">
    <premis:objectIdentifier>
      <premis:objectIdentifierType>UUID</premis:objectIdentifierType>
      <premis:objectIdentifierValue>c74a85b7-919b-409e-8209-9c7ebe0e7945</premis:objectIdentifierValue>
    </premis:objectIdentifier>
    <premis:objectCharacteristics>
      <premis:format>
        <premis:formatDesignation>
          <premis:formatName></premis:formatName>
        </premis:formatDesignation>
      </premis:format>
    </premis:objectCharacteristics>
    <premis:originalName>data/objects/test_transfer/content/cat.jpg</premis:originalName>
  </premis:object>
</premis:premis>
`

const premisObjectAndEventAddContent = `<?xml version="1.0" encoding="UTF-8"?>
<premis:premis xmlns:premis="http://www.loc.gov/premis/v3" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.loc.gov/premis/v3 https://www.loc.gov/standards/premis/premis.xsd" version="3.0">
  <premis:object xsi:type="premis:file">
    <premis:objectIdentifier>
      <premis:objectIdentifierType>UUID</premis:objectIdentifierType>
      <premis:objectIdentifierValue>d14db00a-8d4d-4057-8661-cd0f70b670eb</premis:objectIdentifierValue>
    </premis:objectIdentifier>
    <premis:objectCharacteristics>
      <premis:format>
        <premis:formatDesignation>
          <premis:formatName></premis:formatName>
        </premis:formatDesignation>
      </premis:format>
    </premis:objectCharacteristics>
    <premis:originalName>data/objects/test_transfer/content/cat.jpg</premis:originalName>
    <premis:linkingEventIdentifier>
      <premis:linkingEventIdentifierType>UUID</premis:linkingEventIdentifierType>
      <premis:linkingEventIdentifierValue>a3207f0b-3e09-4535-949f-d15a82972ac9</premis:linkingEventIdentifierValue>
    </premis:linkingEventIdentifier>
  </premis:object>
  <premis:event>
    <premis:eventIdentifier>
      <premis:eventIdentifierType>UUID</premis:eventIdentifierType>
      <premis:eventIdentifierValue>a3207f0b-3e09-4535-949f-d15a82972ac9</premis:eventIdentifierValue>
    </premis:eventIdentifier>
    <premis:eventType>validation</premis:eventType>
    <premis:eventDateTime>2024-12-03T09:51:07-08:00</premis:eventDateTime>
    <premis:eventDetailInformation>
      <premis:eventDetail>name=&quot;Validate SIP metadata&quot;</premis:eventDetail>
    </premis:eventDetailInformation>
    <premis:eventOutcomeInformation>
      <premis:eventOutcome>invalid</premis:eventOutcome>
      <premis:eventOutcomeDetail>
        <premis:eventOutcomeDetailNote>Metadata validation successful</premis:eventOutcomeDetailNote>
      </premis:eventOutcomeDetail>
    </premis:eventOutcomeInformation>
    <premis:linkingAgentIdentifier>
      <premis:linkingAgentIdentifierType valueURI="http://id.loc.gov/vocabulary/identifiers/local">url</premis:linkingAgentIdentifierType>
      <premis:linkingAgentIdentifierValue>https://github.com/artefactual-sdps/preprocessing-sfa</premis:linkingAgentIdentifierValue>
    </premis:linkingAgentIdentifier>
  </premis:event>
</premis:premis>
`

const premisAgentAddContent = `<?xml version="1.0" encoding="UTF-8"?>
<premis:premis xmlns:premis="http://www.loc.gov/premis/v3" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.loc.gov/premis/v3 https://www.loc.gov/standards/premis/premis.xsd" version="3.0">
  <premis:agent>
    <premis:agentIdentifier>
      <premis:agentIdentifierType valueURI="http://id.loc.gov/vocabulary/identifiers/local">url</premis:agentIdentifierType>
      <premis:agentIdentifierValue>https://github.com/artefactual-sdps/preprocessing-sfa</premis:agentIdentifierValue>
    </premis:agentIdentifier>
    <premis:agentName>Enduro</premis:agentName>
    <premis:agentType>software</premis:agentType>
  </premis:agent>
</premis:premis>
`

func TestParseFile(t *testing.T) {
	t.Parallel()

	td := fs.NewDir(t, "", fs.WithFile(
		"agent.xml",
		premisAgentAddContent,
	))

	doc, err := premis.ParseFile(td.Join("agent.xml"))
	assert.NilError(t, err)

	got, err := doc.WriteToString()
	assert.NilError(t, err)
	assert.Equal(t, got, premisAgentAddContent)
}

func TestParseOrInitialize(t *testing.T) {
	t.Parallel()

	t.Run("Parses an existing XML file", func(t *testing.T) {
		t.Parallel()

		td := fs.NewDir(t, "", fs.WithFile(
			"agent.xml",
			premisAgentAddContent,
		))

		doc, err := premis.ParseOrInitialize(td.Join("agent.xml"))
		assert.NilError(t, err)

		got, err := doc.WriteToString()
		assert.NilError(t, err)
		assert.Equal(t, got, premisAgentAddContent)
	})

	t.Run("Creates an empty PREMIS XML file", func(t *testing.T) {
		t.Parallel()

		td := fs.NewDir(t, "")

		doc, err := premis.ParseOrInitialize(td.Join("test.xml"))
		assert.NilError(t, err)

		got, err := doc.WriteToString()
		assert.NilError(t, err)
		assert.Equal(
			t,
			got,
			`<?xml version="1.0" encoding="UTF-8"?>
<premis:premis xmlns:premis="http://www.loc.gov/premis/v3" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.loc.gov/premis/v3 https://www.loc.gov/standards/premis/premis.xsd" version="3.0"></premis:premis>
`)
	})
}

func TestAppendPREMISObjectXML(t *testing.T) {
	t.Parallel()

	// Test with PREMIS object.
	doc, err := premis.NewDoc()
	assert.NilError(t, err)

	err = premis.AppendObjectXML(doc, premis.Object{
		IdType:       "UUID",
		IdValue:      "c74a85b7-919b-409e-8209-9c7ebe0e7945",
		OriginalName: "data/objects/test_transfer/content/cat.jpg",
	})
	assert.NilError(t, err)

	// Get resulting XML string.
	xml, err := premis.WriteIndentedToString(doc)
	assert.NilError(t, err)

	// Compare XML to constant.
	assert.Equal(t, xml, premisObjectAddContent)
}

func TestAppendPREMISEventXML(t *testing.T) {
	t.Parallel()

	// Add test PREMIS object.
	doc, err := premis.NewDoc()
	assert.NilError(t, err)

	err = premis.AppendObjectXML(doc, premis.Object{
		IdType:       "UUID",
		IdValue:      "d14db00a-8d4d-4057-8661-cd0f70b670eb",
		OriginalName: "data/objects/test_transfer/content/cat.jpg",
	})
	assert.NilError(t, err)

	// Test adding PREMIS event.
	err = premis.AppendEventXMLForEachObject(doc, premis.EventSummary{
		IdType:        "UUID",
		IdValue:       "a3207f0b-3e09-4535-949f-d15a82972ac9",
		DateTime:      "2024-12-03T09:51:07-08:00",
		Type:          "validation",
		Detail:        "name=\"Validate SIP metadata\"",
		Outcome:       "invalid",
		OutcomeDetail: "Metadata validation successful",
	}, premis.AgentDefault())
	assert.NilError(t, err)

	// Get resulting XML string.
	xml, err := premis.WriteIndentedToString(doc)
	assert.NilError(t, err)

	// Compare XML to constant.
	assert.Equal(t, xml, premisObjectAndEventAddContent)
}

func TestAppendPREMISAgentXML(t *testing.T) {
	t.Parallel()

	// Test with PREMIS agent.
	doc, err := premis.NewDoc()
	assert.NilError(t, err)

	err = premis.AppendAgentXML(doc, premis.AgentDefault())
	assert.NilError(t, err)

	// Get resulting XML string.
	xml, err := premis.WriteIndentedToString(doc)
	assert.NilError(t, err)

	// Compare XML to constant.
	assert.Equal(t, xml, premisAgentAddContent)

	// Try to add another PREMIS agent to existing XML document.
	err = premis.AppendAgentXML(doc, premis.AgentDefault())
	assert.NilError(t, err)

	// Get resulting XML string.
	xml, err = premis.WriteIndentedToString(doc)
	assert.NilError(t, err)

	// Compare XML to constant to make sure a duplicate agent wasn't added.
	assert.Equal(t, xml, premisAgentAddContent)
}

func TestFilesWithinDirectory(t *testing.T) {
	t.Parallel()

	contentPath := fs.NewDir(t, "",
		fs.WithDir("content",
			fs.WithDir("content",
				fs.WithDir("d_0000001",
					fs.WithFile("00000001.jp2", ""),
					fs.WithFile("00000001_PREMIS.xml", ""),
					fs.WithFile("Prozess_Digitalisierung_PREMIS.xml", ""),
				),
			),
		),
	).Path()

	expectedFiles := []string{
		"content/content/d_0000001/00000001.jp2",
		"content/content/d_0000001/00000001_PREMIS.xml",
		"content/content/d_0000001/Prozess_Digitalisierung_PREMIS.xml",
	}

	files, err := premis.FilesWithinDirectory(contentPath)
	assert.NilError(t, err)

	assert.DeepEqual(t, files, expectedFiles)
}
