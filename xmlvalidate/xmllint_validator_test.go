package xmlvalidate_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
)

// TestXMLLintValidate tests the XMLLintValidator. The tests require xmllint to
// be installed in the host system, so the SDPS_TEST_XMLLINT environment
// variable is used to conditionally run the test, e.g.
// `SDPS_TEST_XMLLINT=1 go test -tags xmllint ./...`.
func TestXMLLintValidate(t *testing.T) {
	t.Parallel()

	do := os.Getenv("SDPS_TEST_XMLLINT")
	if do == "" {
		t.Skip("set SDPS_TEST_XMLLINT to run this test")
	}

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
			want: "Element 'age': This element is not expected. Expected is ( name ).",
		},
		{
			name: "returns a validation error when XSD is invalid",
			args: args{
				xmlPath: "./testdata/person_valid.xml",
				xsdPath: "./testdata/invalid.xsd",
			},
			want: "parser error : Start tag expected, '<' not found",
		},
		{
			name: "returns an error when XML doesn't exist",
			args: args{
				xmlPath: "./testdata/not_here.xml",
				xsdPath: "./testdata/person.xsd",
			},
			wantErr: "XML file does not exist: ./testdata/not_here.xml",
		},
		{
			name: "returns an error when XSD doesn't exist",
			args: args{
				xmlPath: "./testdata/person_valid.xml",
				xsdPath: "./testdata/not_here.xsd",
			},
			wantErr: "XSD file does not exist: ./testdata/not_here.xsd",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := xmlvalidate.NewXMLLintValidator()
			got, err := v.Validate(context.Background(), tt.args.xmlPath, tt.args.xsdPath)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)

			assert.Assert(t, strings.Contains(got, tt.want), "unexpected result: want %q, got %q", tt.want, got)
		})
	}
}
