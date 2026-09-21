// csvconv converts between CSV and newline-delimited JSON (one JSON
// object per line, keyed by the CSV header row).
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	to := flag.String("to", "", `target format: "json" or "csv"`)
	delim := flag.String("delim", ",", "CSV delimiter, one character")
	inPath := flag.String("in", "", "input file (default stdin)")
	outPath := flag.String("out", "", "output file (default stdout)")
	flag.Parse()

	if *to != "json" && *to != "csv" {
		fmt.Fprintln(os.Stderr, `csvconv: -to must be "json" or "csv"`)
		os.Exit(2)
	}
	if len(*delim) != 1 {
		fmt.Fprintln(os.Stderr, "csvconv: -delim must be exactly one character")
		os.Exit(2)
	}

	in := io.Reader(os.Stdin)
	if *inPath != "" {
		f, err := os.Open(*inPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "csvconv: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	}

	out := io.Writer(os.Stdout)
	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "csvconv: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	var err error
	switch *to {
	case "json":
		err = CSVToNDJSON(in, out, rune((*delim)[0]))
	case "csv":
		err = NDJSONToCSV(in, out, rune((*delim)[0]))
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "csvconv: %v\n", err)
		os.Exit(1)
	}
}
