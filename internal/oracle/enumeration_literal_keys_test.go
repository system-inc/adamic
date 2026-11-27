package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestEnumerationLiteralKeys(t *testing.T) {
	t.Parallel()
	const reserved = "Object enumeration of a literal with reserved or symbol key storage"
	for _, row := range []struct {
		name, stdout, stop string
	}{
		{"escaped-chars-both", "\\t\n", ""},
		{"escaped-chars-iterator", "\\t\n", ""},
		{"typeof-ne-hash", "1\n", ""},
		{"keys", "first,second\n", ""},
		{"values", "one,two\n", ""},
		{"k01", "9,10,b,a\n", "a computed field name"},
		{"k02", "own 1\n", "__proto__ in an object literal"},
		{"k03", "__proto__,own\n", ""},
		{"k04", "1,2,a,z one,two,a,zed\n", "a spread after the first field; spread once, first: { ...source, field: value }"},
		{"k05", "shown,more 2\n", "Object enumeration with symbol keys; use own string-keyed data properties"},
		{"k06", "computed=got,plain=p\n", "Object enumeration with a getter or accessor; use own data properties"},
		{"k07", "x,y\n", ""},
		{"identifier-proto", "1\n", reserved},
		{"string-proto-direct", "1\n", reserved},
		{"shorthand-proto-direct", "2\n", reserved},
		{"spread-reserved", "1\n", reserved},
		{"spread-reserved-deep", "1\n", reserved},
		{"spread-parameter", "1\n", reserved},
		{"spread-call", "1\n", reserved},
		{"spread-written", "1\n", reserved},
		{"spread-escaped", "first,#\n", "fixed objects and records have different storage"},
		{"encode-json", "{\"one\":1}\n", reserved},
		{"shorthand_key_inline", "__proto__,own\n", ""},
		{"shorthand_key_then_clean", "__proto__,own\na,z\n", ""},
		{"spread-clean", "a,z\n", ""},
		{"spaced-keys", "north,south\nroom 2,room 1\n", ""},
		{"digit-keys", "two,ten,last\n", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enumeration_literal_keys", row.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != row.stdout {
				t.Fatalf("Node: got %#v, want stdout %q and exit 0", node, row.stdout)
			}
			program, err := lowered(t, path)
			if row.stop != "" {
				if err == nil || !strings.Contains(err.Error(), row.stop) {
					t.Fatalf("want refusal %q, got %v", row.stop, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			// The shorthand cases require #pyak7te before this head is submitted.
			for name, got := range map[string]run{
				"JavaScript": onJavaScriptBackend(t, program),
				"native":     releasedUncached(t, program),
			} {
				if difference := disagreement(node, got); difference != "" {
					t.Errorf("%s: %s", name, difference)
				}
			}
		})
	}
}
