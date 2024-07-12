package bagit

type noopValidator struct{}

func (v noopValidator) Validate(path string) error {
	return nil
}

func NewNoopValidator() noopValidator {
	return noopValidator{}
}

var _ BagValidator = noopValidator{}
