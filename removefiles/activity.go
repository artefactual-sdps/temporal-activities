package removefiles

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

const ActivityName = "remove-files-activity"

type Activity struct {
	logger logr.Logger
}

func NewActivity(logger logr.Logger) *Activity {
	return &Activity{logger: logger}
}

type ActivityParams struct {
	// RemoveNames is the comma separated list of filenames that should be
	// removed from Path.
	RemoveNames string

	// Path is the directory from which files should be removed.
	Path string
}

type ActivityResult struct {
	// Count is the number of files removed from Path.
	Count int
}

// Execute deletes any files named by RemoveNames from the given Path and
// returns a count of deleted files.
func (a *Activity) Execute(
	ctx context.Context,
	params *ActivityParams,
) (*ActivityResult, error) {
	a.logger.V(2).Info("Executing RemoveFilesActivity",
		"RemoveNames", params.RemoveNames,
		"Path", params.Path,
	)

	count, err := Remove(params.Path, params.RemoveNames)
	if err != nil {
		return nil, fmt.Errorf("RemoveFilesActivity: %v", err)
	}

	return &ActivityResult{Count: count}, nil
}
