package bagvalidate

import (
	"errors"
	"strings"
)

var (
	ErrNotABag = errors.New("not a bag")

	ErrInvalid = errors.New("invalid")
)

func RemovePathFromError(path string, err error) string {
	// Remove path from validation messages.
	return strings.Replace(err.Error(), path+" is invalid: ", "", 1)
}
