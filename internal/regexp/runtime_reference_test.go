package regexp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/system-inc/adamic/internal/buildcache"
	"regexp"
	"testing"
)

func TestRuntimeReferenceNode(test *testing.T) {
	test.Parallel()
	cases := []struct {
		Pattern, Flags string
		Valid          bool
	}{
		{"(?<ⸯ>a)", "u", false},
		{"(?<Ᲊ>a)", "u", true},
		{"a\x00b", "u", true},
		{"\x00", "", true},
		{"(?<℘>a)", "u", true},
		{"(?<a·>a)", "u", true},
		{`(?<\u2118>a)`, "u", true},
		{`(?<a\u00b7>a)`, "u", true},
		{`\u{10000000000000000000000}`, "u", false},
		{`\u{ffffffffffffffffffffffff}`, "u", false},
		{`\u{10ffff}`, "u", true},
		{`\u{110000}`, "u", false},
	}
	input, _ := json.Marshal(cases)
	command := exec.Command(
		"node",
		"-e",
		`console.log(JSON.stringify(JSON.parse(require('fs').readFileSync(0,'utf8')).map(x=>{try{new RegExp(x.Pattern,x.Flags);return true}catch(e){if(!(e instanceof SyntaxError))throw e;return false}})))`,
	)
	command.Stdin = bytes.NewReader(input)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node: %v %s", combinedOutputError, output)
	}
	var valid []bool
	if decodeError := json.Unmarshal(output, &valid); decodeError != nil {
		test.Fatal(decodeError)
	}
	for i, c := range cases {
		_, parseError := Parse(c.Pattern, c.Flags)
		if valid[i] != c.Valid || (parseError == nil) != valid[i] {
			test.Errorf("%q: Go=%v Node=%v want=%v", c.Pattern, parseError, valid[i], c.Valid)
		}
	}
	for _, pattern := range []string{
		"a{9223372036854775808,9223372036854775807}",
		"a{2147483648,2147483647}",
		"a{2147483649,2147483648}",
	} {
		_, parseError := Parse(pattern, "")
		var divergence *V8DivergenceError
		if !errors.As(parseError, &divergence) {
			test.Errorf("%s: must refuse with V8DivergenceError, got %T %v", pattern, parseError, parseError)
			continue
		}
		if divergence.Behavior != "clamps quantifier bounds above 2^31-1 before the min > max check" ||
			divergence.Section != "22.2.1.1 Static Semantics: Early Errors, QuantifierPrefix" {
			test.Errorf("wrong ruling: %+v", divergence)
		}
		command := exec.Command("node", "-e", "new RegExp(process.argv[1]);", pattern)
		if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
			test.Errorf("Node no longer accepts %s: %v %s", pattern, combinedOutputError, output)
		}
	}
	for _, pattern := range []string{"a{3,2}", "a{2147483648,2147483646}"} {
		_, parseError := Parse(pattern, "")
		var syntax *SyntaxError
		if !errors.As(parseError, &syntax) {
			test.Errorf("control %s: want SyntaxError got %v", pattern, parseError)
		}
	}
}

// Each isolated source overlay changes the implementation, builds successfully,
// and must fail the Node comparison or the exact divergence-type assertion.
func runRuntimeReferenceMutant(test *testing.T, index int) {
	test.Helper()
	source, readError := os.ReadFile("parser.go")
	if readError != nil {
		test.Fatal(readError)
	}
	mutants := runtimeReferenceMutantShards(test, string(source))
	if checkRuntimeReferenceMutantCoverageError := checkRuntimeReferenceMutantCoverage(
		mutants,
	); checkRuntimeReferenceMutantCoverageError != nil {
		test.Fatal(checkRuntimeReferenceMutantCoverageError)
	}
	mutant := mutants[index]
	command := exec.Command(
		runtimeReferenceMutantProduct(test, index),
		"-test.run=^TestRuntimeReferenceNode$",
		"-test.count=1",
		"-test.timeout=90s",
	)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError == nil {
		test.Fatal("mutant survived")
	}
	if bytes.Contains(output, []byte("build failed")) ||
		!bytes.Contains(output, []byte("--- FAIL: TestRuntimeReferenceNode")) {
		test.Fatalf("mutant did not reach its assertion: %v %s", combinedOutputError, output)
	}
	test.Logf("caught %s: %s", mutant.name, output)
}

type runtimeReferenceMutant struct{ name, old, new string }

