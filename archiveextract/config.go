package archiveextract

import (
	"io/fs"
)

const (
	defaultDirMode  fs.FileMode = 0o700
	defaultFileMode fs.FileMode = 0o600
)

type Config struct {
	DirMode  fs.FileMode
	FileMode fs.FileMode
}

func (c *Config) setDefaults() {
	if c.DirMode == 0 {
		c.DirMode = defaultDirMode
	}

	if c.FileMode == 0 {
		c.FileMode = defaultFileMode
	}
}
