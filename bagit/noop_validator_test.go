package bagit_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/bagit"
)

func TestNoopValidator(t *testing.T) {
	v := bagit.NewNoopValidator()
	assert.NilError(t, v.Validate(""))
}
