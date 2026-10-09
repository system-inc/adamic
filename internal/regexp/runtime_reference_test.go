package regexp

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRuntimeReferenceNode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Pattern, Flags string
		Valid          bool
	}{
		{"(?<ⸯ>a)", "u", false}, {"(?<Ᲊ>a)", "u", true}, {"a\x00b", "u", true}, {"\x00", "", true}, {"(?<℘>a)", "u", true}, {"(?<a·>a)", "u", true},
		{`(?<\u2118>a)`, "u", true}, {`(?<a\u00b7>a)`, "u", true},
		{`\u{10000000000000000000000}`, "u", false},
		{`\u{ffffffffffffffffffffffff}`, "u", false},
		{`\u{10ffff}`, "u", true}, {`\u{110000}`, "u", false},
	}
	input, _ := json.Marshal(cases)
	command := exec.Command("node", "-e", `console.log(JSON.stringify(JSON.parse(require('fs').readFileSync(0,'utf8')).map(x=>{try{new RegExp(x.Pattern,x.Flags);return true}catch(e){if(!(e instanceof SyntaxError))throw e;return false}})))`)
	command.Stdin = bytes.NewReader(input)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v %s", err, output)
	}
	var valid []bool
	if err := json.Unmarshal(output, &valid); err != nil {
		t.Fatal(err)
	}
	for i, c := range cases {
		_, err := Parse(c.Pattern, c.Flags)
		if valid[i] != c.Valid || (err == nil) != valid[i] {
			t.Errorf("%q: Go=%v Node=%v want=%v", c.Pattern, err, valid[i], c.Valid)
		}
	}
	for _, pattern := range []string{"a{9223372036854775808,9223372036854775807}", "a{2147483648,2147483647}", "a{2147483649,2147483648}"} {
		_, err := Parse(pattern, "")
		var divergence *V8DivergenceError
		if !errors.As(err, &divergence) {
			t.Errorf("%s: must refuse with V8DivergenceError, got %T %v", pattern, err, err)
			continue
		}
		if divergence.Behavior != "clamps quantifier bounds above 2^31-1 before the min > max check" || divergence.Section != "22.2.1.1 Static Semantics: Early Errors, QuantifierPrefix" {
			t.Errorf("wrong ruling: %+v", divergence)
		}
		command := exec.Command("node", "-e", "new RegExp(process.argv[1]);", pattern)
		if output, err := command.CombinedOutput(); err != nil {
			t.Errorf("Node no longer accepts %s: %v %s", pattern, err, output)
		}
	}
	for _, pattern := range []string{"a{3,2}", "a{2147483648,2147483646}"} {
		_, err := Parse(pattern, "")
		var syntax *SyntaxError
		if !errors.As(err, &syntax) {
			t.Errorf("control %s: want SyntaxError got %v", pattern, err)
		}
	}
}

// Each isolated source overlay changes the implementation, builds successfully,
// and must fail the Node comparison or the exact divergence-type assertion.
func TestRuntimeReferenceMutants(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("parser.go")
	if err != nil {
		t.Fatal(err)
	}
	mutants := runtimeReferenceMutantShards(t, string(source))
	if err := checkRuntimeReferenceMutantCoverage(mutants); err != nil {
		t.Fatal(err)
	}
	original, _ := filepath.Abs("parser.go")
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			if strings.Count(string(source), mutant.old) != 1 {
				t.Fatal("mutation location changed")
			}
			changed := strings.Replace(string(source), mutant.old, mutant.new, 1)

			dir := t.TempDir()
			replacement := filepath.Join(dir, "parser.go")
			if err := os.WriteFile(replacement, []byte(changed), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			path := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(path, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("go", "test", "-overlay", path, ".", "-run", "^TestRuntimeReferenceNode$", "-count=1")
			output, err := command.CombinedOutput()
			if err == nil {
				t.Fatal("mutant survived")
			}
			if bytes.Contains(output, []byte("build failed")) || !bytes.Contains(output, []byte("--- FAIL: TestRuntimeReferenceNode")) {
				t.Fatalf("mutant did not reach its assertion: %v %s", err, output)
			}
			t.Logf("caught %s: %s", mutant.name, output)
		})
	}
}

type runtimeReferenceMutant struct{ name, old, new string }

func runtimeReferenceMutantShards(t *testing.T, source string) []runtimeReferenceMutant {
	t.Helper()
	divergenceReturn := regexp.MustCompile(`(?s)return nil, &V8DivergenceError\{.*?
\s*\}`).FindString(string(source))
	if divergenceReturn == "" {
		t.Fatal("missing divergence return")
	}
	return []runtimeReferenceMutant{
		{"stop on embedded NUL", "(stop == 0 || p.peek() != stop)", "p.peek() != stop"},
		{"drop Other_ID_Start", "return r == '$' || r == '_' || property.Contains(r)", "return r == '$' || r == '_' || (r != 0x2118 && property.Contains(r))"},
		{"drop Other_ID_Continue", "r == 0x200D || property.Contains(r)", "r == 0x200D || (r != 0x00b7 && property.Contains(r))"},
		{"accept Pattern_Syntax letter", "return r == '$' || r == '_' || property.Contains(r)", "return r == '$' || r == '_' || r == 0x2e2f || property.Contains(r)"},
		{"wrap oversized Unicode escape", "d < 0 || n > (utf8.MaxRune-d)/16", "d < 0"},
		{"silently accept clamped reversed bounds", divergenceReturn, "return &Quantifier{Atom: atom, Min: min, Max: max, Greedy: true}, nil"},
		{"classify divergence as SyntaxError", divergenceReturn, `return nil, p.failAt(start, "quantifier range out of order")`},
	}
}
