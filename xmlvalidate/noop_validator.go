package xmlvalidate

import "context"

// NoopValidator is an implementation of XSDValidator that always returns an
// empty (passing) result.  NoopValidator is used for testing or other scenarios
// where a functional validator is not required.
type NoopValidator struct{}

// Validate always returns empty (passing) results.
func (v NoopValidator) Validate(ctx context.Context, xmlPath, xsdPath string) (string, error) {
	return "", nil
}

func NewNoopValidator() *NoopValidator {
	return &NoopValidator{}
}

var _ XSDValidator = NoopValidator{}
