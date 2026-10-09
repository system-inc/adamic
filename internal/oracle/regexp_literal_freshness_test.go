package oracle

import (
	"flag"
	"os"
	"strings"
	"testing"
)

const literalFreshnessFixture = "internal/oracle/testdata/regexp_literal_freshness.a"

var updateLiteralFreshnessCounts = flag.Bool("update-regexp-literal-freshness-counts", false, "record only the literal freshness fixture")

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{literalFreshnessFixture, true, false})
}

func TestRegExpLiteralFreshnessCounts(t *testing.T) {
	row := counted(t, literalFreshnessFixture, false, nil, false, false)
	t.Log(row)
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, row+"\n") {
		return
	}
	if !*updateLiteralFreshnessCounts {
		t.Fatalf("literal freshness counts missing or changed: %s", row)
	}
	at := strings.Index(text, "| internal/oracle/testdata/read_files.a |")
	if at < 0 {
		t.Fatal("input fixture boundary missing")
	}
	if strings.Contains(text, "| "+literalFreshnessFixture+" |") {
		t.Fatal("existing freshness row changed; review it explicitly")
	}
	if err := os.WriteFile(countsPath, []byte(text[:at]+row+"\n"+text[at:]), 0644); err != nil {
		t.Fatal(err)
	}
}
