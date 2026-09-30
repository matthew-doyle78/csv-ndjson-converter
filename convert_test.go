package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestCSVToNDJSON(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		delim rune
		want  string
	}{
		{
			name:  "basic",
			in:    "name,age\nalice,30\nbob,25\n",
			delim: ',',
			want:  `{"age":"30","name":"alice"}` + "\n" + `{"age":"25","name":"bob"}` + "\n",
		},
		{
			name:  "header only",
			in:    "a,b\n",
			delim: ',',
			want:  "",
		},
		{
			name:  "short row leaves trailing fields unset",
			in:    "a,b,c\n1,2\n",
			delim: ',',
			want:  `{"a":"1","b":"2"}` + "\n",
		},
		{
			name:  "long row drops extra fields",
			in:    "a\n1,2\n",
			delim: ',',
			want:  `{"a":"1"}` + "\n",
		},
		{
			name:  "quoted comma and newline",
			in:    "a,b\n\"x,y\",\"line1\nline2\"\n",
			delim: ',',
			want:  `{"a":"x,y","b":"line1\nline2"}` + "\n",
		},
		{
			name:  "semicolon delimiter",
			in:    "a;b\n1,5;2\n",
			delim: ';',
			want:  `{"a":"1,5","b":"2"}` + "\n",
		},
		{
			name:  "tab delimiter",
			in:    "a\tb\n1\t2\n",
			delim: '\t',
			want:  `{"a":"1","b":"2"}` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := CSVToNDJSON(strings.NewReader(tt.in), &out, tt.delim); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCSVToNDJSONEmptyInput(t *testing.T) {
	var out bytes.Buffer
	err := CSVToNDJSON(strings.NewReader(""), &out, ',')
	if err != io.EOF {
		t.Errorf("got error %v, want io.EOF", err)
	}
	if out.Len() != 0 {
		t.Errorf("wrote %q for empty input", out.String())
	}
}

func TestCSVToNDJSONMalformed(t *testing.T) {
	var out bytes.Buffer
	err := CSVToNDJSON(strings.NewReader("a,b\n\"unterminated,2\n"), &out, ',')
	if err == nil {
		t.Fatal("expected an error for an unterminated quote")
	}
}

func TestNDJSONToCSV(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		delim rune
		want  string
	}{
		{
			name:  "columns sorted alphabetically",
			in:    `{"name":"alice","age":"30"}` + "\n" + `{"name":"bob","age":"25"}` + "\n",
			delim: ',',
			want:  "age,name\n30,alice\n25,bob\n",
		},
		{
			name:  "union of keys with missing values",
			in:    `{"a":"1"}` + "\n" + `{"b":"2"}` + "\n",
			delim: ',',
			want:  "a,b\n1,\n,2\n",
		},
		{
			name:  "quotes fields containing the delimiter",
			in:    `{"a":"x,y","b":"say \"hi\""}` + "\n",
			delim: ',',
			want:  "a,b\n\"x,y\",\"say \"\"hi\"\"\"\n",
		},
		{
			name:  "semicolon delimiter",
			in:    `{"a":"1,5","b":"2"}` + "\n",
			delim: ';',
			want:  "a;b\n1,5;2\n",
		},
		{
			name:  "no trailing newline on input",
			in:    `{"a":"1"}`,
			delim: ',',
			want:  "a\n1\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := NDJSONToCSV(strings.NewReader(tt.in), &out, tt.delim); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNDJSONToCSVErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"number value", `{"a":1}` + "\n"},
		{"boolean value", `{"a":true}` + "\n"},
		{"not an object", `["a","b"]` + "\n"},
		{"truncated", `{"a":"1"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := NDJSONToCSV(strings.NewReader(tt.in), &out, ','); err == nil {
				t.Errorf("expected an error, got output %q", out.String())
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Header is already in alphabetical order because NDJSONToCSV sorts it.
	in := "age,name,note\n30,alice,\"a, b\"\n25,bob,\n"

	var mid bytes.Buffer
	if err := CSVToNDJSON(strings.NewReader(in), &mid, ','); err != nil {
		t.Fatalf("CSVToNDJSON: %v", err)
	}
	var out bytes.Buffer
	if err := NDJSONToCSV(&mid, &out, ','); err != nil {
		t.Fatalf("NDJSONToCSV: %v", err)
	}
	if got := out.String(); got != in {
		t.Errorf("round trip changed data: got %q, want %q", got, in)
	}
}
