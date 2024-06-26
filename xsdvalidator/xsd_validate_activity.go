package xsdvalidate

import (
	"context"
	"io/fs"
	"path/filepath"
	"regexp"

	"go.artefactual.dev/tools/temporal"
)

const XSDValidateActivityName = "xsd-validate-activity"

type XSDValidateActivity struct{}

func NewXSDValidateActivity() *XSDValidateActivity {
	return &XSDValidateActivity{}
}

type XSDValidateActivityParams struct {
	// DirectoryPath is the path of the files to be validated.
	DirectoryPath string

	// Pattern is the regular expression for files to validate.
	Pattern string

	// XSDFilePath is the path of the XSD file to use for validation.
	XSDFilePath string
}

type XSDValidateActivityResult struct {
	Failed []string
}

// Execute checks all files, in a given directory, matching a certain pattern.
func (a *XSDValidateActivity) Execute(
	ctx context.Context,
	params *XSDValidateActivityParams,
) (*XSDValidateActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing XMLValidateActivity", "DirectoryPath", params.DirectoryPath)

	/*
		dest, err := a.create(params.SourcePath, params.BagPath)
		if err != nil {
			return nil, fmt.Errorf("CreateBagActivity: %v", err)
		}
	*/

	return &XSDValidateActivityResult{}, nil
}

func MatchingFilesRelativeToDirectory(directoryPath string, pattern string) ([]string, error) {
	var subpaths []string

	// Compile regular expression pattern.
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(directoryPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories.
		if d.IsDir() {
			return nil
		}

		// Turn path into a subpath.
		subpath, err := filepath.Rel(directoryPath, p)
		if err != nil {
			return err
		}

		// If filename matches, then add to subpaths.
		if r.MatchString(filepath.Base(subpath)) {
			subpaths = append(subpaths, subpath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return subpaths, nil
}
