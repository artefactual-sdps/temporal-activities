package bagvalidate_test

import (
	"testing"

	"github.com/artefactual-sdps/temporal-activities/bagvalidate"
	"gotest.tools/v3/assert"
)

func TestValidatorConfig(t *testing.T) {
	t.Parallel()
	type test struct {
		name    string
		config  bagvalidate.Config
		want    bagvalidate.Config
		wantErr string
	}
	tests := []test{
		{
			name: "valid config",
			config: bagvalidate.Config{
				CacheDir: "/home/enduro/bagvalidator_cache",
				PoolSize: 2,
			},
			want: bagvalidate.Config{
				CacheDir: "/home/enduro/bagvalidator_cache",
				PoolSize: 2,
			},
		},
		{
			name:   "default to a pool size of 1",
			config: bagvalidate.Config{},
			want:   bagvalidate.Config{PoolSize: 1},
		},
		{
			name: "invalid pool size",
			config: bagvalidate.Config{
				PoolSize: -1,
			},
			wantErr: "PoolSize: -1 is less than the minimum value (1)",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.config.Validate()
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, tc.config, tc.want)
		})
	}
}
