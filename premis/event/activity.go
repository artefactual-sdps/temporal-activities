package premis_event

import (
	"context"

	"github.com/artefactual-sdps/temporal-activities/premis"
)

const Name = "add-premis-event"

type (
	Params struct {
		PREMISFilePath string
		Agent          premis.Agent
		Summary        premis.EventSummary
	}

	Result struct{}

	Activity struct{}
)

func New() *Activity {
	return &Activity{}
}

func (a *Activity) Execute(
	ctx context.Context,
	params *Params,
) (*Result, error) {
	doc, err := premis.ParseOrInitialize(params.PREMISFilePath)
	if err != nil {
		return nil, err
	}

	err = premis.AppendEventXMLForEachObject(doc, params.Summary, params.Agent)
	if err != nil {
		return nil, err
	}

	err = premis.WriteIndentedToFile(doc, params.PREMISFilePath)
	if err != nil {
		return nil, err
	}

	return &Result{}, nil
}
