package native

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

type gccInvariant struct{ file, guard, main, seam string }

var gccInvariants = map[string]gccInvariant{
	"DTOA": {"dtoa.c", "if (other->used_digits < 0) { adamic_unreachable(); }", `
int main(int argc,char **argv){
 adamic_start(argc,argv);
 bignum a,b; bignum_init(&a);bignum_init(&b);
 bignum_assign_uint64(&a,123456789);bignum_assign_bignum(&b,&a);
 if(!bignum_equal(&a,&b))return 2;
 a.used_digits=-1;bignum_assign_bignum(&b,&a);return 0;
}
`, ""},
	"Reduction": {"ieee754.c", "if (jz < 0) { adamic_unreachable(); }", `
int main(int argc,char **argv){
 adamic_start(argc,argv);
 volatile double x=1e300;
 printf("%.17g\n",adamic_math_sin(x));return 0;
}
`, "  // fdlibm __kernel_rem_pio2:"},
	"Map": {"map.c", "if (map == NULL) { adamic_unreachable(); }", `
int main(int argc,char **argv){
 adamic_start(argc,argv);
 adamic_map *map=adamic_map_new(false,false);
 adamic_map_iterator *it=adamic_map_iterate(map);adamic_value key,value;
 if(adamic_map_iterator_next(it,&key,&value))return 2;
 adamic_release(it);adamic_release(map);
 adamic_map_iterate(NULL);return 0;
}
`, ""},
}

// Build the normal sanitized archive once. Only the inspected C translation
// unit is embedded in each private control/mutant main, with its public names
// renamed so the whole shipping archive remains linked and unchanged.
func gccInvariantProduct(t *testing.T, kind string) string {
	t.Helper()
	spec, ok := gccInvariants[kind]
	if !ok {
		t.Fatal(kind)
	}
	options := Options{Sanitize: true}
	files, flags, compiler, version, err := runtimeInputs("", options, nil)
	if err != nil {
		t.Fatal(err)
	}
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "gcc-invariant-" + kind, Files: []string{"internal/native/runtime", "internal/native/gcc_invariants_test.go", "internal/native/native.go", "internal/native/library.go", "internal/native/build_runtime.go", "internal/native/target.go"}, Flags: portableKeyParts(files, flags, compiler, headers), Toolchain: []string{buildcache.Tool(compiler, "--version"), buildcache.Tool(archiverName(compiler), "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		cache, err := cacheDirectory("gcc-invariant")
		if err != nil {
			return err
		}
		library, err := cachedRuntime(files, flags, compiler, version, cache)
		if err != nil {
			return err
		}
		data, err := runtime.ReadFile("runtime/" + spec.file)
		if err != nil {
			return err
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
			for _, match := range regexp.MustCompile(`(?m)^[A-Za-z_][^;\n]*?\b(adamic_[A-Za-z0-9_]+)\([^;\n]*\) \{`).FindAllStringSubmatch(text, -1) {
				fmt.Fprintf(&renames, "#define %s probe_%s\n", match[1], match[1])
			}
			source = "#include <stdio.h>\n" + renames.String() + source + spec.main
			if err := BuildWithRuntime(source, filepath.Join(directory, variant), options, library); err != nil {
				return err
			}
		}
		return nil
	})
}
func checkGCCInvariant(t *testing.T, kind string, mutant bool) {
	t.Helper()
	variant := "control"
	if mutant {
		variant = "mutant"
	}
	result := inlineArrayRun(t, filepath.Join(gccInvariantProduct(t, kind), variant))
	stop := result.exit == 70 && strings.Contains(string(result.stderr), "compiler bug: a function ended without returning")
	if !mutant {
		if !stop || len(result.stdout) != 0 {
			t.Fatalf("missing loud invariant stop: exit %d stdout %q stderr %s", result.exit, result.stdout, result.stderr)
		}
	} else {
		if stop {
			t.Fatal("guard-removal mutant survived")
		}
		if kind == "Reduction" {
			if result.exit != 0 || len(result.stdout) == 0 {
				t.Fatalf("reduction mutant must return a wrong result: %+v", result)
			}
		} else if result.exit != 1 || !strings.Contains(string(result.stderr), "runtime error:") {
			t.Fatalf("missing UBSan witness: exit %d %s", result.exit, result.stderr)
		}
	}
	t.Logf("%s %s: exit %d stdout %q", kind, variant, result.exit, result.stdout)
}
func TestProduct_GCCInvariantDTOA(t *testing.T) { t.Parallel(); gccInvariantProduct(t, "DTOA") }
func TestProduct_GCCInvariantReduction(t *testing.T) {
	t.Parallel()
	gccInvariantProduct(t, "Reduction")
}
func TestProduct_GCCInvariantMap(t *testing.T) { t.Parallel(); gccInvariantProduct(t, "Map") }
func TestGCCInvariantDTOA(t *testing.T)        { t.Parallel(); checkGCCInvariant(t, "DTOA", false) }
func TestGCCInvariantReduction(t *testing.T)   { t.Parallel(); checkGCCInvariant(t, "Reduction", false) }
func TestGCCInvariantMap(t *testing.T)         { t.Parallel(); checkGCCInvariant(t, "Map", false) }
func TestGCCInvariantDTOAMutant(t *testing.T)  { t.Parallel(); checkGCCInvariant(t, "DTOA", true) }
func TestGCCInvariantReductionMutant(t *testing.T) {
	t.Parallel()
	checkGCCInvariant(t, "Reduction", true)
}
func TestGCCInvariantMapMutant(t *testing.T) { t.Parallel(); checkGCCInvariant(t, "Map", true) }
