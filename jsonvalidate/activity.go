package jsonvalidate

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.artefactual.dev/tools/temporal"
)

const Name = "json-validate"

type Params struct {
	JSONPath   string
	SchemaPath string
}

type Result struct {
	Failures []string
}
type Activity struct {
	validator JSONValidator
}

func New(validator JSONValidator) *Activity {
	return &Activity{validator: validator}
}

// Execute checks a JSON document against the JSON Schema provided and returns
// validation output. A validator implementation must be supplied to New.
func (a *Activity) Execute(ctx context.Context, params *Params) (*Result, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing json-validate activity")

	if params == nil || params.JSONPath == "" {
		return nil, temporal.NewNonRetryableError(errors.New("JSON path cannot be empty"))
	}
	if params.SchemaPath == "" {
		return nil, temporal.NewNonRetryableError(errors.New("JSON schema path cannot be empty"))
	}

	if err := errors.Join(
		fileExists(params.JSONPath),
		fileExists(params.SchemaPath),
	); err != nil {
		return nil, temporal.NewNonRetryableError(err)
	}

	out, err := a.validator.Validate(ctx, params.JSONPath, params.SchemaPath)
	if err != nil {
		return nil, fmt.Errorf("jsonvalidate: %w", err)
	}

	return &Result{Failures: out}, nil
}

func fileExists(path string) error {
	_, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		err = errors.New(path + " is missing")
	}
	return err
}
