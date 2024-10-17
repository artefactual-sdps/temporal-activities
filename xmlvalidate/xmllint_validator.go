package xmlvalidate

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
)

// XMLLintValidator represents an XSDValidator implementation using the xmllint
// C library.
type XMLLintValidator struct{}

var _ XSDValidator = &XMLLintValidator{}

func NewXMLLintValidator() *XMLLintValidator {
	return &XMLLintValidator{}
}

// Validate validates the XML document at xmlPath against the XSD schema at
// xsdPath using xmllint. If xmllint fails to run due to an error, then an error
// is returned. If xmllint runs and finds validation errors the errors will be
// returned as a string and the error return will be nil.
func (v *XMLLintValidator) Validate(ctx context.Context, xmlPath, xsdPath string) (string, error) {
	toolPath, err := exec.LookPath("xmllint")
	if err != nil {
		return "", err
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, toolPath, "--schema", xsdPath, xmlPath, "--noout") // #nosec G204
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		var x *exec.ExitError
		if errors.As(err, &x) {
			switch x.ExitCode() {

			// Exit code 2 is an error in the DTD.
			case 2:
				fallthrough

			// Exit codes 3 & 4 are validation errors.
			case 3, 4:
				fallthrough

			// Exit code 5 is an error in schema compilation.
			case 5:
				// Validation and DTD/schema errors are returned as a string
				// intended for an operator to resolve.
				return stderr.String(), nil

			// Other exit codes are not about the XML or DTD/schema being valid,
			// and are returned as an error intended for a system administrator
			// or developer to help debug the issue.
			default:
				return "", errors.New(stderr.String())
			}
		}

		return stderr.String(), nil
	}

	return "", nil
}
