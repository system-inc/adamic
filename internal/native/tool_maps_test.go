package native

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// An object compiled with debug info names no part of where the compiler is installed: clang's own headers are among
// the include directories it records (.debug_line_str), and clangIn maps their installation out, so a key, which never
// names where the tools sit, builds one set of bytes wherever they do (#8wgk65j). The same compile run without the maps
// is the control: it must name the installation, or this box's clang records nothing the check could catch, and the
// test says so and skips. Mutant: toolMaps returning nothing, red.
func TestAnObjectNamesNotWhereTheCompilerIsInstalled(test *testing.T) {
	test.Parallel()
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	installation := compilerInstallation(compiler)
	if installation == "" {
		test.Skipf("clang at %s has no installation directory to map", compiler)
	}
	directory := test.TempDir()
	// stddef.h is one of clang's own headers, from its resource directory under the installation.
	source := "#include <stddef.h>\nsize_t adamic_probe(size_t value) { return value + sizeof(ptrdiff_t); }\n"
	if writeError := os.WriteFile(filepath.Join(directory, "probe.c"), []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	arguments := append(Flags(Options{Sanitize: true}), prefixMaps(directory, "/adamic-probe")...)
	compile := func(object string, mapped bool) []byte {
		test.Helper()
		all := append(append([]string{}, arguments...), "-c", "probe.c", "-o", object)
		command := toolchain.CommandIn(directory, compiler, all...)
		if mapped {
			command = clangIn(directory, compiler, all...)
		}
		if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
			test.Fatalf("clang: %v\n%s", combinedOutputError, output)
		}
		contents, readError := os.ReadFile(filepath.Join(directory, object))
		if readError != nil {
			test.Fatal(readError)
		}
		return contents
	}
	if control := compile("control.o", false); !bytes.Contains(control, []byte(installation)) {
		test.Skipf(
			"this clang records no path under its installation %s in debug info, so there is nothing to map",
			installation,
		)
	}
	if mapped := compile("mapped.o", true); bytes.Contains(mapped, []byte(installation)) {
		test.Fatalf(
			"the object names where clang is installed (%s), so its bytes depend on where the tools sit",
			installation,
		)
	}
	test.Logf("the unmapped object names %s; the mapped one doesn't", installation)
}

// toolMaps map a sysroot named in the arguments too, since a wasm32 build's sysroot sits under the tools directory and
// its headers are recorded the same way. Mutant: the --sysroot loop dropped.
func TestToolMapsTakeTheSysrootOut(test *testing.T) {
	test.Parallel()
	sysroot := test.TempDir()
	maps := toolMaps("/nonexistent/bin/clang", []string{"-O2", "--sysroot=" + sysroot})
	want := "-ffile-prefix-map=" + sysroot + "=/adamic-sysroot"
	for _, flag := range maps {
		if flag == want {
			return
		}
	}
	test.Fatalf("maps %q, want %q among them", maps, want)
}

// A sanitized object's UBSan source locations name a header by its file alone, never the directory it was found in: no
// prefix map reaches those strings, and a sysroot under the tools directory put its root in every object that checks
// code from its headers (Hot's arm64 on #8wgk65j: 60 string.h locations naming <root>/llvm/bin/../../sysroot). Here the
// header sits in a directory of its own outside the compile's, found by an absolute -I as a sysroot's headers are, with
// an inline function UBSan checks. Mutant: -fsanitize-undefined-strip-path-components dropped from Flags, red.
func TestASanitizedObjectNamesNoHeadersDirectory(test *testing.T) {
	test.Parallel()
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	directory, headers := test.TempDir(), test.TempDir()
	header := "static inline int adamic_probe_add(int left, int right) { return left + right; }\n"
	if writeError := os.WriteFile(filepath.Join(headers, "probe_header.h"), []byte(header), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	source := "#include <probe_header.h>\nint adamic_probe(int value) { return adamic_probe_add(value, 1); }\n"
	if writeError := os.WriteFile(filepath.Join(directory, "probe.c"), []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	// The headers' directory is mapped as the compiler's installation is, so its debug info names it no more: only
	// UBSan's strings, which no map reaches, could still carry it.
	arguments := append(append(
		append(Flags(Options{Sanitize: true}), prefixMaps(directory, "/adamic-probe")...),
		prefixMaps(headers, "/adamic-headers")...,
	), "-I", headers, "-c", "probe.c", "-o", "probe.o")
	if output, combinedOutputError := clangIn(
		directory,
		compiler,
		arguments...,
	).CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("clang: %v\n%s", combinedOutputError, output)
	}
	object, readError := os.ReadFile(filepath.Join(directory, "probe.o"))
	if readError != nil {
		test.Fatal(readError)
	}
	if !bytes.Contains(object, []byte("probe_header.h")) {
		test.Fatal("the object names no probe_header.h at all, so UBSan checked nothing in it and this test can't fail")
	}
	if bytes.Contains(object, []byte(headers)) {
		test.Fatalf(
			"the sanitized object names the headers' directory %s, so its bytes depend on where they sit",
			headers,
		)
	}
}
