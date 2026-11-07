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

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testguard"
)

// The sanitizer witness proves the combined checker groups can fail on what only the sanitizer sees: a memory error
// between rules in one process, which an agreeing output would hide. It plants an out-of-bounds heap write at the top
// of no-bitwise's bitwiseOperator in the port's emitted C, so the write happens only when no-bitwise runs, and runs the
// one group that holds no-bitwise against it. That group must fail on AddressSanitizer's report (#5ggwf8c).
const compilerAgreementHeapWitnessRule = "no-bitwise"

// Match the declaration spelling and stable digest, including any readable module prefix.
// The definition anchor excludes its prototype and calls.
var compilerAgreementHeapWitnessDefinition = regexp.MustCompile(
	`(?m)^(static adamic_string \* adamic_function_[A-Za-z0-9_]*bitwiseOperator_[a-f0-9]{32}\([^)]*\) \{)$`,
)

func TestStableCompilerAgreementHeapWitnessReader(test *testing.T) {
	test.Parallel()
	program := &ir.Program{Strings: []string{"ok"}, Functions: []ir.Function{
		{
			Name:    "bitwiseOperator",
			Source:  ir.SourceIdentity{Module: "test-module.a", Declaration: []string{"bitwiseOperator"}},
			Returns: ir.String,
			Body:    []ir.Statement{ir.Return{Value: ir.StringConstant{Index: 0}}},
		},
		{
			Name:    "bitwiseOperatorOther",
			Source:  ir.SourceIdentity{Module: "test-module.a", Declaration: []string{"bitwiseOperatorOther"}},
			Returns: ir.String,
			Body:    []ir.Statement{ir.Return{Value: ir.StringConstant{Index: 0}}},
		},
	}}
	source := []byte(native.C(program))
	if found := len(compilerAgreementHeapWitnessDefinition.FindAllIndex(source, -1)); found != 1 {
		test.Fatalf("stable bitwiseOperator definition matches: %d, want 1", found)
	}
	planted := compilerAgreementHeapWitnessDefinition.ReplaceAll(
		source,
		[]byte("${1}"+compilerAgreementHeapWitnessPlant),
	)
	if strings.Count(string(planted), compilerAgreementHeapWitnessPlant) != 1 {
		test.Fatal("heap witness must be planted once")
	}
}

// Clang's builtins need no header, and the volatile pointer keeps the write past the optimizer.
const compilerAgreementHeapWitnessPlant = "\n\t{ volatile unsigned char *adamic_witness = (volatile unsigned char *)__builtin_malloc(8); adamic_witness[8] = 1; __builtin_free((void *)adamic_witness); }"

func compilerAgreementHeapWitnessNative(test *testing.T) string {
	test.Helper()
	options := native.Options{Sanitize: true, Jobs: 1}
	flags := append(
		native.Flags(options),
		"Split=false",
		"Jobs=1",
		compilerAgreementHeapWitnessDefinition.String(),
		compilerAgreementHeapWitnessPlant,
	)
	inputs := compilerAgreementInputs(test, "compiler-agreement-checker-heap-witness-v1", flags)
	inputs.Files = append(
		append([]string(nil), inputs.Files...),
		"stage1/cohere/lint/compiler_stage1_combined_sanitizer_witness_test.go",
	)
	product := compilerAgreementProduct(test, inputs, func(directory string) error {
		source, err := os.ReadFile(filepath.Join(compilerAgreementLowered(test), "lint.c"))
		if err != nil {
			return err
		}
		if found := len(compilerAgreementHeapWitnessDefinition.FindAllIndex(source, -1)); found != 1 {
			return fmt.Errorf(
				"the heap witness needs exactly one bitwiseOperator definition in the emitted C, found %d",
				found,
			)
		}
		planted := compilerAgreementHeapWitnessDefinition.ReplaceAll(
			source,
			[]byte("${1}"+compilerAgreementHeapWitnessPlant),
		)
		if err := os.WriteFile(filepath.Join(directory, "lint.c"), planted, 0644); err != nil {
			return err
		}
		return native.Build(string(planted), filepath.Join(directory, "lint"), options)
	})
	return filepath.Join(product, "lint")
}

