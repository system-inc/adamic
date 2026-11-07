package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{
		"scanner",
		"scanner_required",
		"scanner_explicit",
		"number",
		"string",
		"boolean",
		"object",
		"defaults",
		"methods",
		"reader",
		"reader_direct",
		"reader_override",
		"reader_string",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/omitted_" + name + ".a", true, false})
	}
}

// Present zeros are valid inputs, so only Node's output comparison kills this
// mutant. Sanitizers and the leak check must stay green.
func TestOmittedArgumentZeroMutantIsCaught(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/omitted_scanner.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	changed := 0
	for index, statement := range program.Main {
		evaluate, ok := statement.(ir.Evaluate)
		if !ok {
			continue
		}
		call, ok := evaluate.Value.(ir.CallClosure)
		if !ok {
			continue
		}
		if property, ok := call.Closure.(ir.Property); ok && property.Name == "setText" {
			for len(call.Arguments) < 3 {
				call.Arguments = append(call.Arguments, ir.MaybeOf{Of: ir.MaybeNumber, Value: ir.NumberConstant{}})
				changed++
			}
		}

		evaluate.Value = call
		program.Main[index] = evaluate
	}
	if changed != 2 {
		test.Fatalf("want two omitted scalar slots, changed %d", changed)
	}
	native, sanitized := natively(test, program)
	if native.exitCode != 0 || len(native.stderr) != 0 {
		test.Fatalf("mutant must execute cleanly: %+v", native)
	}
	if report := leaks(test, program, sanitized); report != "" {
		test.Fatalf("mutant must not leak: %s", report)
	}
	if difference := disagreement(onNode(test, path), native); difference != "stdout differs" {
		test.Fatalf("want Node to catch zero padding, got %q", difference)
	}
	test.Logf("zero padding caught by Node: native %q, Node 11", native.stdout)
}

// Preserve the supplied source verbatim, as tsc wrote it: a .ts, where a postfix ! is a checked
// unwrap (a .a refuses it). The omitted branch never evaluates that assertion, as Node also proves.
func TestOmittedOriginalProbePolicy(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/refusals/omitted_scanner_original.ts"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	observed := onNode(test, path)
	if observed.exitCode != 0 || string(observed.stdout) != "11\n" || len(observed.stderr) != 0 {
		test.Fatalf("want Node 11, got %+v", observed)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	native, sanitized := natively(test, program)
	backend := onJavaScriptBackend(test, program)
	for label, actual := range map[string]run{"native": native, "JavaScript": backend} {
		if difference := disagreement(observed, actual); difference != "" {
			test.Fatalf("%s: %s", label, difference)
		}
	}
	if report := leaks(test, program, sanitized); report != "" {
		test.Fatal(report)
	}

}

func TestOmittedReaderZeroMutantIsCaught(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/omitted_reader.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	changed := 0
	for index, statement := range program.Main {
		declared, ok := statement.(ir.Declare)
		if !ok {
			continue
		}
		call, ok := declared.Value.(ir.CallClosure)
		if !ok {
			continue
		}
		if property, ok := call.Closure.(ir.Property); ok && property.Name == "pick" && len(call.Arguments) == 0 {
			call.Arguments = append(call.Arguments, ir.MaybeOf{Of: ir.MaybeNumber, Value: ir.NumberConstant{}})
			changed++
		}

		declared.Value = call
		program.Main[index] = declared
	}
	if changed != 1 {
		test.Fatalf("want one omitted method slot, changed %d", changed)
	}
	native, sanitized := natively(test, program)
	if native.exitCode != 0 || len(native.stderr) != 0 {
		test.Fatalf("mutant must execute cleanly: %+v", native)
	}
	if report := leaks(test, program, sanitized); report != "" {
		test.Fatalf("mutant must not leak: %s", report)
	}
	if difference := disagreement(onNode(test, path), native); difference != "stdout differs" {
		test.Fatalf("want Node to catch method zero padding, got %q", difference)
	}
	test.Logf("method zero padding caught by Node: native %q", native.stdout)
}
