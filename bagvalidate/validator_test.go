package bagvalidate_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
)

func TestNoopValidator(t *testing.T) {
	v := bagvalidate.NewNoopValidator()
	assert.NilError(t, v.Validate(""))
}
