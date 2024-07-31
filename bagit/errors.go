package bagit

import (
	"errors"
	"fmt"
	"strings"

	bagit_gython "github.com/artefactual-labs/bagit-gython"
)

var ErrInvalid = errors.New("invalid")

func convertError(err error) error {
	if errors.Is(err, bagit_gython.ErrInvalid) {
		msg, _ := strings.CutPrefix(err.Error(), "invalid: ")
		err = fmt.Errorf("%w: %s", ErrInvalid, msg)
	}

	return err
}
