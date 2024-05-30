package filesys

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.artefactual.dev/tools/temporal"
)

const RemoveActivityName = "filesys-remove-activity"

type RemoveActivity struct{}

func NewRemoveActivity() *RemoveActivity {
	return &RemoveActivity{}
}

type RemoveActivityParams struct {
	// Paths is the list of file paths to remove.
	Paths []string
}

type RemoveActivityResult struct{}

// Execute removes all param.Paths any children they contain and returns any
// errors encountered. If a given path doesn't exist, no error will be returned.
func (a *RemoveActivity) Execute(ctx context.Context, params *RemoveActivityParams) (*RemoveActivityResult, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing filepath.RemoveActivity", "Paths", params.Paths)

	var errs error
	for _, p := range params.Paths {
		if err := os.RemoveAll(p); err != nil {
			errs = errors.Join(errs, fmt.Errorf("couldn't remove path: %v", err))
		}
	}

	return &RemoveActivityResult{}, errs
}
