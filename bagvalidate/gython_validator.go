package bagvalidate

import (
	"fmt"

	bagit_gython "github.com/artefactual-labs/bagit-gython"
)

// gythonValidator wraps bagit_gython.BagIt ensuring a new instance
// is created and cleaned for each validation.
type gythonValidator struct{}

// NewGythonValidator creates a new instance of gythonValidator.
func NewGythonValidator() gythonValidator {
	return gythonValidator{}
}

// Validate creates a new bagit_gython.BagIt validator, runs validation
// on the given path, and calls Cleanup before returning.
func (gv gythonValidator) Validate(path string) error {
	v, err := bagit_gython.NewBagIt()
	if err != nil {
		return fmt.Errorf("failed to create gython validator: %v", err)
	}

	defer func() {
		// Ignore cleanup error to avoid mixing it with validation errors.
		_ = v.Cleanup()
	}()

	return v.Validate(path)
}

var _ BagValidator = gythonValidator{}
