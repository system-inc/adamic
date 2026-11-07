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
	files, flags, compiler, _, gccAuditRuntimeInputsError := gccAuditRuntimeInputs(options)
	if gccAuditRuntimeInputsError != nil {
		test.Fatal(gccAuditRuntimeInputsError)
	}
	headers, headersReadError := headersRead(files, flags, compiler)
	if headersReadError != nil {
		test.Fatal(headersReadError)
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
		quote := func(text string) string { return "'" + strings.ReplaceAll(text, "'", "'\\''") + "'" }
		wrapper := filepath.Join(scratch, "clang")
		script := fmt.Sprintf(
			"#!/bin/sh\nsource=\nfor argument do case \"$argument\" in *.c) source=${argument##*/} ;; esac; done\nif [ -n \"$source\" ]; then printf '%%s\\n' \"$@\" > %s/\"$source\"; fi\nexec %s \"$@\"\n",
			quote(scratch),
			quote(compiler),
		)
		if writeError := os.WriteFile(wrapper, []byte(script), 0700); writeError != nil {
			return writeError
		}
		jobserver, defaultJobserverError := defaultJobserver()
		if defaultJobserverError != nil {
			return defaultJobserverError
		}
		if compileRuntimeError := compileRuntime(
			files,
			flags,
			wrapper,
			directory,
			4,
			jobserver,
			"runtime-unfused-"+profile,
		); compileRuntimeError != nil {
			return compileRuntimeError
		}
		var audit strings.Builder
		for _, file := range files {
			if !strings.HasSuffix(file.name, ".c") {
				continue
			}
			args, readError := os.ReadFile(filepath.Join(scratch, file.name))
			if readError != nil {
				return readError
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
		if writeError := os.WriteFile(
			filepath.Join(directory, "audit"),
			[]byte(audit.String()),
			0600,
		); writeError != nil {
			return writeError
		}
		if profile == "release" {
			// Compile one real unit without the flag. Reuse every other object;
			// this is an omission in a compiler invocation, not a fabricated log.
			mutant := slices.DeleteFunc(
				slices.Clone(flags),
				func(text string) bool { return strings.HasPrefix(text, "-ffp-contract=") },
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
			actual, readError := os.ReadFile(filepath.Join(scratch, "dtoa.c"))
			if readError != nil {
				return readError
			}
			contraction := "missing"
			for _, arg := range strings.Split(string(actual), "\n") {
				if strings.HasPrefix(arg, "-ffp-contract=") {
					contraction = arg
				}
			}
			omitted := strings.Replace(audit.String(), "dtoa.c\t-ffp-contract=off\n", "dtoa.c\t"+contraction+"\n", 1)
			if writeError := os.WriteFile(
				filepath.Join(directory, "omission-audit"),
				[]byte(omitted),
				0600,
			); writeError != nil {
				return writeError
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
	data, readError := os.ReadFile(filepath.Join(runtimeUnfusedProduct(test, profile), "audit"))
	if readError != nil {
		test.Fatal(readError)
	}
	if checkRuntimeUnfusedAuditError := checkRuntimeUnfusedAudit(string(data)); checkRuntimeUnfusedAuditError != nil {
		test.Fatal(checkRuntimeUnfusedAuditError)
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
	data, readError := os.ReadFile(filepath.Join(runtimeUnfusedProduct(test, "release"), "omission-audit"))
	if readError != nil {
		test.Fatal(readError)
	}
	checkRuntimeUnfusedAuditError := checkRuntimeUnfusedAudit(string(data))
	if checkRuntimeUnfusedAuditError == nil || !strings.Contains(checkRuntimeUnfusedAuditError.Error(), "dtoa.c") {
		test.Fatalf(
			"actual dtoa.c compile omitted contraction flag but audit accepted it: %v",
			checkRuntimeUnfusedAuditError,
		)
	}
	test.Log(checkRuntimeUnfusedAuditError)
}
