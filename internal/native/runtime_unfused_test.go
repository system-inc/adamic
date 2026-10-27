package native

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func runtimeUnfusedProduct(t *testing.T, profile string) string {
	t.Helper()
	options := Options{Release: profile == "release", Sanitize: profile == "sanitized", Count: profile == "counted"}
	options.ProgramRegion = true
	files, flags, compiler, _, err := gccAuditRuntimeInputs(options)
	if err != nil {
		t.Fatal(err)
	}
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "runtime-unfused-" + profile, Files: []string{"go.mod", "internal/native/runtime", "bridge/tsgo/tsgo.h", "internal/native/gcc_runtime_test.go", "internal/native/runtime_unfused_test.go", "internal/native/native.go", "internal/native/library.go", "internal/native/target.go"}, Flags: portableKeyParts(files, flags, compiler, headers), Toolchain: []string{buildcache.Tool(compiler, "--version"), buildcache.Tool("ar", "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		scratch := t.TempDir()
		quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
		wrapper := filepath.Join(scratch, "clang")
		script := fmt.Sprintf("#!/bin/sh\nsource=\nfor argument do case \"$argument\" in *.c) source=${argument##*/} ;; esac; done\nif [ -n \"$source\" ]; then printf '%%s\\n' \"$@\" > %s/\"$source\"; fi\nexec %s \"$@\"\n", quote(scratch), quote(compiler))
		if err := os.WriteFile(wrapper, []byte(script), 0700); err != nil {
			return err
		}
		jobserver, err := defaultJobserver()
		if err != nil {
			return err
		}
		if err := compileRuntime(files, flags, wrapper, directory, 4, jobserver, "runtime-unfused-"+profile); err != nil {
			return err
		}
		var audit strings.Builder
		for _, file := range files {
			if !strings.HasSuffix(file.name, ".c") {
				continue
			}
			args, err := os.ReadFile(filepath.Join(scratch, file.name))
			if err != nil {
				return err
			}
			// The last contraction option wins. Record the effective option, not
			// just the presence of an earlier flag overridden by a later one.
			contraction := "missing"
			for _, arg := range strings.Split(string(args), "\n") {
				if strings.HasPrefix(arg, "-ffp-contract=") {
					contraction = arg
				}
			}
			fmt.Fprintf(&audit, "%s\t%s\n", file.name, contraction)
		}
		if err := os.WriteFile(filepath.Join(directory, "audit"), []byte(audit.String()), 0600); err != nil {
			return err
		}
		if profile == "release" {
			// Compile one real unit without the flag. Reuse every other object;
			// this is an omission in a compiler invocation, not a fabricated log.
			mutant := slices.DeleteFunc(slices.Clone(flags), func(s string) bool { return strings.HasPrefix(s, "-ffp-contract=") })
			object := filepath.Join(scratch, "dtoa-mutant.o")
			staticsCompile(t, wrapper, append(append(mutant, prefixMaps(directory, "/adamic-runtime")...), "-I", directory, "-c", filepath.Join(directory, "dtoa.c"), "-o", object)...)
			actual, err := os.ReadFile(filepath.Join(scratch, "dtoa.c"))
			if err != nil {
				return err
			}
			contraction := "missing"
			for _, arg := range strings.Split(string(actual), "\n") {
				if strings.HasPrefix(arg, "-ffp-contract=") {
					contraction = arg
				}
			}
			omitted := strings.Replace(audit.String(), "dtoa.c\t-ffp-contract=off\n", "dtoa.c\t"+contraction+"\n", 1)
			if err := os.WriteFile(filepath.Join(directory, "omission-audit"), []byte(omitted), 0600); err != nil {
				return err
			}

		}
		return nil
	})
}

func checkRuntimeUnfusedAudit(data string) error {
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(data), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 || parts[1] != "-ffp-contract=off" || seen[parts[0]] {
			return fmt.Errorf("unprotected runtime compile: %q", line)
		}
		seen[parts[0]] = true
	}
	for _, name := range gccRuntimeUnits {
		if !seen[name] {
			return fmt.Errorf("missing runtime compile: %s", name)
		}
	}
	if len(seen) != len(gccRuntimeUnits) {
		return fmt.Errorf("unexpected runtime compile")
	}
	return nil
}
func checkRuntimeUnfused(t *testing.T, profile string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(runtimeUnfusedProduct(t, profile), "audit"))
	if err != nil {
		t.Fatal(err)
	}
	if err := checkRuntimeUnfusedAudit(string(data)); err != nil {
		t.Fatal(err)
	}
	t.Logf("all %d runtime translation units: -ffp-contract=off (%s)", len(gccRuntimeUnits), profile)
}
func TestProduct_RuntimeUnfusedRelease(t *testing.T) {
	t.Parallel()
	runtimeUnfusedProduct(t, "release")
}
func TestProduct_RuntimeUnfusedSanitized(t *testing.T) {
	t.Parallel()
	runtimeUnfusedProduct(t, "sanitized")
}
func TestProduct_RuntimeUnfusedCounted(t *testing.T) {
	t.Parallel()
	runtimeUnfusedProduct(t, "counted")
}
func TestRuntimeCompilesEveryUnitUnfusedRelease(t *testing.T) {
	t.Parallel()
	checkRuntimeUnfused(t, "release")
}
func TestRuntimeCompilesEveryUnitUnfusedSanitized(t *testing.T) {
	t.Parallel()
	checkRuntimeUnfused(t, "sanitized")
}
func TestRuntimeCompilesEveryUnitUnfusedCounted(t *testing.T) {
	t.Parallel()
	checkRuntimeUnfused(t, "counted")
}

func TestRuntimeUnfusedOmissionMutant(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join(runtimeUnfusedProduct(t, "release"), "omission-audit"))
	if err != nil {
		t.Fatal(err)
	}
	err = checkRuntimeUnfusedAudit(string(data))
	if err == nil || !strings.Contains(err.Error(), "dtoa.c") {
		t.Fatalf("actual dtoa.c compile omitted contraction flag but audit accepted it: %v", err)
	}
	t.Log(err)
}
