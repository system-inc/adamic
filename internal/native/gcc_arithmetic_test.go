package native

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/buildcache"
)

func gccArithmeticProduct(t *testing.T) string {
	t.Helper()
	compiler := gccAuditCompiler(t)
	cpu, why, ok := fusingProcessor()
	if !ok {
		t.Skip(why)
	}
	flags := slices.DeleteFunc(Flags(Options{cpu: cpu}), func(s string) bool { return s == "-Wcast-function-type-strict" || s == "-Wno-self-assign" })
	files := []runtimeFile{{"main.c", []byte(fusedHarness)}}
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "gcc-arithmetic", Files: []string{"go.mod", "bridge/tsgo/tsgo.h", "internal/native/gcc_arithmetic_test.go", "internal/native/contract_test.go", "internal/native/native.go", "internal/native/library.go", "internal/native/target.go"}, Flags: portableKeyParts(files, flags, compiler, headers), Toolchain: []string{buildcache.Tool(compiler, "--version"), buildcache.Tool("node", "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(fusedHarness), 0600); err != nil {
			return err
		}
		for _, variant := range []string{"control", "mutant"} {
			selected := slices.Clone(flags)
			if variant == "mutant" {
				for i, flag := range selected {
					if flag == "-ffp-contract=off" {
						selected[i] = "-ffp-contract=fast"
					}
				}
			}
			args := append(append(selected, prefixMaps(directory, "/adamic-gcc-arithmetic")...), "main.c", "-o", variant)
			command, release := boundedrun.Command(30*time.Second, compiler, args...)
			command.Dir = directory
			output, err := command.CombinedOutput()
			release()
			if err != nil {
				return fmt.Errorf("GCC arithmetic %s: %w %s", variant, err, output)
			}
		}
		return nil
	})
}
func checkGCCArithmetic(t *testing.T, mutant bool) {
	t.Helper()
	variant := "control"
	if mutant {
		variant = "mutant"
	}
	truth := inlineArrayRun(t, "node", "-e", "const a=.1,b=10,c=1,x=1/3; console.log(a*b-c,x*3-c,x*x-x*x)")
	got := inlineArrayRun(t, filepath.Join(gccArithmeticProduct(t), variant))
	if truth.exit != 0 || string(truth.stdout) != "0 0 0\n" || got.exit != 0 || len(got.stderr) != 0 {
		t.Fatalf("Node %+v GCC %+v", truth, got)
	}
	same := string(got.stdout) == string(truth.stdout)
	if same == mutant {
		t.Fatalf("%s: Node %q GCC %q", variant, truth.stdout, got.stdout)
	}
	t.Logf("Node %q GCC %s %q", strings.TrimSpace(string(truth.stdout)), variant, strings.TrimSpace(string(got.stdout)))
}
func TestProduct_GCCArithmetic(t *testing.T)          { t.Parallel(); gccArithmeticProduct(t) }
func TestGCCArithmeticUnfused(t *testing.T)           { t.Parallel(); checkGCCArithmetic(t, false) }
func TestGCCArithmeticContractionMutant(t *testing.T) { t.Parallel(); checkGCCArithmetic(t, true) }
