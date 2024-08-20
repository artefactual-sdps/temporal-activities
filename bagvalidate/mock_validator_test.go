package bagvalidate_test

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
)

func TestMockValidator(t *testing.T) {
	t.Parallel()
	t.Run("Return a nil error", func(t *testing.T) {
		t.Parallel()
		v := bagvalidate.NewMockValidator()
		assert.NilError(t, v.Validate(""))
	})

	t.Run("Return an ErrInvalid error", func(t *testing.T) {
		t.Parallel()
		v := bagvalidate.NewMockValidator().SetErr(bagvalidate.ErrInvalid)
		assert.ErrorIs(t, v.Validate(""), bagvalidate.ErrInvalid)
	})

	t.Run("Return a general error", func(t *testing.T) {
		t.Parallel()
		v := bagvalidate.NewMockValidator().SetErr(errors.New("system error"))
		err := v.Validate("")
		assert.Assert(t, !errors.Is(err, bagvalidate.ErrInvalid))
		assert.Error(t, err, "system error")
	})
}
