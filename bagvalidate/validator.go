package bagvalidate

import (
	"fmt"
	"os"
	"path/filepath"

	gobagit "github.com/nyudlts/go-bagit"
)

type BagValidator interface {
	Validate(path string) error
}

type noopValidator struct{}

func (v noopValidator) Validate(path string) error {
	return nil
}

type validator struct{}

func (v validator) Validate(path string) error {
	// Check if path is a bag.
	if _, err := os.Stat(filepath.Join(path, "bagit.txt")); err != nil {
		// Do nothing if not a bag (bagit.txt doesn't exist).
		return fmt.Errorf("%w: %s", ErrNotABag, "bagit.txt not found")
	}

	// Validate bag.
	b, err := gobagit.GetExistingBag(path)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalid, RemovePathFromError(path, err))
	}

	err = b.ValidateBag(true, false)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalid, RemovePathFromError(path, err))
	}

	return nil
}

func NewNoopValidator() noopValidator {
	return noopValidator{}
}

func NewValidator() validator {
	return validator{}
}

var _ BagValidator = validator{}
