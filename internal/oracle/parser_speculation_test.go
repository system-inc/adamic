package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"strings"
	"testing"
)

var parserSpeculationNames = []string{
	"named-lookahead",
	"named-tryparse",
	"arrow-tryparse",
	"truthiness-rewind",
	"misfit",
}

func init() {
	for _, name := range parserSpeculationNames {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/parser-next/speculation/" + name + ".a", true, name == "misfit"})
	}
}
func parserSpeculationPath(test *testing.T, name string) string {
	test.Helper()
	p, err := filepath.Abs(filepath.Join(repository, "stage3/parser-next/speculation", name+".a"))
	if err != nil {
		test.Fatal(err)
	}
	return p
}

func TestParserSpeculationSourcesNamedLookahead(test *testing.T) {
	test.Parallel()
	assertParserSpeculationSource(test, "named-lookahead")
}

func TestParserSpeculationSourcesNamedTryparse(test *testing.T) {
	test.Parallel()
	assertParserSpeculationSource(test, "named-tryparse")
}

func TestParserSpeculationSourcesArrowTryparse(test *testing.T) {
	test.Parallel()
	assertParserSpeculationSource(test, "arrow-tryparse")
}

func TestParserSpeculationSourcesTruthinessRewind(test *testing.T) {
	test.Parallel()
	assertParserSpeculationSource(test, "truthiness-rewind")
}

func assertParserSpeculationSource(test *testing.T, name string) {
	test.Helper()
	path := parserSpeculationPath(test, name)
	truth := onSliceNode(test, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		test.Fatalf("Node %+v", truth)
	}
	program, err := loweredParser(test, path)
	if err != nil {
		test.Fatal(err)
	}
	actual, binary := nativelyUncached(test, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onSliceJavaScript(test, program)} {
		if d := disagreement(truth, got); d != "" {
			test.Fatalf("%s: %s", backend, d)
		}
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Logf("Node stdout %q; native sanitizers/leaks and JavaScript agree", truth.stdout)
}

func TestParserSpeculationMisfit(test *testing.T) {
	test.Parallel()
	path := parserSpeculationPath(test, "misfit")
	truth := onSliceNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != "wrong\n" {
		test.Fatalf("Node %+v", truth)
	}
	program, err := loweredParser(test, path)
	if err != nil {
		test.Fatal(err)
	}
	message := "adamic: panic: speculative result failed: callback misfit at " +
		parserDisplayedPath(test, path) +
		":3:14; expected number\n"
	actual, _ := nativelyUncached(test, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onSliceJavaScript(test, program)} {
		if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
			test.Errorf("%s: %+v, want %q", backend, got, message)
		}
	}
}
func TestParserSpeculationRewindMutant(test *testing.T) {
	test.Parallel()
	path := parserSpeculationPath(test, "truthiness-rewind")
	truth := onSliceNode(test, path)
	program, err := loweredParser(test, path)
	if err != nil {
		test.Fatal(err)
	}
	changed := 0
	for i := range program.Functions {
		if program.Functions[i].Name != "lookAhead" {
			continue
		}
		for j, statement := range program.Functions[i].Body {
			if assign, ok := statement.(ir.Assign); ok && program.Locals[assign.Local].Name == "position" {
				assign.Value = ir.Read{Local: assign.Local, Of: program.Locals[assign.Local].Type}
				program.Functions[i].Body[j] = assign
				changed++
			}
		}
	}
	if changed != 1 {
		test.Fatalf("rewind sites %d", changed)
	}
	actual, _ := nativelyUncached(test, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onSliceJavaScript(test, program)} {
		if got.exitCode != 0 ||
			len(got.stderr) != 0 ||
			disagreement(truth, got) == "" ||
			!strings.Contains(string(got.stdout), "99|7|3|false") {
			test.Fatalf("%s: rewind mutant not executed/caught %+v", backend, got)
		}
	}
	test.Log("skipped scanner-position rewind caught by Node stdout in both backends")
}
