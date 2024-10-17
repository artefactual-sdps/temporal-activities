package xmlvalidate_test

import (
	"context"
	"os"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
)

// TestXMLLintValidate tests the XMLLintValidator. The tests require xmllint to
// be installed in the host system, so the SDPS_TEST_XMLLINT environment
// variable is used to conditionally run the test, e.g.
// `SDPS_TEST_XMLLINT=1 go test -tags xmllint ./...`.
func TestXMLLintValidate(t *testing.T) {
	do := os.Getenv("SDPS_TEST_XMLLINT")
	if do == "" {
		t.Skip("set SDPS_TEST_XMLLINT to run this test")
	}

	t.Parallel()

	type args struct {
		xmlPath string
		xsdPath string
	}
	type test struct {
		name    string
		args    args
		want    string
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "returns nothing when XML is valid",
			args: args{
				xmlPath: "./testdata/person_valid.xml",
				xsdPath: "./testdata/person.xsd",
			},
		},
		{
			name: "returns a validation error when XML is invalid",
			args: args{
				xmlPath: "./testdata/person_invalid.xml",
				xsdPath: "./testdata/person.xsd",
			},
			want: `./testdata/person_invalid.xml:3: element age: Schemas validity error : Element 'age': This element is not expected. Expected is ( name ).
./testdata/person_invalid.xml fails to validate
`,
		},
		{
			name: "returns a validation error when XSD is invalid",
			args: args{
				xmlPath: "./testdata/person_valid.xml",
				xsdPath: "./testdata/invalid.xsd",
			},
			want: `./testdata/invalid.xsd:1: parser error : Start tag expected, '<' not found
junk
^
Schemas parser error : Failed to parse the XML resource './testdata/invalid.xsd'.
WXS schema ./testdata/invalid.xsd failed to compile
`,
		},
		{
			name: "returns a validation error when XML doesn't exist",
			args: args{
				xmlPath: "./testdata/not_here.xml",
				xsdPath: "./testdata/person.xsd",
			},
			wantErr: "warning: failed to load external entity \"./testdata/not_here.xml\"\n",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := xmlvalidate.NewXMLLintValidator()
			got, err := v.Validate(context.Background(), tt.args.xmlPath, tt.args.xsdPath)

			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.Equal(t, got, tt.want)
		})
	}
}
