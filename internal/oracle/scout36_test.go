package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"scout36_token_spellings.a", "scout36_literal_cache.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

// Each valid, leak-clean mutant must fail only the original source's Node answer.
func TestScout36SourceMutants(t *testing.T) {
	for _, one := range []struct{ name, from, to string }{
		{"token_spellings", "return tokenStrings[token];", "return tokenStrings[token + 1];"},
		{"literal_cache", "cache.set(value, type);", "// mutant: omit interning"},
	} {
		t.Run(one.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout36_"+one.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(data), one.from) != 1 {
				t.Fatal("mutant anchor changed")
			}
			mutant := filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(mutant, []byte(strings.Replace(string(data), one.from, one.to, 1)), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, mutant)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			got, binary := natively(t, program)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node: %+v", truth)
			}
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed outside output comparison: %+v", got)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatalf("mutant leaked: %s", report)
			}
			if bytes.Equal(got.stdout, truth.stdout) {
				t.Fatal("mutant survived Node comparison")
			}
			t.Logf("Node rejects %s mutant; sanitized execution and leak check pass", one.name)
		})
	}
}
