package bagit_test

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/bagit"
)

func TestMockValidator(t *testing.T) {
	t.Parallel()
	t.Run("Return a nil error", func(t *testing.T) {
		t.Parallel()
		v := bagit.NewMockValidator()
		assert.NilError(t, v.Validate(""))
	})

	t.Run("Return an ErrInvalid error", func(t *testing.T) {
		t.Parallel()
		v := bagit.NewMockValidator().SetErr(bagit.ErrInvalid)
		assert.ErrorIs(t, v.Validate(""), bagit.ErrInvalid)
	})

	t.Run("Return a general error", func(t *testing.T) {
		t.Parallel()
		v := bagit.NewMockValidator().SetErr(errors.New("system error"))
		err := v.Validate("")
		assert.Assert(t, !errors.Is(err, bagit.ErrInvalid))
		assert.Error(t, err, "system error")
	})
}
