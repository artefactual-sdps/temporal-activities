package xmlvalidate_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	temporalsdk_activity "go.temporal.io/sdk/activity"
	temporalsdk_testsuite "go.temporal.io/sdk/testsuite"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
)

type fakeValidator struct {
	Msg string
	Err error
}

var _ xmlvalidate.XSDValidator = &fakeValidator{}

func (v fakeValidator) Validate(ctx context.Context, xmlPath, xsdPath string) (string, error) {
	return v.Msg, v.Err
}

func newFakeValidator(msg string, err error) fakeValidator {
	return fakeValidator{Msg: msg, Err: err}
}

func TestActivity(t *testing.T) {
	t.Parallel()

	type test struct {
		name      string
		params    xmlvalidate.Params
		validator xmlvalidate.XSDValidator
		want      xmlvalidate.Result
		wantErr   string
	}
	for _, tt := range []test{
		{
			name: "Test XSD validate activity with a valid XML file and a valid XSD file",
			params: xmlvalidate.Params{
				XMLPath: filepath.Join("testdata", "person_valid.xml"),
				XSDPath: filepath.Join("testdata", "person.xsd"),
			},
			validator: newFakeValidator("", nil),
			want:      xmlvalidate.Result{},
		},
		{
			name: "Test XSD validate activity with an invalid XML file and a valid XSD file ",
			params: xmlvalidate.Params{
				XMLPath: filepath.Join("testdata", "person_invalid.xml"),
				XSDPath: filepath.Join("testdata", "person.xsd"),
			},
			validator: newFakeValidator(
				`person_invalid.xml:3: element age: Schemas validity error : Element 'age': This element is not expected. Expected is ( name )"
person_invalid.xml fails to validate
`,
				nil,
			),
			want: xmlvalidate.Result{Failures: []string{
				`person_invalid.xml:3: element age: Schemas validity error : Element 'age': This element is not expected. Expected is ( name )"
person_invalid.xml fails to validate
`,
			}},
		},
		{
			name: "Test XSD validate activity with a valid XML file and an invalid XSD file",
			params: xmlvalidate.Params{
				XMLPath: filepath.Join("testdata", "person_valid.xml"),
				XSDPath: filepath.Join("testdata", "invalid.xsd"),
			},
			validator: newFakeValidator(
				`invalid.xsd:1: parser error : Start tag expected, '<' not found"
junk
^
Schemas parser error : Failed to parse the XML resource 'invalid.xsd'.
WXS schema invalid.xsd failed to compile
`,
				nil,
			),
			want: xmlvalidate.Result{Failures: []string{
				`invalid.xsd:1: parser error : Start tag expected, '<' not found"
junk
^
Schemas parser error : Failed to parse the XML resource 'invalid.xsd'.
WXS schema invalid.xsd failed to compile
`,
			}},
		},
		{
			name: "Test XSD validate activity with a non-existent XML file and a valid XSD file",
			params: xmlvalidate.Params{
				XMLPath: "not_here.xml",
				XSDPath: filepath.Join("testdata", "person.xsd"),
			},
			validator: newFakeValidator("", nil),
			wantErr:   "no such file or directory",
		},
		{
			name: "Test XSD validate activity with a valid XML file and a non-existent XSD file",
			params: xmlvalidate.Params{
				XMLPath: filepath.Join("testdata", "person_valid.xml"),
				XSDPath: "not_here.xsd",
			},
			validator: newFakeValidator("", nil),
			wantErr:   "no such file or directory",
		},
		{
			name:      "Errors if validator errors",
			validator: newFakeValidator("", errors.New("error!")),
			wantErr:   "activity error (type: xml-validate, scheduledEventID: 0, startedEventID: 0, identity: ): stat : no such file or directory (type: PathError, retryable: true): no such file or directory (type: Errno, retryable: true)",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := &temporalsdk_testsuite.WorkflowTestSuite{}
			env := ts.NewTestActivityEnvironment()
			env.RegisterActivityWithOptions(
				xmlvalidate.New(tt.validator).Execute,
				temporalsdk_activity.RegisterOptions{Name: xmlvalidate.Name},
			)

			var res xmlvalidate.Result
			future, err := env.ExecuteActivity(xmlvalidate.Name, tt.params)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			future.Get(&res)
			assert.DeepEqual(t, res, tt.want)
			assert.NilError(t, err)
		})
	}
}
