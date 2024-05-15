package bagit

import (
	"fmt"
	"slices"
	"strings"
)

var hashingAlgorithms = []string{"md5", "sha1", "sha256", "sha512"}

type Config struct {
	// ChecksumAlgorithm specifies the hashing algorithm used to generate file
	// checksums. Valid values are "md5", "sha1", "sha256", "sha512" (default)
	ChecksumAlgorithm string
}

func (c *Config) setDefaults() {
	if c.ChecksumAlgorithm == "" {
		c.ChecksumAlgorithm = "sha512"
	}
}

func (c *Config) Validate() error {
	c.setDefaults()

	if !slices.Contains(hashingAlgorithms, c.ChecksumAlgorithm) {
		return fmt.Errorf(
			"ChecksumAlgorithm: invalid value %q, must be one of (%s)",
			c.ChecksumAlgorithm,
			strings.Join(hashingAlgorithms, ", "),
		)
	}

	return nil
}
