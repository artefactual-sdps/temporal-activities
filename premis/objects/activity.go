package premis_objects

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/temporal-activities/premis"
)

const Name = "add-premis-objects"

type (
	Params struct {
		SIPPath        string
		PREMISFilePath string
	}

	Result struct{}

	Activity struct {
		rng io.Reader
	}
)

func New(rand io.Reader) *Activity {
	return &Activity{rng: rand}
}

func (a *Activity) Execute(
	ctx context.Context,
	params *Params,
) (*Result, error) {
	// Create PREMIS file parent directory or directories, if necessary.
	mdPath := filepath.Dir(params.PREMISFilePath)
	if err := os.MkdirAll(mdPath, 0o700); err != nil {
		return nil, err
	}

	// Get subpaths of files in transfer.
	subpaths, err := premis.FilesWithinDirectory(params.SIPPath)
	if err != nil {
		return nil, err
	}

	doc, err := premis.ParseOrInitialize(params.PREMISFilePath)
	if err != nil {
		return nil, err
	}

	for _, subpath := range subpaths {
		id, err := uuid.NewRandomFromReader(a.rng)
		if err != nil {
			return nil, fmt.Errorf("generate UUID: %v", err)
		}

		object := premis.Object{
			IdType:       "UUID",
			IdValue:      id.String(),
			OriginalName: subpath,
		}

		err = premis.AppendObjectXML(doc, object)
		if err != nil {
			return nil, err
		}
	}

	err = premis.WriteIndentedToFile(doc, params.PREMISFilePath)
	if err != nil {
		return nil, err
	}

	return &Result{}, nil
}
