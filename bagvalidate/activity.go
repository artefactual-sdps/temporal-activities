package bagvalidate

import (
	"context"
	"errors"
	"fmt"

	"go.artefactual.dev/tools/temporal"
)

const Name = "bag-validate"

type (
	Params struct {
		// Path is the full path of the Bag to be validated.
		Path string
	}
	Result struct {
		// Valid is true if the Bag is valid.
		Valid bool

		// Error is a message indicating why validation failed, and will always be
		// empty when Valid is true.
		Error string
	}
	Activity struct {
		validator BagValidator
	}
)

// New creates a new bagvalidate activity.
// If the provided validator is nil, it defaults to using gythonValidator.
func New(validator BagValidator) *Activity {
	if validator == nil {
		validator = NewGythonValidator()
	}
	return &Activity{validator: validator}
}

// Execute validates the BagIt Bag located at Path.
//
// If validation succeeds Execute returns `&ValidateActivityResult{Valid: true},
// nil`.
// If validation fails Execute returns `&ValidateActivityResult{Valid: false,
// Error: "message"}, nil`.
// If an application error occurs Execute returns `nil, error("message")`
func (a *Activity) Execute(ctx context.Context, params *Params) (*Result, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing bag-validate activity", "Path", params.Path)

	if err := a.validator.Validate(params.Path); err != nil {
		if errors.Is(convertError(err), ErrInvalid) {
			return &Result{
				Valid: false,
				Error: err.Error(),
			}, nil
		}

		return nil, fmt.Errorf("bagvalidate: %v", err)
	}

	return &Result{Valid: true}, nil
}
