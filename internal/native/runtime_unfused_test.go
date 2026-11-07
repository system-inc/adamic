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

func runtimeUnfusedProduct(test *testing.T, profile string) string {
	test.Helper()
	options := Options{Release: profile == "release", Sanitize: profile == "sanitized", Count: profile == "counted"}
	options.ProgramRegion = true
	files, flags, compiler, _, err := gccAuditRuntimeInputs(options)
	if err != nil {
		test.Fatal(err)
	}
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		test.Fatal(err)
	}
	inputs := buildcache.Inputs{
		Name: "runtime-unfused-" + profile,
		Files: []string{
			"go.mod",
			"internal/native/runtime",
			"bridge/tsgo/tsgo.h",
			"internal/native/gcc_runtime_test.go",
			"internal/native/runtime_unfused_test.go",
			"internal/native/native.go",
			"internal/native/library.go",
			"internal/native/target.go",
		},
		Flags:     portableKeyParts(files, flags, compiler, headers),
		Toolchain: []string{buildcache.Tool(compiler, "--version"), buildcache.Tool("ar", "--version")},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		scratch := test.TempDir()
		quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
		wrapper := filepath.Join(scratch, "clang")
		script := fmt.Sprintf(
			"#!/bin/sh\nsource=\nfor argument do case \"$argument\" in *.c) source=${argument##*/} ;; esac; done\nif [ -n \"$source\" ]; then printf '%%s\\n' \"$@\" > %s/\"$source\"; fi\nexec %s \"$@\"\n",
			quote(scratch),
			quote(compiler),
		)
		if err := os.WriteFile(wrapper, []byte(script), 0700); err != nil {
			return err
		}
		jobserver, err := defaultJobserver()
		if err != nil {
			return err
		}
		if err := compileRuntime(
			files,
			flags,
			wrapper,
			directory,
			4,
			jobserver,
			"runtime-unfused-"+profile,
		); err != nil {
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
			mutant := slices.DeleteFunc(
				slices.Clone(flags),
				func(s string) bool { return strings.HasPrefix(s, "-ffp-contract=") },
			)
			object := filepath.Join(scratch, "dtoa-mutant.o")
			staticsCompile(test, wrapper, append(
				append(mutant, prefixMaps(directory, "/adamic-runtime")...),
				"-I",
				directory,
				"-c",
				filepath.Join(directory, "dtoa.c"),
				"-o",
				object,
			)...)
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
func checkRuntimeUnfused(test *testing.T, profile string) {
	test.Helper()
	data, err := os.ReadFile(filepath.Join(runtimeUnfusedProduct(test, profile), "audit"))
	if err != nil {
		test.Fatal(err)
	}
	if err := checkRuntimeUnfusedAudit(string(data)); err != nil {
		test.Fatal(err)
	}
	test.Logf("all %d runtime translation units: -ffp-contract=off (%s)", len(gccRuntimeUnits), profile)
}
func TestProduct_RuntimeUnfusedRelease(test *testing.T) {
	test.Parallel()
	runtimeUnfusedProduct(test, "release")
}
func TestProduct_RuntimeUnfusedSanitized(test *testing.T) {
	test.Parallel()
	runtimeUnfusedProduct(test, "sanitized")
}
func TestProduct_RuntimeUnfusedCounted(test *testing.T) {
	test.Parallel()
	runtimeUnfusedProduct(test, "counted")
}
func TestRuntimeCompilesEveryUnitUnfusedRelease(test *testing.T) {
	test.Parallel()
	checkRuntimeUnfused(test, "release")
}
func TestRuntimeCompilesEveryUnitUnfusedSanitized(test *testing.T) {
	test.Parallel()
	checkRuntimeUnfused(test, "sanitized")
}
func TestRuntimeCompilesEveryUnitUnfusedCounted(test *testing.T) {
	test.Parallel()
	checkRuntimeUnfused(test, "counted")
}

func TestRuntimeUnfusedOmissionMutant(test *testing.T) {
	test.Parallel()
	data, err := os.ReadFile(filepath.Join(runtimeUnfusedProduct(test, "release"), "omission-audit"))
	if err != nil {
		test.Fatal(err)
	}
	err = checkRuntimeUnfusedAudit(string(data))
	if err == nil || !strings.Contains(err.Error(), "dtoa.c") {
		test.Fatalf("actual dtoa.c compile omitted contraction flag but audit accepted it: %v", err)
	}
	test.Log(err)
}
