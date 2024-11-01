package ffvalidate

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.artefactual.dev/tools/temporal"
)

const Name = "validate-file-formats"

type (
	Activity struct {
		cfg Config
	}
	Params struct {
		Path string
	}
	Result struct {
		Failures []string
	}
)

func New(cfg Config) *Activity {
	return &Activity{cfg: cfg}
}

// FormatIdentifier provides a interface to identify a file's format.
type FormatIdentifier interface {
	Identify(path string) (*FileFormat, error)
	Version() string
}

// An IdentifyResult represents the result of a file format identification run.
type FileFormat struct {
	Namespace  string // Format identifier Namespace (e.g. "PRONOM")
	ID         string // PRONOM PUID (e.g. "fmt/40")
	CommonName string // Common name of format (e.g. "Microsoft Word Document")
	Version    string // Format version (e.g. "97-2003")
	MIMEType   string // MIME type (e.g. "application/msword")
	Basis      string // Basis for identification
	Warning    string // Identification warning message
}

type formatList map[string]struct{}

func (a *Activity) Execute(ctx context.Context, params *Params) (*Result, error) {
	logger := temporal.GetLogger(ctx)

	if a.cfg.AllowlistPath == "" {
		logger.Info(Name + ": No allowlist path configured, skipping file format validation")

		return nil, nil
	}

	f, err := os.Open(a.cfg.AllowlistPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", Name, err)
	}
	defer f.Close()

	allowed, err := parseFormatList(f)
	if err != nil {
		return nil, fmt.Errorf("%s: load allowed formats: %v", Name, err)
	}

	failures, err := checkAllowed(allowed, params.Path)
	if err != nil {
		return nil, fmt.Errorf("%s: check allowed formats: %v", Name, err)
	}

	return &Result{Failures: failures}, nil
}

func parseFormatList(r io.Reader) (formatList, error) {
	var i, puidIndex int
	formats := make(formatList)

	cr := csv.NewReader(r)
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid CSV: %v", err)
		}

		if i == 0 {
			// Get the index of the PRONOM PUID column.
			puidIndex = slices.IndexFunc(row, func(s string) bool {
				return strings.EqualFold(s, "pronom puid")
			})
			if puidIndex == -1 {
				return nil, errors.New(`missing "PRONOM PUID" column`)
			}
		} else {
			// Get the PRONOM PUID and ignore the rest of the columns.
			s := strings.TrimSpace(row[puidIndex])
			if s != "" {
				formats[s] = struct{}{}
			}
		}

		i++
	}

	if len(formats) == 0 {
		return nil, fmt.Errorf("no allowed file formats")
	}

	return formats, nil
}

func checkAllowed(allowed formatList, base string) ([]string, error) {
	var failures []string

	sf := NewSiegfriedEmbed()
	err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ff, err := sf.Identify(p)
		if err != nil {
			return fmt.Errorf("identify format: %v", err)
		}

		rel, err := filepath.Rel(base, p)
		if err != nil {
			return fmt.Errorf("get relative path: %v", err)
		}

		if _, ok := allowed[ff.ID]; !ok {
			failures = append(failures, fmt.Sprintf("file format %q not allowed: %q", ff.ID, rel))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return failures, nil
}
