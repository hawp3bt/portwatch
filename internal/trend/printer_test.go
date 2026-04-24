package trend

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrint_NoData(t *testing.T) {
	tr := New(5)
	var buf bytes.Buffer
	Print(tr, PrintOptions{Out: &buf, Header: false})
	if !strings.Contains(buf.String(), "No data") {
		t.Fatalf("expected 'No data' message, got: %s", buf.String())
	}
}

func TestPrint_ContainsHeader(t *testing.T) {
	tr := New(5)
	tr.Record(3)
	var buf bytes.Buffer
	Print(tr, PrintOptions{Out: &buf, Header: true})
	if !strings.Contains(buf.String(), "PORT COUNT TREND") {
		t.Fatalf("expected header, got: %s", buf.String())
	}
}

func TestPrint_ShowsAllSamples(t *testing.T) {
	tr := New(5)
	tr.Record(2)
	tr.Record(4)
	tr.Record(6)
	var buf bytes.Buffer
	Print(tr, PrintOptions{Out: &buf, Header: false})
	out := buf.String()
	for _, want := range []string{"2 ports", "4 ports", "6 ports"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in output, got:\n%s", want, out)
		}
	}
}

func TestPrint_RisingIcon(t *testing.T) {
	tr := New(5)
	tr.Record(1)
	tr.Record(9)
	var buf bytes.Buffer
	Print(tr, PrintOptions{Out: &buf, Header: false})
	if !strings.Contains(buf.String(), "↑") {
		t.Fatalf("expected rising icon, got: %s", buf.String())
	}
}

func TestPrint_FallingIcon(t *testing.T) {
	tr := New(5)
	tr.Record(9)
	tr.Record(1)
	var buf bytes.Buffer
	Print(tr, PrintOptions{Out: &buf, Header: false})
	if !strings.Contains(buf.String(), "↓") {
		t.Fatalf("expected falling icon, got: %s", buf.String())
	}
}

func TestPrint_StableIcon(t *testing.T) {
	tr := New(5)
	tr.Record(5)
	tr.Record(5)
	var buf bytes.Buffer
	Print(tr, PrintOptions{Out: &buf, Header: false})
	if !strings.Contains(buf.String(), "→") {
		t.Fatalf("expected stable icon, got: %s", buf.String())
	}
}
