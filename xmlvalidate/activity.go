package xmlvalidate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"go.artefactual.dev/tools/temporal"
)

const Name = "xml-validate"

type (
	Params struct {
		// XMLFilePath is the path of the file to be validated.
		XMLFilePath string

		// XSDFilePath is the path of the XSD file to use for validation.
		XSDFilePath string
	}
	Result struct {
		Failures []byte
	}
	Activity struct{}
)

func New() *Activity {
	return &Activity{}
}

// Execute checks an XML file using the XSD file provided and returns error output.
func (a *Activity) Execute(ctx context.Context, params *Params) (*Result, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing xml-validate activity", "XMLFilePath", params.XMLFilePath)

	// Check XML file using XSD.
	lintErrors, err := checkXML(ctx, params.XMLFilePath, params.XSDFilePath)
	if err != nil {
		return nil, fmt.Errorf("xmlvalidate: %w", err)
	}

	return &Result{Failures: lintErrors}, nil
}

func checkXML(ctx context.Context, xmlFilePath, xsdFilePath string) ([]byte, error) {
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
