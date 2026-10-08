package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"scanner", "scanner_required", "scanner_explicit", "number", "string", "boolean", "object", "defaults", "methods", "reader", "reader_direct", "reader_override", "reader_string"} {
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
func TestOmittedArgumentZeroMutantIsCaught(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/omitted_scanner.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
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
		t.Fatalf("want two omitted scalar slots, changed %d", changed)
	}
	native, sanitized := natively(t, program)
	if native.exitCode != 0 || len(native.stderr) != 0 {
		t.Fatalf("mutant must execute cleanly: %+v", native)
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatalf("mutant must not leak: %s", report)
	}
	if difference := disagreement(onNode(t, path), native); difference != "stdout differs" {
		t.Fatalf("want Node to catch zero padding, got %q", difference)
	}
	t.Logf("zero padding caught by Node: native %q, Node 11", native.stdout)
}

// Preserve the supplied source and Node output. October 8 00:27 keeps postfix
// non-null assertions refused in .a, including ones the omitted branch avoids.
func TestOmittedOriginalProbePolicy(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/refusals/omitted_scanner_original.a"))
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || string(observed.stdout) != "11\n" || len(observed.stderr) != 0 {
		t.Fatalf("want Node 11, got %+v", observed)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.What != "the non-null assertion !" || refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
		t.Fatalf("want .a non-null assertion refusal, got %v", err)
	}
}

func TestOmittedReaderZeroMutantIsCaught(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/omitted_reader.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
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
		t.Fatalf("want one omitted method slot, changed %d", changed)
	}
	native, sanitized := natively(t, program)
	if native.exitCode != 0 || len(native.stderr) != 0 {
		t.Fatalf("mutant must execute cleanly: %+v", native)
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatalf("mutant must not leak: %s", report)
	}
	if difference := disagreement(onNode(t, path), native); difference != "stdout differs" {
		t.Fatalf("want Node to catch method zero padding, got %q", difference)
	}
	t.Logf("method zero padding caught by Node: native %q", native.stdout)
}