func TestProduct_CompilerAgreementCheckerHeapWitness(test *testing.T) {
	test.Parallel()
	compilerAgreementMeasureProduct(test, func(test *testing.T) string {
		binary := compilerAgreementHeapWitnessNative(test)
		requireLintArtifacts(test, filepath.Dir(binary), "lint.c")
		requireLintExecutable(test, binary)
		return binary
	})
}

// compilerAgreementHeapWitnessOwner is the one group that holds no-bitwise, the only rule whose code reaches the write.
func compilerAgreementHeapWitnessOwner(test *testing.T, groups []compilerAgreementCheckerGroup) int {
	test.Helper()
	for index, group := range groups {
		for _, d := range group.rules {
			if d.Slug == compilerAgreementHeapWitnessRule {
				return index
			}
		}
	}
	test.Fatalf("no checker group holds %s", compilerAgreementHeapWitnessRule)
	return -1
}

// Only the group that holds no-bitwise runs against the planted binary: the union test already proves no-bitwise is in
// exactly one group, and the control below shows another group never reaches the write.
func TestCompilerAndStage1AgreeCombinedSanitizerWitness(test *testing.T) {
	test.Parallel()
	groups := compilerAgreementCombinedGroups(test)
	part := compilerAgreementHeapWitnessOwner(test, groups)
	owned := compilerAgreementCheckerAllOwned(test)
	if len(owned) != 1 {
		test.Fatalf("checker/all has %d cases, want one", len(owned))
	}
	path := manifest(test, []string{owned[0].path + "\t" + groups[part].selection()})
	binary := compilerAgreementHeapWitnessNative(test)
	command := exec.Command(binary, "--manifest", path)
	var stderr bytes.Buffer
	command.Stdout, command.Stderr = &bytes.Buffer{}, &stderr
	err := testguard.Run(command, testguard.Budget, testguard.Ceiling)
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		test.Fatalf(
			"group %d against the heap witness: want the sanitizer to end the run, got %v\n%.2000s",
			part,
			err,
			&stderr,
		)
	}
	// The plant's signature: a one-byte write just past an eight-byte heap block. A box without a symbolizer prints
	// the stack as addresses, so the function's name is logged when present but never required.
	report := stderr.String()
	if !strings.Contains(report, "AddressSanitizer: heap-buffer-overflow") ||
		!strings.Contains(report, "0 bytes after 8-byte region") {
		test.Fatalf(
			"group %d against the heap witness exited %d without AddressSanitizer reporting the planted write:\n%.2000s",
			part,
			exit.ExitCode(),
			report,
		)
	}
	test.Logf(
		"group %d (%d rules): AddressSanitizer reported the planted write, exit %d, bitwiseOperator named %t",
		part,
		len(groups[part].rules),
		exit.ExitCode(),
		strings.Contains(report, "bitwiseOperator"),
	)
}

// The control: the cheapest other group, on the same planted binary, runs clean and agrees with Go byte for byte, so
// the witness's red comes from no-bitwise's code alone and not from the plant breaking the binary.
func TestCompilerAndStage1AgreeCombinedSanitizerWitnessControl(test *testing.T) {
	test.Parallel()
	groups := compilerAgreementCombinedGroups(test)
	owner := compilerAgreementHeapWitnessOwner(test, groups)
	control := -1
	for index, group := range groups {
		if index != owner && (control < 0 || group.millis < groups[control].millis) {
			control = index
		}
	}
	if control < 0 {
		test.Fatal("no checker group besides the witness's owner")
	}
	got, want := compilerAgreementCheckerGroupRun(test, control, compilerAgreementHeapWitnessNative(test))
	compilerAgreementCompare(test, got.output, want.output)
	test.Logf(
		"control group %d (%d rules) on the planted binary: clean, byte-identical to Go",
		control,
		len(groups[control].rules),
	)
}
