package xmlvalidate

import (
	"bytes"
	"context"
	"os/exec"
)

// An XSDValidator validates the XML document at xmlPath against the XSD schema
// at xsdPath.
type XSDValidator interface {
	Validate(ctx context.Context, xmlPath, xsdPath string) (string, error)
}

// XMLLintValidator represents an XSDValidator implementation using the xmllint
// C library.
type XMLLintValidator struct{}

var _ XSDValidator = &XMLLintValidator{}

func NewXMLLintValidator() *XMLLintValidator {
	return &XMLLintValidator{}
}

// Validate validates the XML document at xmlPath against the XSD schema at
// xsdPath using xmllint.
func (v *XMLLintValidator) Validate(ctx context.Context, xmlPath, xsdPath string) (string, error) {
	toolPath, err := exec.LookPath("xmllint")
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, toolPath, "--schema", xsdPath, xmlPath, "--noout") // #nosec G204

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stderr.String(), nil
	}

	return "", nil
}
