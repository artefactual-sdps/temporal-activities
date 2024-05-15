package bagit_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/bagit"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		cfg     bagit.Config
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "No errors on valid config",
			cfg:  bagit.Config{ChecksumAlgorithm: "md5"},
		},
		{
			name:    "Errors on invalid ChecksumAlgorithm",
			cfg:     bagit.Config{ChecksumAlgorithm: "foo"},
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
