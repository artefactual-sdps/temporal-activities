package bagvalidate

import (
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
	b, err := gobagit.GetExistingBag(path)
	if err != nil {
		return err
	}

	// Validate bag.
	err = b.ValidateBag(true, false)
	if err != nil {
		return err
	}

	return nil
}

func NewNoopValidator() noopValidator {
	return noopValidator{}
}

func NewValidator() validator {
	return validator{}
}

var _ BagValidator = noopValidator{}
