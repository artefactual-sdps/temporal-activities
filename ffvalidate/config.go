package ffvalidate

import "errors"

type Config struct {
	AllowlistPath    string
	DisallowlistPath string
}

// Validate checks for invalid configuration.
//
// Call Validate before registering the activity if you want configuration
// errors to fail fast instead of waiting for activity execution failures.
func (c Config) Validate() error {
	if c.AllowlistPath != "" && c.DisallowlistPath != "" {
		return errors.New("AllowlistPath and DisallowlistPath cannot both be set")
	}

	return nil
}
