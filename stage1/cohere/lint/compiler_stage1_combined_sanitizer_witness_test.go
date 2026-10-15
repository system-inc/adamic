package lint

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testguard"
)

// The sanitizer witness proves the combined checker groups can fail on what only the sanitizer sees: a memory error
// between rules in one process, which an agreeing output would hide. It plants an out-of-bounds heap write at the top
// of no-bitwise's bitwiseOperator in the port's emitted C, so the write happens only when no-bitwise runs, and runs the
// one group that holds no-bitwise against it. That group must fail on AddressSanitizer's report (#5ggwf8c).
const compilerAgreementHeapWitnessRule = "no-bitwise"

// The emitted C numbers its functions, so the plant finds bitwiseOperator's definition by pattern, never by a number
// that moves when the port does.
var compilerAgreementHeapWitnessDefinition = regexp.MustCompile(`(?m)^(static adamic_string \* adamic_function_[0-9]+_bitwiseOperator\([^)]*\) \{)$`)

// Clang's builtins need no header, and the volatile pointer keeps the write past the optimizer.
const compilerAgreementHeapWitnessPlant = "\n\t{ volatile unsigned char *adamic_witness = (volatile unsigned char *)__builtin_malloc(8); adamic_witness[8] = 1; __builtin_free((void *)adamic_witness); }"

func compilerAgreementHeapWitnessNative(t *testing.T) string {
	t.Helper()
	options := native.Options{Sanitize: true, Jobs: 1}
	flags := append(native.Flags(options), "Split=false", "Jobs=1", compilerAgreementHeapWitnessDefinition.String(), compilerAgreementHeapWitnessPlant)
	inputs := compilerAgreementInputs(t, "compiler-agreement-checker-heap-witness-v1", flags)
	inputs.Files = append(append([]string(nil), inputs.Files...), "stage1/cohere/lint/compiler_stage1_combined_sanitizer_witness_test.go")
	product := compilerAgreementProduct(t, inputs, func(directory string) error {
		source, err := os.ReadFile(filepath.Join(compilerAgreementLowered(t), "lint.c"))
		if err != nil {
			return err
		}
		if found := len(compilerAgreementHeapWitnessDefinition.FindAllIndex(source, -1)); found != 1 {
			return fmt.Errorf("the heap witness needs exactly one bitwiseOperator definition in the emitted C, found %d", found)
		}
		planted := compilerAgreementHeapWitnessDefinition.ReplaceAll(source, []byte("${1}"+compilerAgreementHeapWitnessPlant))
		if err := os.WriteFile(filepath.Join(directory, "lint.c"), planted, 0644); err != nil {
			return err
		}
		return native.Build(string(planted), filepath.Join(directory, "lint"), options)
	})
	return filepath.Join(product, "lint")
}

func TestProduct_CompilerAgreementCheckerHeapWitness(t *testing.T) {
	t.Parallel()
	compilerAgreementMeasureProduct(t, compilerAgreementHeapWitnessNative)
}

// Only the group that holds no-bitwise runs against the planted binary: the union test already proves no-bitwise is in
// exactly one group, so no other group can reach the write.
func TestCompilerAndStage1AgreeCombinedSanitizerWitness(t *testing.T) {
	t.Parallel()
	groups := compilerAgreementCombinedGroups(t)
	part := -1
	for index, group := range groups {
		for _, d := range group.rules {
			if d.Slug == compilerAgreementHeapWitnessRule {
				part = index
			}
		}
	}
	if part < 0 {
		t.Fatalf("no checker group holds %s", compilerAgreementHeapWitnessRule)
	}
	owned := compilerAgreementCheckerAllOwned(t)
	if len(owned) != 1 {
		t.Fatalf("checker/all has %d cases, want one", len(owned))
	}
	path := manifest(t, []string{owned[0].path + "\t" + groups[part].selection()})
	binary := compilerAgreementHeapWitnessNative(t)
	command := exec.Command(binary, "--manifest", path)
	var stderr bytes.Buffer
	command.Stdout, command.Stderr = &bytes.Buffer{}, &stderr
	err := testguard.Run(command, testguard.Budget, testguard.Ceiling)
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("group %d against the heap witness: want the sanitizer to end the run, got %v\n%.2000s", part, err, &stderr)
	}
	// The plant's signature: a one-byte write just past an eight-byte heap block. A box without a symbolizer prints
	// the stack as addresses, so the function's name is logged when present but never required.
	report := stderr.String()
	if !strings.Contains(report, "AddressSanitizer: heap-buffer-overflow") || !strings.Contains(report, "0 bytes after 8-byte region") {
		t.Fatalf("group %d against the heap witness exited %d without AddressSanitizer reporting the planted write:\n%.2000s", part, exit.ExitCode(), report)
	}
	t.Logf("group %d (%d rules): AddressSanitizer reported the planted write, exit %d, bitwiseOperator named %t", part, len(groups[part].rules), exit.ExitCode(), strings.Contains(report, "bitwiseOperator"))
}
