package native

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

type gccInvariant struct{ file, guard, main, seam, what string }

var gccInvariants = map[string]gccInvariant{
	"DTOA": {"dtoa.c", "if (other->used_digits < 0) { adamic_invariant(\"bignum digit count below zero\"); }", `
int main(int argc,char **argv){
 adamic_start(argc,argv);
 bignum a,b; bignum_init(&a);bignum_init(&b);
 bignum_assign_uint64(&a,123456789);bignum_assign_bignum(&b,&a);
 if(!bignum_equal(&a,&b))return 2;
 a.used_digits=-1;bignum_assign_bignum(&b,&a);return 0;
}
`, "", "bignum digit count below zero"},
	"Reduction": {"ieee754.c", "if (jz < 0) { adamic_invariant(\"argument reduction left no terms\"); }", `
int main(int argc,char **argv){
 adamic_start(argc,argv);
 volatile double x=1e300;
 printf("%.17g\n",adamic_math_sin(x));return 0;
}
`, "  // fdlibm __kernel_rem_pio2:", "argument reduction left no terms"},
	"Map": {"map.c", "if (map == NULL) { adamic_invariant(\"map iteration of a missing map\"); }", `
int main(int argc,char **argv){
 adamic_start(argc,argv);
 adamic_map *map=adamic_map_new(false,false);
 adamic_map_iterator *it=adamic_map_iterate(map);adamic_value key,value;
 if(adamic_map_iterator_next(it,&key,&value))return 2;
 adamic_release(it);adamic_release(map);
 adamic_map_iterate(NULL);return 0;
}
`, "", "map iteration of a missing map"},
}

// Build the normal sanitized archive once. Only the inspected C translation
// unit is embedded in each private control/mutant main, with its public names
// renamed so the whole shipping archive remains linked and unchanged.
func gccInvariantProduct(test *testing.T, kind string) string {
	test.Helper()
	spec, ok := gccInvariants[kind]
	if !ok {
		test.Fatal(kind)
	}
	options := Options{Sanitize: true}
	files, flags, compiler, version, runtimeInputsError := runtimeInputs("", options, nil)
	if runtimeInputsError != nil {
		test.Fatal(runtimeInputsError)
	}
	headers, headersReadError := headersRead(files, flags, compiler)
	if headersReadError != nil {
		test.Fatal(headersReadError)
	}
	inputs := buildcache.Inputs{
		Name: "gcc-invariant-" + kind,
		Files: []string{
			"go.mod",
			"bridge/tsgo/tsgo.h",
			"internal/native/runtime",
			"internal/native/gcc_invariants_test.go",
			"internal/native/native.go",
			"internal/native/library.go",
			"internal/native/build_runtime.go",
			"internal/native/target.go",
		},
		Flags: portableKeyParts(files, flags, compiler, headers),
		Toolchain: []string{
			buildcache.Tool(compiler, "--version"),
			buildcache.Tool(archiverName(compiler), "--version"),
		},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		cache, cacheDirectoryError := cacheDirectory("gcc-invariant")
		if cacheDirectoryError != nil {
			return cacheDirectoryError
		}
		library, cachedRuntimeError := cachedRuntime(files, flags, compiler, version, cache)
		if cachedRuntimeError != nil {
			return cachedRuntimeError
		}
		data, readFileError := runtime.ReadFile("runtime/" + spec.file)
		if readFileError != nil {
			return readFileError
		}
		text := string(data)
		if strings.Count(text, spec.guard) != 1 {
			return fmt.Errorf("guard seam missing: %s", kind)
		}
		if spec.seam != "" {
			if strings.Count(text, spec.seam) != 1 {
				return fmt.Errorf("fault seam missing")
			}
			// Exercise the impossible state immediately before its guard. Initialize
			// fq[0] only in this private copy so the missing-guard mutant has a
			// deterministic observable return, rather than an uninitialized read.
			text = strings.Replace(text, spec.seam, "  jz=-1; fq[0]=0;\n"+spec.seam, 1)
		}
		for _, variant := range []string{"control", "mutant"} {
			source := text
			if variant == "mutant" {
				source = strings.Replace(source, spec.guard, "/* invariant guard removed */", 1)
			}
			var renames strings.Builder
			for _, match := range regexp.MustCompile(
				`(?m)^[A-Za-z_][^;\n]*?\b(adamic_[A-Za-z0-9_]+)\([^;\n]*\) \{`,
			).FindAllStringSubmatch(text, -1) {
				fmt.Fprintf(&renames, "#define %s probe_%s\n", match[1], match[1])
			}
			source = "#include <stdio.h>\n" + renames.String() + source + spec.main
			if buildWithRuntimeError := BuildWithRuntime(
				source,
				filepath.Join(directory, variant),
				options,
				library,
			); buildWithRuntimeError != nil {
				return buildWithRuntimeError
			}
		}
		return nil
	})
}
func checkGccInvariant(test *testing.T, kind string, mutant bool) {
	test.Helper()
	variant := "control"
	if mutant {
		variant = "mutant"
	}
	result := inlineArrayRun(test, filepath.Join(gccInvariantProduct(test, kind), variant))
	stop := result.exit == 70 &&
		string(result.stderr) == "adamic: panic: runtime invariant broken: "+gccInvariants[kind].what+"\n"
	if !mutant {
		if !stop || len(result.stdout) != 0 {
			test.Fatalf(
				"missing loud invariant stop: exit %d stdout %q stderr %s",
				result.exit,
				result.stdout,
				result.stderr,
			)
		}
	} else {
		if stop {
			test.Fatal("guard-removal mutant survived")
		}
		if kind == "Reduction" {
			if result.exit != 0 || len(result.stdout) == 0 {
				test.Fatalf("reduction mutant must return a wrong result: %+v", result)
			}
		} else if result.exit != 1 || !strings.Contains(string(result.stderr), "runtime error:") {
			test.Fatalf("missing UBSan witness: exit %d %s", result.exit, result.stderr)
		}
	}
	test.Logf("%s %s: exit %d stdout %q", kind, variant, result.exit, result.stdout)
}
func TestProduct_GccInvariantDtoa(test *testing.T) {
	test.Parallel()
	gccInvariantProduct(test, "DTOA")
}
func TestProduct_GccInvariantReduction(test *testing.T) {
	test.Parallel()
	gccInvariantProduct(test, "Reduction")
}
func TestProduct_GccInvariantMap(test *testing.T) { test.Parallel(); gccInvariantProduct(test, "Map") }
func TestGccInvariantDtoa(test *testing.T)        { test.Parallel(); checkGccInvariant(test, "DTOA", false) }
func TestGccInvariantReduction(test *testing.T) {
	test.Parallel()
	checkGccInvariant(test, "Reduction", false)
}
func TestGccInvariantMap(test *testing.T) { test.Parallel(); checkGccInvariant(test, "Map", false) }
func TestGccInvariantDtoaMutant(test *testing.T) {
	test.Parallel()
	checkGccInvariant(test, "DTOA", true)
}
func TestGccInvariantReductionMutant(test *testing.T) {
	test.Parallel()
	checkGccInvariant(test, "Reduction", true)
}
func TestGccInvariantMapMutant(test *testing.T) {
	test.Parallel()
	checkGccInvariant(test, "Map", true)
}
