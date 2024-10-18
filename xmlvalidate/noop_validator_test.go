package xmlvalidate_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
)

func TestNoopValidator(t *testing.T) {
	t.Parallel()

	v := xmlvalidate.NewNoopValidator()
	out, err := v.Validate(context.Background(), "", "")
	assert.NilError(t, err)
	assert.Equal(t, out, "")
}
