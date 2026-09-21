package main

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"sort"
)

// CSVToNDJSON reads a CSV stream (header row first) and writes one JSON
// object per line, keyed by the header. Short rows leave trailing fields
// unset rather than erroring, since ragged exports are common in practice.
func CSVToNDJSON(r io.Reader, w io.Writer, delim rune) error {
	cr := csv.NewReader(r)
	cr.Comma = delim
	cr.FieldsPerRecord = -1

	header, err := cr.Read()
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		obj := make(map[string]string, len(header))
		for i, h := range header {
			if i < len(rec) {
				obj[h] = rec[i]
			}
		}
		if err := enc.Encode(obj); err != nil {
			return err
		}
	}
}

// NDJSONToCSV reads one JSON object per line and writes CSV with a header
// row. Values must be JSON strings; the column set is the union of all
// keys seen, sorted alphabetically since a JSON object has no field order
// of its own to fall back on. The whole input is buffered in memory,
// which is fine for the sizes this tool is meant for and simple to reason
// about.
func NDJSONToCSV(r io.Reader, w io.Writer, delim rune) error {
	dec := json.NewDecoder(r)

	var records []map[string]string
	seen := map[string]bool{}
	for dec.More() {
		var rec map[string]string
		if err := dec.Decode(&rec); err != nil {
			return err
		}
		records = append(records, rec)
		for k := range rec {
			seen[k] = true
		}
	}

	headers := make([]string, 0, len(seen))
	for k := range seen {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	cw := csv.NewWriter(w)
	cw.Comma = delim
	if err := cw.Write(headers); err != nil {
		return err
	}
	for _, rec := range records {
		row := make([]string, len(headers))
		for i, h := range headers {
			row[i] = rec[h]
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
