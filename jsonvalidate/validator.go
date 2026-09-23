package jsonvalidate

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// If an error is populated it is assumed some system error occurred, for validation
// errors check the validationErrors slice.
type JSONValidator interface {
	Validate(ctx context.Context, jsonPath, schemaPath string) (validationErrors []string, err error)
}

type ValidatorImpl struct{}

func NewValidator() *ValidatorImpl {
	return &ValidatorImpl{}
}

func (v *ValidatorImpl) Validate(ctx context.Context, jsonPath, schemaPath string) ([]string, error) {
	schema, err := jsonschema.NewCompiler().Compile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema %q: %w", schemaPath, err)
	}

	jsonContents, err := loadJSONContents(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file %q: %w", jsonPath, err)
	}

	err = schema.Validate(jsonContents)
	if err == nil {
		// Validation successful, return nothing.
		return nil, nil
	}

	validationErrors := []string{}
	if schemaErr, ok := errors.AsType[*jsonschema.ValidationError](err); ok {
		for _, e := range schemaErr.BasicOutput().Errors {
			var msg string
			if e.Error != nil {
				msg = e.Error.String()
			}
			if e.InstanceLocation != "" {
				msg = fmt.Sprintf("%s: %s", e.InstanceLocation, msg)
			}
			validationErrors = append(validationErrors, msg)
		}
	} else {
		return nil, fmt.Errorf("unexpected validation error: %w", err)
	}

	return validationErrors, nil
}

func loadJSONContents(path string) (any, error) {
	f, err := os.Open(path) // #nosec G304 -- trusted path.
	if err != nil {
		return nil, fmt.Errorf("open json %q: %w", path, err)
	}
	defer f.Close()

	doc, err := jsonschema.UnmarshalJSON(f)
	if err != nil {
		return nil, fmt.Errorf("decode json %q: %w", path, err)
	}
	return doc, nil
}
