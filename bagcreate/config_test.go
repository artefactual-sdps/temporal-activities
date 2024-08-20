package bagcreate_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/bagcreate"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		cfg     bagcreate.Config
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "No errors on valid config",
			cfg:  bagcreate.Config{ChecksumAlgorithm: "md5"},
		},
		{
			name:    "Errors on invalid ChecksumAlgorithm",
			cfg:     bagcreate.Config{ChecksumAlgorithm: "foo"},
			wantErr: "ChecksumAlgorithm: invalid value \"foo\", must be one of (md5, sha1, sha256, sha512)",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.Validate()
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
		})
	}
}
