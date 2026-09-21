# csv-ndjson-converter

A small command-line tool that converts CSV to newline-delimited JSON
(one JSON object per line) and back.

I keep needing this glue: an export lands as CSV, but the next step in
the pipeline (a log shipper, `jq`, a script that reads JSON lines) wants
NDJSON. Or the reverse — something spits out NDJSON and I need it in a
spreadsheet. There's no dependency-free way to do that from the shell
without writing a one-off script every time, so this is that script,
made reusable.

## Build

```
go build -o csvconv .
```

## Usage

CSV to NDJSON, using the first row as the field names:

```
$ cat people.csv
name,age,city
Alice,34,Denver
Bob,41,Reno

$ ./csvconv -to json -in people.csv
{"age":"34","city":"Denver","name":"Alice"}
{"age":"41","city":"Reno","name":"Bob"}
```

NDJSON back to CSV:

```
$ ./csvconv -to csv -in people.ndjson
age,city,name
34,Denver,Alice
41,Reno,Bob
```

Both directions read stdin and write stdout by default, so it pipes:

```
$ cat people.csv | ./csvconv -to json | jq 'select(.city == "Reno")'
```

## Flags

- `-to json|csv` — required, sets the conversion direction.
- `-delim` — CSV delimiter, defaults to `,`. Use `-delim $'\t'` for TSV.
- `-in` — input file, defaults to stdin.
- `-out` — output file, defaults to stdout.

## Known limitations

- NDJSON values are read as strings only; numbers, booleans, nested
  objects, and arrays in the input will fail to decode.
- Output columns are sorted alphabetically rather than preserving the
  original field order, because a decoded JSON object doesn't retain one.
- The whole input is read into memory for CSV-to-NDJSON... no, actually
  only the NDJSON-to-CSV direction buffers everything, since it needs to
  see every record before it knows the full column set.

## License

MIT, see LICENSE.
