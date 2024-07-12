package bagit

import (
	"context"

	"go.artefactual.dev/tools/temporal"
)

const ValidateActivityName = "validate-bag-activity"

type BagValidator interface {
	Validate(path string) error
}

type ValidateActivity struct {
	validator BagValidator
}

func NewValidateActivity(validator BagValidator) *ValidateActivity {
	return &ValidateActivity{validator: validator}
}

type ValidateActivityParams struct {
	// Path is the full path of the Bag to be validated.
	Path string
}

type ValidateActivityResult struct {
	// Valid is true if the Bag is valid.
	Valid bool

	// Error is a message indicating why validation failed, and will always be
	// empty when Valid is true.
	Error string
}

// Execute validates the BagIt Bag located at Path.
func (a *ValidateActivity) Execute(
	ctx context.Context,
	params *ValidateActivityParams,
) (*ValidateActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing BagIt Validate Activity", "Path", params.Path)

	if err := a.validator.Validate(params.Path); err != nil {
		return &ValidateActivityResult{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &ValidateActivityResult{Valid: true}, nil
}
