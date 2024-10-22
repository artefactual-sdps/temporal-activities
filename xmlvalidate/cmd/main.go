package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/pflag"

	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
)

func seeHelp() string {
	return "See 'xmlvalidate --help' for usage."
}

func usage() string {
	return `
USAGE: xmlvalidate [options] XMLFILE

XMLFILE  Path of the XML file to be validated

OPTIONS:
  --xsd PATH  Path of an XSD file to validate against
  --help      Display this help and exit

`
}

// To test: `go run ./cmd --xsd testdata/person.xsd testdata/person_valid.xml`
func main() {
	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	// Execute program.
	out, err := Run(ctx, os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, out)
}

func Run(ctx context.Context, args []string) (string, error) {
	var (
		help bool
		xsd  string
	)

	p := pflag.NewFlagSet(args[0], pflag.ExitOnError)
	p.BoolVarP(&help, "help", "h", false, "display help and exit")
	p.StringVar(&xsd, "xsd", "", "XSD file `path`")

	var out string
	p.Usage = func() {
		out = usage()
	}

	if err := p.Parse(args[1:]); err != nil {
		if err == pflag.ErrHelp {
			return out, nil
		}
		return "", err
	}
	args = p.Args()

	if help {
		return usage(), nil
	}

	if len(args) == 0 {
		return "", fmt.Errorf("missing XMLFILE\n%s", seeHelp())
	}

	if xsd == "" {
		return "", fmt.Errorf("an --xsd file must be specified\n%s", seeHelp())
	}

	v := xmlvalidate.NewXMLLintValidator()
	out, err := v.Validate(ctx, args[0], xsd)
	if err != nil {
		return "", err
	}

	if out == "" {
		out = "OK"
	}

	return out, nil
}
