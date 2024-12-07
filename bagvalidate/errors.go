package bagvalidate

import (
	"errors"
	"strings"
)

var ErrNotABag = errors.New("not a bag")

var ErrInvalid = errors.New("invalid")

func RemovePathFromError(path string, err error) string {
	// Remove path from validation messages.
	message := strings.Replace(err.Error(), path+" is invalid: ", "", 1)

	// Convert to lower case.
	return strings.ToLower(message)
}
