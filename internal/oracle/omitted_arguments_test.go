package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
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
		for argument, value := range call.Arguments {
			missing, ok := value.(ir.MaybeOf)
			if ok && missing.Of == ir.MaybeNumber && missing.Value == nil {
				missing.Value = ir.NumberConstant{Value: 0}
				call.Arguments[argument] = missing
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

// Preserve the supplied source verbatim: Node confirms the expected output,
// while the original still needs string logical-expression lowering. The
// executable normalized witness must not be mistaken for this original.
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
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "BinaryExpression with a string and a string") {
		t.Fatalf("want the remaining string logical-expression blocker, got %v", err)
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
		for argument, value := range call.Arguments {
			missing, ok := value.(ir.MaybeOf)
			if ok && missing.Of == ir.MaybeNumber && missing.Value == nil {
				missing.Value = ir.NumberConstant{Value: 0}
				call.Arguments[argument] = missing
				changed++
			}
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
