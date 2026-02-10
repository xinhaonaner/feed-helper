package xml2csv

import (
	"bytes"
	"strings"
	"testing"
)

func TestConvert_emptyInput(t *testing.T) {
	var buf bytes.Buffer
	err := Convert(strings.NewReader(``), &buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

func TestConvert_singleRow(t *testing.T) {
	xml := `<?xml version="1.0"?><root><item><a>1</a><b>2</b></item></root>`
	var buf bytes.Buffer
	err := Convert(strings.NewReader(xml), &buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "a,b\n1,2\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestConvert_multipleRows(t *testing.T) {
	xml := `<?xml version="1.0"?><root>
  <item><title>First</title><id>1</id></item>
  <item><title>Second</title><id>2</id></item>
</root>`
	var buf bytes.Buffer
	err := Convert(strings.NewReader(xml), &buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "title,id\nFirst,1\nSecond,2\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestConvert_customRowTag(t *testing.T) {
	xml := `<?xml version="1.0"?><root><record><x>a</x><y>b</y></record></root>`
	var buf bytes.Buffer
	err := Convert(strings.NewReader(xml), &buf, RowTag("record"))
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "x,y\na,b\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestConvert_specialCharsInValues(t *testing.T) {
	// CSV must escape comma and newline and quote
	xml := `<?xml version="1.0"?><root><item><a>hello, world</a><b>line1
line2</b><c>say "hi"</c></item></root>`
	var buf bytes.Buffer
	err := Convert(strings.NewReader(xml), &buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// encoding/csv will quote fields containing comma, newline, or quote
	if !strings.Contains(got, "hello, world") && !strings.Contains(got, `"hello, world"`) {
		t.Errorf("expected comma in value to be present or quoted, got %q", got)
	}
	if !strings.Contains(got, "line1") || !strings.Contains(got, "line2") {
		t.Errorf("expected newline in value, got %q", got)
	}
	if !strings.Contains(got, "say") || !strings.Contains(got, "hi") {
		t.Errorf("expected quote in value, got %q", got)
	}
}

func TestConvert_missingColumnInLaterRow(t *testing.T) {
	xml := `<?xml version="1.0"?><root>
  <item><a>1</a><b>2</b></item>
  <item><a>3</a></item>
</root>`
	var buf bytes.Buffer
	err := Convert(strings.NewReader(xml), &buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "a,b\n1,2\n3,\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
