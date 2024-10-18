package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/artefactual-sdps/temporal-activities/xmlvalidate"
)

func usage() {
	fmt.Print(`
USAGE: xmlvalidate [options] XMLFILE

XMLFILE  Path of the XML file to be validated

OPTIONS:
  --xsd PATH  Path of an XSD file to validate against
  --help      Display this help and exit

`)
}

// To test: `go run ./cmd --xsd testdata/person.xsd testdata/person_valid.xml`
func main() {
	var (
		xsd  string
		help bool
	)

	p := pflag.NewFlagSet("xmlvalidate", pflag.ExitOnError)
	p.StringVar(&xsd, "xsd", "", "XSD file path")
	p.BoolVar(&help, "help", false, "display help and exit")
	_ = p.Parse(os.Args[1:])
	args := p.Args()

	if help {
		usage()
		os.Exit(0)
	}

	if len(args) == 0 {
		usage()
		fmt.Println("Error: missing XMLFILE")
		os.Exit(1)
	}

	if xsd == "" {
		usage()
		fmt.Println("Error: An --xsd file must be specified.", "")
		os.Exit(1)
	}

	v := xmlvalidate.NewXMLLintValidator()
	out, err := v.Validate(context.Background(), args[0], xsd)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if out == "" {
		fmt.Println("OK")
	} else {
		fmt.Println(out)
	}
}
