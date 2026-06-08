package ffvalidate_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/temporal-activities/ffvalidate"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		cfg     ffvalidate.Config
		wantErr string
	}

	for _, tt := range []test{
		{
			name: "No errors with allowlist path",
			cfg:  ffvalidate.Config{AllowlistPath: "/path/to/allowlist.csv"},
		},
		{
			name: "No errors with disallowlist path",
			cfg:  ffvalidate.Config{DisallowlistPath: "/path/to/disallowlist.csv"},
		},
		{
			name: "No errors with no list path",
		},
		{
			name: "Errors when both list paths are configured",
			cfg: ffvalidate.Config{
				AllowlistPath:    "/path/to/allowlist.csv",
				DisallowlistPath: "/path/to/disallowlist.csv",
			},
			wantErr: "AllowlistPath and DisallowlistPath cannot both be set",
		},
	} {
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
