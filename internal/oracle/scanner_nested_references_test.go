package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

var scannerNestedReferenceProbes = []string{
	"nested-sibling-callback",
	"nested-generic-sibling-call",
	"nested-ancestor-call",
}

func init() {
	for _, name := range scannerNestedReferenceProbes {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/drivers/scanner/probes/" + name + ".a", true, false})
	}
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/nested_reference_identity.a", true, false})
}

// Each mutant changes the reached sibling/ancestor implementation. Both compiled
// backends run normally; only the independent source Node comparison catches it.
func TestScannerNestedReferenceMutants(test *testing.T) {
	test.Parallel()
	for _, name := range scannerNestedReferenceProbes {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "stage3/drivers/scanner/probes", name+".a"))
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
			}
			expected := onNode(test, path)
			target := "worker"
			if name == "nested-sibling-callback" {
				target = "error"
			}
			if name == "nested-ancestor-call" {
				target = "read"
			}
			changed := 0
			for index := range program.Functions {
				function := &program.Functions[index]
				if function.Name != target {
					continue
				}
				function.Body = nil
				if function.Returns != 0 {
					function.Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{}}}
				}
				changed++
			}
			if changed != 1 {
				test.Fatalf("want one reached implementation, changed %d", changed)
			}
			backend := onJavaScriptBackend(test, program)
			native, binary := nativelyUncached(test, program)
			for label, actual := range map[string]run{"native": native, "JavaScript": backend} {
				if actual.exitCode != 0 || len(actual.stderr) != 0 {
					test.Fatalf("%s mutant did not run cleanly: %+v", label, actual)
				}
				if difference := disagreement(expected, actual); difference != "stdout differs" {
					test.Fatalf("%s: want stdout catcher, got %q", label, difference)
				}
				test.Logf("%s mutant caught: output %q next to Node %q", label, actual.stdout, expected.stdout)
			}
			if report := leaksUncached(test, program, binary); report != "" {
				test.Fatal(report)
			}
		})
	}
}

func TestScannerNestedReferences(test *testing.T) {
	test.Parallel()
	paths := []string{"internal/oracle/testdata/nested_reference_identity.a"}
	for _, name := range scannerNestedReferenceProbes {
		paths = append(paths, "stage3/drivers/scanner/probes/"+name+".a")
	}
	for _, source := range paths {
		test.Run(filepath.Base(source), func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, source))
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
			}
			expected := onNode(test, path)
			native, binary := nativelyUncached(test, program)
			for label, actual := range map[string]run{
				"native":     native,
				"release":    releasedUncached(test, program),
				"JavaScript": onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(expected, actual); difference != "" {
					test.Fatalf(
						"%s %s: output %q, Node %q, stderr %q",
						label,
						difference,
						actual.stdout,
						expected.stdout,
						actual.stderr,
					)
				}
				test.Logf(
					"%s output %q next to Node %q; exits %d/%d",
					label,
					actual.stdout,
					expected.stdout,
					actual.exitCode,
					expected.exitCode,
				)
			}
			if report := leaksUncached(test, program, binary); report != "" {
				test.Fatal(report)
			}
		})
	}
}

func TestNestedReferenceIdentityMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/nested_reference_identity.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	changed := 0
	for index := range program.Functions {
		if program.Functions[index].FrameIdentity != 0 {
			program.Functions[index].FrameIdentity = 0
			changed++
		}
	}
	if changed == 0 {
		test.Fatal("no canonical frame changed")
	}
	expected := onNode(test, path)
	native, binary := nativelyUncached(test, program)
	if native.exitCode != 0 || len(native.stderr) != 0 {
		test.Fatalf("mutant did not run cleanly: %+v", native)
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	for label, actual := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(test, program)} {
		if disagreement(expected, actual) != "stdout differs" {
			test.Fatalf("%s identity mutant survived", label)
		}
		test.Logf("%s identity mutant caught: output %q next to Node %q", label, actual.stdout, expected.stdout)
	}
}