func runtimeReferenceMutantShards(test testing.TB, source string) []runtimeReferenceMutant {
	test.Helper()
	divergenceReturn := regexp.MustCompile(`(?s)return nil, &V8DivergenceError\{.*?
\s*\}`).FindString(string(source))
	if divergenceReturn == "" {
		test.Fatal("missing divergence return")
	}
	return []runtimeReferenceMutant{
		{"stop on embedded NUL", "(stop == 0 || p.peek() != stop)", "p.peek() != stop"},
		{
			"drop Other_ID_Start",
			"return r == '$' || r == '_' || property.Contains(r)",
			"return r == '$' || r == '_' || (r != 0x2118 && property.Contains(r))",
		},
		{
			"drop Other_ID_Continue",
			"r == 0x200D || property.Contains(r)",
			"r == 0x200D || (r != 0x00b7 && property.Contains(r))",
		},
		{
			"accept Pattern_Syntax letter",
			"return r == '$' || r == '_' || property.Contains(r)",
			"return r == '$' || r == '_' || r == 0x2e2f || property.Contains(r)",
		},
		{"wrap oversized Unicode escape", "d < 0 || n > (utf8.MaxRune-d)/16", "d < 0"},
		{
			"silently accept clamped reversed bounds",
			divergenceReturn,
			"return &Quantifier{Atom: atom, Min: min, Max: max, Greedy: true}, nil",
		},
		{
			"classify divergence as SyntaxError",
			divergenceReturn,
			`return nil, p.failAt(start, "quantifier range out of order")`,
		},
	}
}

func TestRuntimeReferenceMutant0(test *testing.T) {
	test.Parallel()
	runRuntimeReferenceMutant(test, 0)
}

func TestRuntimeReferenceMutant1(test *testing.T) {
	test.Parallel()
	runRuntimeReferenceMutant(test, 1)
}

func TestRuntimeReferenceMutant2(test *testing.T) {
	test.Parallel()
	runRuntimeReferenceMutant(test, 2)
}

func TestRuntimeReferenceMutant3(test *testing.T) {
	test.Parallel()
	runRuntimeReferenceMutant(test, 3)
}

func TestRuntimeReferenceMutant4(test *testing.T) {
	test.Parallel()
	runRuntimeReferenceMutant(test, 4)
}

func TestRuntimeReferenceMutant5(test *testing.T) {
	test.Parallel()
	runRuntimeReferenceMutant(test, 5)
}

func TestRuntimeReferenceMutant6(test *testing.T) {
	test.Parallel()
	runRuntimeReferenceMutant(test, 6)
}

func runtimeReferenceMutantProduct(test testing.TB, index int) string {
	test.Helper()
	source, readError := os.ReadFile("parser.go")
	if readError != nil {
		test.Fatal(readError)
	}
	mutant := runtimeReferenceMutantShards(test, string(source))[index]
	return buildcache.GoTestMutated(
		test,
		"",
		fmt.Sprintf("runtime-reference-mutant-%d.test", index),
		regexpMutantSources(test, "capture_facts_test.go", "runtime_reference_test.go", "shards_test.go"),
		nil,
		[]buildcache.Mutation{{
			File: "internal/regexp/parser.go", Before: mutant.old, After: mutant.new,
		}},
	)
}

func TestProduct_RuntimeReferenceMutant0(test *testing.T) {
	test.Parallel()
	runtimeReferenceMutantProduct(test, 0)
}

func TestProduct_RuntimeReferenceMutant1(test *testing.T) {
	test.Parallel()
	runtimeReferenceMutantProduct(test, 1)
}

func TestProduct_RuntimeReferenceMutant2(test *testing.T) {
	test.Parallel()
	runtimeReferenceMutantProduct(test, 2)
}

func TestProduct_RuntimeReferenceMutant3(test *testing.T) {
	test.Parallel()
	runtimeReferenceMutantProduct(test, 3)
}

func TestProduct_RuntimeReferenceMutant4(test *testing.T) {
	test.Parallel()
	runtimeReferenceMutantProduct(test, 4)
}

func TestProduct_RuntimeReferenceMutant5(test *testing.T) {
	test.Parallel()
	runtimeReferenceMutantProduct(test, 5)
}

func TestProduct_RuntimeReferenceMutant6(test *testing.T) {
	test.Parallel()
	runtimeReferenceMutantProduct(test, 6)
}
