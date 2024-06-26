package xml

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"go.artefactual.dev/tools/temporal"
)

const XSDValidateActivityName = "xsd-validate-activity"

type XSDValidateActivity struct{}

func NewXSDValidateActivity() *XSDValidateActivity {
	return &XSDValidateActivity{}
}

type XSDValidateActivityParams struct {
	// XMLFilePath is the path of the file to be validated.
	XMLFilePath string

	// XSDFilePath is the path of the XSD file to use for validation.
	XSDFilePath string
}

type XSDValidateActivityResult struct {
	Failures []byte
}

// Execute checks an XML file using the XSD file provided and returns error output.
func (a *XSDValidateActivity) Execute(
	ctx context.Context,
	params *XSDValidateActivityParams,
) (*XSDValidateActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing XMLValidateActivity", "XMLFilePath", params.XMLFilePath)

	// Check XML file using XSD.
	lintErrors, err := checkXML(ctx, params.XMLFilePath, params.XSDFilePath)
	if err != nil {
		return nil, err
	}

	return &XSDValidateActivityResult{Failures: lintErrors}, nil
}

func checkXML(ctx context.Context, xmlFilePath string, xsdFilePath string) ([]byte, error) {
	toolFilePath, err := exec.LookPath("xmllint")
	if err != nil {
		return nil, err
	}

	xsdFilePath = filepath.Clean(xsdFilePath)
	_, err = os.Stat(xsdFilePath)
	if err != nil {
		return nil, err
	}

	xmlFilePath = filepath.Clean(xmlFilePath)
	_, err = os.Stat(xmlFilePath)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, toolFilePath, "--schema", xsdFilePath, xmlFilePath, "--noout") // #nosec G204

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return stderr.Bytes(), nil
	}

	return nil, err
}
