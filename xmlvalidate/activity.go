package xmlvalidate

import (
	"context"
	"fmt"
	"os"

	"go.artefactual.dev/tools/temporal"
)

const Name = "xml-validate"

// An XSDValidator validates the XML document at xmlPath against the XSD schema
// at xsdPath.
type XSDValidator interface {
	Validate(ctx context.Context, xmlPath, xsdPath string) (string, error)
}

type (
	Params struct {
		// XMLPath is the path of the file to be validated.
		XMLPath string

		// XSDPath is the path of the XSD file to use for validation.
		XSDPath string
	}
	Result struct {
		Failures []string
	}
	Activity struct {
		validator XSDValidator
	}
)

func New(validator XSDValidator) *Activity {
	return &Activity{validator: validator}
}

// Execute checks an XML file using the XSD file provided and returns error output.
func (a *Activity) Execute(ctx context.Context, params *Params) (*Result, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing xml-validate activity", "XMLPath", params.XMLPath)

	if _, err := os.Stat(params.XMLPath); err != nil {
		return nil, err
	}
	if _, err := os.Stat(params.XSDPath); err != nil {
		return nil, err
	}

	// Validate XML file against XSD.
	out, err := a.validator.Validate(ctx, params.XMLPath, params.XSDPath)
	if err != nil {
		return nil, fmt.Errorf("xmlvalidate: %w", err)
	}

	var failures []string
	if out != "" {
		failures = []string{out}
	}

	return &Result{Failures: failures}, nil
}
