package oracle

import (
	"flag"
	"os"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/regexp_surrogate_methods.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/regexp_surrogate_limit_refused.a", false, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/regexp_surrogate_nested_limit_refused.a", false, false})
	for _, one := range []struct {
		path   string
		lowers bool
	}{
		{"internal/oracle/testdata/regexp_compile.a", true},
		{"internal/oracle/testdata/regexp_metadata.a", true},
		{"internal/oracle/testdata/regexp_compile_source_refused.a", false},
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{one.path, one.lowers, false})
	}

}

var updateSurrogateCounts = flag.Bool("update-regexp-surrogate-counts", false, "update only surrogate fixture counts")

func TestRegExpSurrogateCounts(t *testing.T) {
	path := "internal/oracle/testdata/regexp_surrogate_methods.a"
	row := counted(t, path, false, nil, false, false)
	data, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, row+"\n") {
		return
	}
	if !*updateSurrogateCounts {
		t.Fatalf("fixture counts changed: %s", row)
	}
	prefix := "| " + path + " |"
	at := strings.Index(text, prefix)
	if at < 0 {
		at = strings.Index(text, "| internal/oracle/testdata/read_files.a |")
		if at < 0 {
			t.Fatal("input fixture boundary missing")
		}
		text = text[:at] + row + "\n" + text[at:]
	} else {
		end := at + strings.Index(text[at:], "\n")
		text = text[:at] + row + text[end:]
	}
	if err := os.WriteFile(countsPath, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
	t.Log(row)
}
