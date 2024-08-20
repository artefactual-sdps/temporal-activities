package removepaths

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.artefactual.dev/tools/temporal"
)

const Name = "remove-paths"

type (
	Params struct {
		// Paths is the list of file paths to remove.
		Paths []string
	}
	Result   struct{}
	Activity struct{}
)

func New() *Activity {
	return &Activity{}
}

// Execute removes all param.Paths any children they contain and returns any
// errors encountered. If a given path doesn't exist, no error will be returned.
func (a *Activity) Execute(ctx context.Context, params *Params) (*Result, error) {
	logger := temporal.GetLogger(ctx)
	logger.V(1).Info("Executing remove-paths", "Paths", params.Paths)

	var errs error
	for _, p := range params.Paths {
		if err := os.RemoveAll(p); err != nil {
			errs = errors.Join(errs, fmt.Errorf("removepaths: couldn't remove path: %v", err))
		}
	}

	return &Result{}, errs
}
