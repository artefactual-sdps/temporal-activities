package xml_test

import (
	"fmt"
	"path/filepath"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"
	tfs "gotest.tools/v3/fs"

	"github.com/artefactual-sdps/temporal-activities/xml"
)

func TestXSDValidateActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		params  func() xml.XSDValidateActivityParams
		want    func(params xml.XSDValidateActivityParams) xml.XSDValidateActivityResult
		wantErr string
	}
	for _, tt := range []test{
		{
			name: "Test XSD validate activity with a valid XML file and a valid XSD file",
			params: func() xml.XSDValidateActivityParams {
				validXMLFilePath := filepath.Join("testdata", "person_valid.xml")
				XSDTestFilePath := filepath.Join("testdata", "person.xsd")

				return xml.XSDValidateActivityParams{
					XMLFilePath: validXMLFilePath,
					XSDFilePath: XSDTestFilePath,
				}
			},
			want: func(params xml.XSDValidateActivityParams) xml.XSDValidateActivityResult {
				return xml.XSDValidateActivityResult{Failures: nil}
			},
		},
		{
			name: "Test XSD validate activity with an invalid XML file and a valid XSD file ",
			params: func() xml.XSDValidateActivityParams {
				invalidXMLFilePath := filepath.Join("testdata", "person_invalid.xml")
				XSDTestFilePath := filepath.Join("testdata", "person.xsd")

				return xml.XSDValidateActivityParams{
					XMLFilePath: invalidXMLFilePath,
					XSDFilePath: XSDTestFilePath,
				}
			},
			want: func(params xml.XSDValidateActivityParams) xml.XSDValidateActivityResult {
				invalidXMLFailures := fmt.Sprintf(
					"%s:3: element age: Schemas validity error : Element 'age': This element is not expected. Expected is ( name ).\n"+
						"%s fails to validate\n",
					params.XMLFilePath,
					params.XMLFilePath,
				)

				return xml.XSDValidateActivityResult{Failures: []byte(invalidXMLFailures)}
			},
		},
		{
			name: "Test XSD validate activity with a valid XML file and an invalid XSD file",
			params: func() xml.XSDValidateActivityParams {
				validXMLFilePath := filepath.Join("testdata", "person_valid.xml")
				invalidXSDFilePath := filepath.Join("testdata", "invalid.xsd")

				return xml.XSDValidateActivityParams{
					XMLFilePath: validXMLFilePath,
					XSDFilePath: invalidXSDFilePath,
				}
			},
			want: func(params xml.XSDValidateActivityParams) xml.XSDValidateActivityResult {
				invalidXSDFailures := fmt.Sprintf(
					"%s:1: parser error : Start tag expected, '<' not found\n"+
						"junk\n"+
						"^\n"+
						"Schemas parser error : Failed to parse the XML resource '%s'.\n"+
						"WXS schema %s failed to compile\n",
					params.XSDFilePath,
					params.XSDFilePath,
					params.XSDFilePath,
				)

				return xml.XSDValidateActivityResult{Failures: []byte(invalidXSDFailures)}
			},
		},
		{
			name: "Test XSD validate activity with a non-existent XML file and a valid XSD file",
			params: func() xml.XSDValidateActivityParams {
				nonExistentFile := tfs.NewFile(t, "removed_file")
				nonExistentFile.Remove()

				XSDTestFilePath := filepath.Join("testdata", "person.xsd")

				return xml.XSDValidateActivityParams{
					XMLFilePath: nonExistentFile.Path(),
					XSDFilePath: XSDTestFilePath,
				}
			},
			want: func(params xml.XSDValidateActivityParams) xml.XSDValidateActivityResult {
				return xml.XSDValidateActivityResult{}
			},
			wantErr: "no such file or directory",
		},
		{
			name: "Test XSD validate activity with a valid XML file and a non-existent XSD file",
			params: func() xml.XSDValidateActivityParams {
				validXMLFilePath := filepath.Join("testdata", "person_valid.xml")

				nonExistentFile := tfs.NewFile(t, "removed_file")
				nonExistentFile.Remove()

				return xml.XSDValidateActivityParams{
					XMLFilePath: validXMLFilePath,
					XSDFilePath: nonExistentFile.Path(),
				}
			},
			want: func(params xml.XSDValidateActivityParams) xml.XSDValidateActivityResult {
				return xml.XSDValidateActivityResult{}
			},
			wantErr: "no such file or directory",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				xml.NewXSDValidateActivity().Execute,
				temporalsdk_activity.RegisterOptions{Name: xml.XSDValidateActivityName},
			)

			var res xml.XSDValidateActivityResult
			future, err := env.ExecuteActivity(xml.XSDValidateActivityName, tt.params())
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			future.Get(&res)
			assert.DeepEqual(t, res, tt.want(tt.params()))
			assert.NilError(t, err)
		})
	}
}
