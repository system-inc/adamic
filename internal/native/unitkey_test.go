package native

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Each key test changes one input of a unit's compile alone and asserts its key moves, and where a compile is cheap
// that the object built after the change is built fresh and holds the change. The tests read the key as newUnitKeying
// reads it for a build, through the same clang the units compile with.

// unitKeyFlags are the flags a key test compiles with when the flags aren't what it changes.
var unitKeyFlags = []string{"-O2"}

// unitKeyCompiler is the clang the units compile with, as buildUnitsWithLibrary finds it.
func unitKeyCompiler(test *testing.T) string {
	test.Helper()
	compiler, err := toolchain.Path("clang")
	if err != nil {
		test.Fatal(err)
	}
	return compiler
}

// unitKeyEnvironment is toolchain.Env without any setting the key reads or refuses, so a test that sets one sets it
// alone.
func unitKeyEnvironment() []string {
	var environment []string
	for _, setting := range toolchain.Env() {
		name, _, _ := strings.Cut(setting, "=")
		if slices.Contains(unitEnvironmentKeyed, name) ||
			slices.Contains(unitEnvironmentRefused, name) ||
			strings.HasPrefix(name, "CCC_") {
			continue
		}
		environment = append(environment, setting)
	}
	return environment
}

// testUnitKeying is newUnitKeying's keying for these inputs, failing the test when the key can't account for them.
func testUnitKeying(
	test *testing.T,
	first string,
	common []runtimeFile,
	flags []string,
	compiler string,
	environment []string,
) *unitKeying {
	test.Helper()
	keying, err := newUnitKeying(first, common, flags, compiler, environment, false)
	if err != nil {
		test.Fatal(err)
	}
	return keying
}

// writeUnitKeyFile writes contents at path, making the directories it sits in.
func writeUnitKeyFile(test *testing.T, path string, contents string) {
	test.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		test.Fatal(err)
	}
}

// observeUnitKey links object, which defines adamic_probe, with a main that prints what it returns, runs it, and is
// what it printed.
func observeUnitKey(test *testing.T, compiler string, object string) string {
	test.Helper()
	directory := test.TempDir()
	main := filepath.Join(directory, "main.c")
	writeUnitKeyFile(
		test,
		main,
		"#include <stdio.h>\nint adamic_probe(void);\nint main(void) { printf(\"%d\\n\", adamic_probe()); return 0; }\n",
	)
	binary := filepath.Join(directory, "probe")
	if output, err := toolchain.Command(compiler, main, object, "-o", binary).CombinedOutput(); err != nil {
		test.Fatalf("link: %v\n%s", err, output)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		test.Fatalf("run: %v\n%s", err, output)
	}
	return strings.TrimSpace(string(output))
}

// unitKeyProbe is a unit that returns the value of the macro named value, which its headers define.
func unitKeyProbe(value string) compilationUnit {
	return compilationUnit{"probe.c", unitFirst + "int adamic_probe(void) { return " + value + "; }\n"}
}

// compileUnitKey compiles unit under keying in a cache of directory's, failing the test on an error, and is the object.
func compileUnitKey(
	test *testing.T,
	unit compilationUnit,
	common []runtimeFile,
	flags []string,
	compiler string,
	keying *unitKeying,
	directory string,
) string {
	test.Helper()
	object, err := compileUnit(
		unit,
		common,
		flags,
		compiler,
		keying,
		filepath.Join(directory, "cache"),
		directory,
		nil,
		false,
	)
	if err != nil {
		test.Fatal(err)
	}
	return object
}

// The unit's own bytes and its name are both the key's: another body, or the same body under another name, is another
// object.
func TestUnitKeyHoldsTheUnit(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	common := []runtimeFile{{"units.h", []byte("#define VALUE 1\n")}}
	keying := testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, unitKeyEnvironment())
	unit := unitKeyProbe("1")
	changed := unitKeyProbe("2")
	renamed := compilationUnit{"renamed.c", unit.source}
	if keying.key(changed) == keying.key(unit) {
		test.Fatal("a unit's bytes changed and its key held")
	}
	if keying.key(renamed) == keying.key(unit) {
		test.Fatal("a unit's name changed and its key held")
	}
	directory := test.TempDir()
	first := compileUnitKey(test, unit, common, unitKeyFlags, compiler, keying, directory)
	second := compileUnitKey(test, changed, common, unitKeyFlags, compiler, keying, directory)
	if first == second {
		test.Fatal("a changed unit was taken from the cache")
	}
	if got := observeUnitKey(test, compiler, second); got != "2" {
		test.Fatalf("a changed unit's object returned %s, want 2", got)
	}
}

// Every file beside the unit is the key's by name and bytes: a header the unit reads, and a file nothing includes.
func TestUnitKeyHoldsCommonFiles(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	unit := unitKeyProbe("VALUE")
	common := []runtimeFile{{"units.h", []byte("#define VALUE 1\n")}}
	keying := testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, environment)
	changed := []runtimeFile{{"units.h", []byte("#define VALUE 2\n")}}
	changedKeying := testUnitKeying(test, unitFirst, changed, unitKeyFlags, compiler, environment)
	if changedKeying.key(unit) == keying.key(unit) {
		test.Fatal("a common header's bytes changed and the key held")
	}
	// Complete before minimal: a file nothing includes still moves the key.
	added := append(slices.Clone(common), runtimeFile{"unused.h", []byte("#define UNUSED 1\n")})
	if testUnitKeying(test, unitFirst, added, unitKeyFlags, compiler, environment).key(unit) == keying.key(unit) {
		test.Fatal("a common file was added and the key held")
	}
	directory := test.TempDir()
	if got := observeUnitKey(
		test,
		compiler,
		compileUnitKey(test, unit, common, unitKeyFlags, compiler, keying, directory),
	); got != "1" {
		test.Fatalf("the unit returned %s, want 1", got)
	}
	if got := observeUnitKey(
		test,
		compiler,
		compileUnitKey(test, unit, changed, unitKeyFlags, compiler, changedKeying, directory),
	); got != "2" {
		test.Fatalf("after its header changed the unit returned %s, want 2", got)
	}
}

// A header found on clang's search path is the key's by its bytes: a probe whose first line includes it from an
// -isystem directory keys apart when it changes.
func TestUnitKeyHoldsSystemHeaders(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	system := test.TempDir()
	planted := filepath.Join(system, "planted.h")
	writeUnitKeyFile(test, planted, "#define PLANTED 1\n")
	first := "#include <planted.h>\n"
	unit := compilationUnit{"probe.c", first + "int adamic_probe(void) { return PLANTED; }\n"}
	flags := append(slices.Clone(unitKeyFlags), "-isystem", system)
	keying := testUnitKeying(test, first, nil, flags, compiler, environment)
	directory := test.TempDir()
	if got := observeUnitKey(
		test,
		compiler,
		compileUnitKey(test, unit, nil, flags, compiler, keying, directory),
	); got != "1" {
		test.Fatalf("the unit returned %s, want 1", got)
	}
	writeUnitKeyFile(test, planted, "#define PLANTED 2\n")
	changed := testUnitKeying(test, first, nil, flags, compiler, environment)
	if changed.key(unit) == keying.key(unit) {
		test.Fatal("a system header's bytes changed and the key held")
	}
	if got := observeUnitKey(
		test,
		compiler,
		compileUnitKey(test, unit, nil, flags, compiler, changed, directory),
	); got != "2" {
		test.Fatalf("after its system header changed the unit returned %s, want 2", got)
	}
}

// A header clang finds in a later search directory is shadowed by one planted under its name in an earlier one, and
// the plant moves the key.
func TestUnitKeyHoldsShadowing(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	earlier, later := test.TempDir(), test.TempDir()
	writeUnitKeyFile(test, filepath.Join(later, "shadow.h"), "#define SHADOW 1\n")
	common := []runtimeFile{{"units.h", []byte("#include <shadow.h>\n")}}
	flags := append(slices.Clone(unitKeyFlags), "-isystem", earlier, "-isystem", later)
	unit := unitKeyProbe("SHADOW")
	keying := testUnitKeying(test, unitFirst, common, flags, compiler, environment)
	directory := test.TempDir()
	if got := observeUnitKey(
		test,
		compiler,
		compileUnitKey(test, unit, common, flags, compiler, keying, directory),
	); got != "1" {
		test.Fatalf("the unit returned %s, want 1", got)
	}
	writeUnitKeyFile(test, filepath.Join(earlier, "shadow.h"), "#define SHADOW 2\n")
	shadowed := testUnitKeying(test, unitFirst, common, flags, compiler, environment)
	if shadowed.key(unit) == keying.key(unit) {
		test.Fatal("a header planted in an earlier search directory shadowed a later one and the key held")
	}
	if got := observeUnitKey(
		test,
		compiler,
		compileUnitKey(test, unit, common, flags, compiler, shadowed, directory),
	); got != "2" {
		test.Fatalf("after the shadow was planted the unit returned %s, want 2", got)
	}
}

// A header a __has_include asks after is the key's even while it's absent: planting it moves the key, and the build
// after the plant compiles fresh and sees it.
func TestUnitKeyHoldsHasInclude(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	earlier, later := test.TempDir(), test.TempDir()
	// units.h asks after maybe.h without including it, so only the __has_include names it.
	common := []runtimeFile{
		{"units.h", []byte("#if __has_include(<maybe.h>)\n#define MAYBE 7\n#else\n#define MAYBE 0\n#endif\n")},
	}
	flags := append(slices.Clone(unitKeyFlags), "-isystem", earlier, "-isystem", later)
	unit := unitKeyProbe("MAYBE")
	keying := testUnitKeying(test, unitFirst, common, flags, compiler, environment)
	// The table names the absent header itself, wherever clang would look: clang's -M lists a file a __has_include
	// found, so a plant moves the key through the closure too, but an absent one leaves only this line (Node's mutant:
	// no __has_include name read).
	for _, directory := range []string{earlier, later} {
		if line := "candidate " +
			filepath.Join(directory, "maybe.h") +
			" absent"; !slices.Contains(keying.parts, line) {
			test.Fatalf("the key doesn't hold %q", line)
		}
	}
	directory := test.TempDir()
	absent := compileUnitKey(test, unit, common, flags, compiler, keying, directory)
	if got := observeUnitKey(test, compiler, absent); got != "0" {
		test.Fatalf("with maybe.h absent the unit returned %s, want 0", got)
	}
	writeUnitKeyFile(test, filepath.Join(later, "maybe.h"), "#define MAYBE 7\n")
	planted := testUnitKeying(test, unitFirst, common, flags, compiler, environment)
	if planted.key(unit) == keying.key(unit) {
		test.Fatal("a header a __has_include asks after was planted and the key held")
	}
	present := compileUnitKey(test, unit, common, flags, compiler, planted, directory)
	if present == absent {
		test.Fatal("the object compiled before the plant was reused")
	}
	if got := observeUnitKey(test, compiler, present); got != "7" {
		test.Fatalf("with maybe.h planted the unit returned %s, want 7", got)
	}
}

// #include_next and __has_include_next look in the search directories after the including header's own, and a header
// planted in any of them moves the key: one that shadows the next inner.h, and one a __has_include_next asks after.
func TestUnitKeyHoldsIncludeNext(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	// wrap.h asks after later.h without including it, so only the __has_include_next names it, and only its
	// #include_next names inner.h.
	wrap := "#if __has_include_next(<later.h>)\n#define LATER 5\n#else\n#define LATER 0\n#endif\n#include_next <inner.h>\n"
	for _, plan := range []struct {
		name, plant, value string
		before, after      string
	}{
		{"IncludeNext", "middle/inner.h", "INNER", "3", "5"},
		{"HasIncludeNext", "last/later.h", "LATER", "0", "5"},
	} {
		test.Run(plan.name, func(test *testing.T) {
			test.Parallel()
			root := test.TempDir()
			writeUnitKeyFile(test, filepath.Join(root, "first", "wrap.h"), wrap)
			writeUnitKeyFile(test, filepath.Join(root, "last", "inner.h"), "#define INNER 3\n")
			if err := os.MkdirAll(filepath.Join(root, "middle"), 0o755); err != nil {
				test.Fatal(err)
			}
			common := []runtimeFile{{"units.h", []byte("#include <wrap.h>\n")}}
			flags := slices.Clone(unitKeyFlags)
			for _, name := range []string{"first", "middle", "last"} {
				flags = append(flags, "-isystem", filepath.Join(root, name))
			}
			unit := unitKeyProbe(plan.value)
			keying := testUnitKeying(test, unitFirst, common, flags, compiler, environment)
			directory := test.TempDir()
			before := compileUnitKey(test, unit, common, flags, compiler, keying, directory)
			if got := observeUnitKey(test, compiler, before); got != plan.before {
				test.Fatalf("before the plant the unit returned %s, want %s", got, plan.before)
			}
			writeUnitKeyFile(test, filepath.Join(root, plan.plant), "#define "+plan.value+" 5\n")
			planted := testUnitKeying(test, unitFirst, common, flags, compiler, environment)
			if planted.key(unit) == keying.key(unit) {
				test.Fatalf("%s was planted and the key held", plan.plant)
			}
			after := compileUnitKey(test, unit, common, flags, compiler, planted, directory)
			if after == before {
				test.Fatal("the object compiled before the plant was reused")
			}
			if got := observeUnitKey(test, compiler, after); got != plan.after {
				test.Fatalf("after the plant the unit returned %s, want %s", got, plan.after)
			}
		})
	}
}

// A header the key can't follow refuses the build's key: an include by a macro, whose name the key can't read, a
// header reading the time of the compile, and an #embed or a __has_embed.
func TestUnitKeyRefusesWhatItCannotFollow(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	for _, plan := range []struct {
		name   string
		common []runtimeFile
		why    string
	}{
		// x.h is beside units.h, so clang resolves the macro's include and only the key refuses.
		{
			"Macro",
			[]runtimeFile{{"units.h", []byte("#define H \"x.h\"\n#include H\n")}, {"x.h", []byte("#define X 1\n")}},
			"by a macro",
		},
		{"Date", []runtimeFile{{"units.h", []byte("static const char *adamic_when = __DATE__;\n")}}, "reads the time"},
		{"Time", []runtimeFile{{"units.h", []byte("static const char *adamic_when = __TIME__;\n")}}, "reads the time"},
		{
			"Timestamp",
			[]runtimeFile{{"units.h", []byte("static const char *adamic_when = __TIMESTAMP__;\n")}},
			"reads the time",
		},
		// clang 20 reads a file for #embed and asks after one with __has_embed, neither of which the table follows
		// (Node's M3).
		{"Embed", []runtimeFile{
			{"units.h", []byte("static const unsigned char adamic_data[] = {\n#embed \"data.bin\"\n};\n")},
			{"data.bin", []byte("1")},
		}, "#embed or __has_embed"},
		{
			"HasEmbed",
			[]runtimeFile{{"units.h", []byte("#if __has_embed(\"absent.bin\")\n#define VALUE 1\n#endif\n")}},
			"#embed or __has_embed",
		},
	} {
		test.Run(plan.name, func(test *testing.T) {
			test.Parallel()
			_, err := newUnitKeying(unitFirst, plan.common, unitKeyFlags, compiler, unitKeyEnvironment(), false)
			if !errors.Is(err, errUncachedUnit) || !strings.Contains(err.Error(), plan.why) {
				test.Fatalf("want errUncachedUnit saying %q, got %v", plan.why, err)
			}
		})
	}
}

// A header that defines __has_include itself, as macOS's sys/cdefs.h does for a compiler without it, asks after no
// header there, so the key reads it; one that wraps __has_include around a macro's argument asks by a name the key
// can't read, and refuses. Mutant: the definition read as a question (every split build on a Mac compiled uncached).
func TestUnitKeyReadsAHasIncludeDefinition(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	fallback := []runtimeFile{{
		"units.h",
		[]byte("#ifndef __has_include\n#define __has_include(x) 0\n#endif\n#  undef  __has_include\n#define VALUE 1\n"),
	}}
	if _, err := newUnitKeying(unitFirst, fallback, unitKeyFlags, compiler, unitKeyEnvironment(), false); err != nil {
		test.Fatalf("a header defining __has_include refused the key: %v", err)
	}
	wrapped := []runtimeFile{{"units.h", []byte("#define ASKS(x) __has_include(x)\n#define VALUE 1\n")}}
	if _, err := newUnitKeying(
		unitFirst,
		wrapped,
		unitKeyFlags,
		compiler,
		unitKeyEnvironment(),
		false,
	); !errors.Is(err, errUncachedUnit) || !strings.Contains(err.Error(), "by a macro") {
		test.Fatalf("a __has_include of a macro's argument: want errUncachedUnit saying by a macro, got %v", err)
	}
}

// A unit's own source the key can't take by its bytes compiles uncached: one that doesn't begin with the probe's line,
// one with a directive past it, one reading the time or asking after a header.
func TestUnitSourceRefusal(test *testing.T) {
	test.Parallel()
	for _, plan := range []struct {
		name, source, why string
	}{
		{"Keyed", unitFirst + "int adamic_a;\n", ""},
		{"OtherFirst", "#include \"other.h\"\nint adamic_a;\n", "doesn't begin with the line"},
		{"Directive", unitFirst + "int adamic_a;\n#define ADAMIC_B 1\n", "holds a directive past its first line"},
		{"Time", unitFirst + "const char *adamic_t = __TIME__;\n", "reads __TIME__"},
		{"HasInclude", unitFirst + "int adamic_a = __has_include(<stdio.h>);\n", "reads __has_include"},
		{"HasEmbed", unitFirst + "int adamic_a = __has_embed(\"absent.bin\");\n", "reads __has_embed"},
	} {
		test.Run(plan.name, func(test *testing.T) {
			test.Parallel()
			why := unitSourceRefusal(compilationUnit{"unit.c", plan.source}, unitFirst)
			if plan.why == "" && why != "" || !strings.Contains(why, plan.why) {
				test.Fatalf("unitSourceRefusal said %q, want %q", why, plan.why)
			}
		})
	}
	// A refused unit compiles uncached under a keying, and the cache holds nothing of it.
	compiler := unitKeyCompiler(test)
	common := []runtimeFile{{"units.h", []byte("#define VALUE 4\n")}}
	keying := testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, unitKeyEnvironment())
	directory := test.TempDir()
	unit := compilationUnit{"probe.c", unitFirst + "#define ADAMIC_OTHER 1\nint adamic_probe(void) { return VALUE; }\n"}
	object := compileUnitKey(test, unit, common, unitKeyFlags, compiler, keying, directory)
	if object != filepath.Join(directory, "uncached-probe.c.o") {
		test.Fatalf("a unit with a directive past its first line compiled to %s, not uncached", object)
	}
	if _, err := os.Stat(filepath.Join(directory, "cache")); !errors.Is(err, os.ErrNotExist) {
		test.Fatalf("a refused unit wrote the cache: %v", err)
	}
}

// copyUnitKeyProgram puts the program clang resolves to into directory, named clang and runnable, and is the copy. On
// macOS it's a hard link, the same file under a new name, since the kernel kills a copy its signature no longer covers
// and signing one would run codesign, a tool the gate's list doesn't declare (chg_qv6ny4dd's tools red): the link is
// still clang's program by its bytes, in a directory of the test's own for config files. Elsewhere it's a copy, so a
// test that rewrites it never writes the toolchain's clang through a link; macOS's caller rewrites nothing.
func copyUnitKeyProgram(test *testing.T, directory string, compiler string) string {
	test.Helper()
	program, err := filepath.EvalSymlinks(compiler)
	if err != nil {
		test.Fatal(err)
	}
	copied := filepath.Join(directory, "clang")
	if goruntime.GOOS == "darwin" {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			test.Fatal(err)
		}
		if err := os.Link(program, copied); err != nil {
			test.Skip("clang's program can't be linked beside the test, and a copy of it would need signing: ", err)
		}
		return copied
	}
	contents, err := os.ReadFile(program)
	if err != nil {
		test.Fatal(err)
	}
	writeUnitKeyFile(test, copied, string(contents))
	if err := os.Chmod(copied, 0o755); err != nil {
		test.Fatal(err)
	}
	return copied
}

// clang's build is the key's by what its program holds and the config files beside it, each read again when it changes
// under a running process: a config file written beside the program moves the key, as does one beside a link to it,
// and the program rewritten with another byte at its end moves it too. A script standing in for clang refuses the key,
// since the program it runs is one the key can't read. Mutants: the program's digest left out; the config files left
// out of the memo's identity.
// Not parallel: it publishes an executable copy before other tests fork.
func TestUnitKeyHoldsClangsBuild(test *testing.T) {
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	common := []runtimeFile{{"units.h", []byte("#define VALUE 1\n")}}
	unit := unitKeyProbe("VALUE")
	key := func(compiler string) string {
		test.Helper()
		return testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, environment).key(unit)
	}
	root, err := filepath.EvalSymlinks(test.TempDir())
	if err != nil {
		test.Fatal(err)
	}
	copied := copyUnitKeyProgram(test, filepath.Join(root, "copied"), compiler)
	plain := key(copied)
	writeUnitKeyFile(test, filepath.Join(root, "copied", "clang.cfg"), "-DADAMIC_CONFIGURED=1\n")
	if key(copied) == plain {
		test.Fatal("a config file written beside clang's program left the key")
	}
	if err := os.Remove(filepath.Join(root, "copied", "clang.cfg")); err != nil {
		test.Fatal(err)
	}
	if key(copied) != plain {
		test.Fatal("the config file removed, the key didn't come back")
	}
	linked := filepath.Join(root, "linked", "clang")
	if err := os.MkdirAll(filepath.Dir(linked), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.Symlink(copied, linked); err != nil {
		test.Fatal(err)
	}
	beside := key(linked)
	writeUnitKeyFile(test, filepath.Join(root, "linked", "clang.cfg"), "-DADAMIC_CONFIGURED=1\n")
	if key(linked) == beside {
		test.Fatal("a config file written beside a link to clang left the key")
	}
	script := filepath.Join(root, "script", "clang")
	writeUnitKeyFile(test, script, fmt.Sprintf("#!/bin/sh\nexec '%s' \"$@\"\n", compiler))
	if err := os.Chmod(script, 0o755); err != nil {
		test.Fatal(err)
	}
	if _, err := newUnitKeying(
		unitFirst,
		common,
		unitKeyFlags,
		script,
		environment,
		false,
	); !errors.Is(err, errUncachedUnit) || !strings.Contains(err.Error(), "neither ELF nor Mach-O") {
		test.Fatalf("a script standing in for clang: want errUncachedUnit, got %v", err)
	}
	// A byte past the program's end changes no instruction it runs, so it still runs, and is another program by its
	// bytes. macOS won't sign a program with bytes past its signature, so there this step is Linux's to show.
	if goruntime.GOOS == "darwin" {
		return
	}
	program, err := os.OpenFile(copied, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		test.Fatal(err)
	}
	if _, err := program.Write([]byte{0}); err != nil {
		test.Fatal(err)
	}
	if err := program.Close(); err != nil {
		test.Fatal(err)
	}
	if key(copied) == plain {
		test.Fatal("clang's program was rewritten and the key held")
	}
}

// Flags that use clang modules refuse the key: module maps and the module cache are reads the resolution table doesn't
// walk (Node's should 3).
func TestUnitKeyRefusesModules(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	common := []runtimeFile{{"units.h", []byte("#define VALUE 1\n")}}
	for _, flag := range []string{"-fmodules", "-fmodule-map-file=/adamic-none", "-fimplicit-module-maps"} {
		if _, err := newUnitKeying(
			unitFirst,
			common,
			append(slices.Clone(unitKeyFlags), flag),
			compiler,
			unitKeyEnvironment(),
			false,
		); !errors.Is(err, errUncachedUnit) || !strings.Contains(err.Error(), "modules") {
			test.Fatalf("%s: want errUncachedUnit saying modules, got %v", flag, err)
		}
	}
}

// The house's clang links no LLVM or clang library, whose code its digest wouldn't cover (Node's should 1), and a
// program that isn't there has no libraries to read.
func TestUnitKeyReadsClangsLibraries(test *testing.T) {
	test.Parallel()
	program, err := filepath.EvalSymlinks(unitKeyCompiler(test))
	if err != nil {
		test.Fatal(err)
	}
	libraries, err := programLibraries(program)
	if err != nil {
		test.Fatal(err)
	}
	for _, library := range libraries {
		if lower := strings.ToLower(library); strings.Contains(lower, "llvm") || strings.Contains(lower, "clang") {
			test.Fatalf("this box's clang links %s, so every split build here compiles uncached", library)
		}
	}
	if _, err := programLibraries(filepath.Join(test.TempDir(), "absent")); err == nil {
		test.Fatal("a program that isn't there read as having no libraries")
	}
}

// Every flag is the key's: another optimization level, and a -D added.
func TestUnitKeyHoldsFlags(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	common := []runtimeFile{{"units.h", []byte("#ifndef VALUE\n#define VALUE 1\n#endif\n")}}
	unit := unitKeyProbe("VALUE")
	keying := testUnitKeying(test, unitFirst, common, []string{"-O2"}, compiler, environment)
	if testUnitKeying(test, unitFirst, common, []string{"-O1"}, compiler, environment).key(unit) == keying.key(unit) {
		test.Fatal("-O1 keyed as -O2")
	}
	defined := []string{"-O2", "-DVALUE=6"}
	definedKeying := testUnitKeying(test, unitFirst, common, defined, compiler, environment)
	if definedKeying.key(unit) == keying.key(unit) {
		test.Fatal("a -D was added and the key held")
	}
	directory := test.TempDir()
	compileUnitKey(test, unit, common, []string{"-O2"}, compiler, keying, directory)
	if got := observeUnitKey(
		test,
		compiler,
		compileUnitKey(test, unit, common, defined, compiler, definedKeying, directory),
	); got != "6" {
		test.Fatalf("with -DVALUE=6 the unit returned %s, want 6", got)
	}
}

// The target is the key's: one unit keyed for two targets keys apart. clang needs no sysroot for either here, since the
// probe reads only the units' own header.
func TestUnitKeyHoldsTarget(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	common := []runtimeFile{{"units.h", []byte("#define VALUE 1\n")}}
	unit := unitKeyProbe("VALUE")
	x86 := testUnitKeying(test, unitFirst, common, []string{"-O2", "--target=x86_64-linux-gnu"}, compiler, environment)
	arm := testUnitKeying(test, unitFirst, common, []string{"-O2", "--target=aarch64-linux-gnu"}, compiler, environment)
	if x86.key(unit) == arm.key(unit) {
		test.Fatal("x86_64-linux-gnu keyed as aarch64-linux-gnu")
	}
}

// unitKeyKeyed and unitKeyRefused are the environment lists as this test pins them: a name dropped from either
// production list fails here.
var unitKeyKeyed = []string{"DEVELOPER_DIR", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET", "IPHONEOS_DEPLOYMENT_TARGET",
	"TVOS_DEPLOYMENT_TARGET", "WATCHOS_DEPLOYMENT_TARGET", "XROS_DEPLOYMENT_TARGET", "SOURCE_DATE_EPOCH", "TZ",
	"RC_DEBUG_OPTIONS", "RC_DEBUG_PREFIX_MAP", "COMPILER_PATH"}

var unitKeyRefused = []string{
	"CPATH",
	"C_INCLUDE_PATH",
	"CPLUS_INCLUDE_PATH",
	"OBJC_INCLUDE_PATH",
	"OBJCPLUS_INCLUDE_PATH",
}

// Every variable clang reads that can change an object is the key's by its value, and every one that adds a search
// directory or rewrites clang's arguments refuses the key.
func TestUnitKeyHoldsEnvironment(test *testing.T) {
	test.Parallel()
	if !slices.Equal(unitEnvironmentKeyed, unitKeyKeyed) {
		test.Fatalf("unitEnvironmentKeyed is %q, want %q", unitEnvironmentKeyed, unitKeyKeyed)
	}
	if !slices.Equal(unitEnvironmentRefused, unitKeyRefused) {
		test.Fatalf("unitEnvironmentRefused is %q, want %q", unitEnvironmentRefused, unitKeyRefused)
	}
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	common := []runtimeFile{{"units.h", []byte("#define VALUE 1\n")}}
	unit := unitKeyProbe("VALUE")
	key := testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, environment).key(unit)
	// Each value is one clang still compiles under, so the key reads it rather than refusing: macOS's developer
	// directory and SDK are where they already are.
	values := map[string]string{
		"MACOSX_DEPLOYMENT_TARGET":   "14.0",
		"IPHONEOS_DEPLOYMENT_TARGET": "17.0",
		"TVOS_DEPLOYMENT_TARGET":     "17.0",
		"WATCHOS_DEPLOYMENT_TARGET":  "10.0",
		"XROS_DEPLOYMENT_TARGET":     "1.0",
		"SOURCE_DATE_EPOCH":          "1",
		"TZ":                         "UTC",
		"RC_DEBUG_OPTIONS":           "1",
		"RC_DEBUG_PREFIX_MAP":        "/adamic-a=/adamic-b",
		"COMPILER_PATH":              filepath.Join(test.TempDir(), "none"),
		"DEVELOPER_DIR":              "/adamic-none",
		"SDKROOT":                    "/adamic-none",
	}
	if goruntime.GOOS == "darwin" {
		for name, arguments := range map[string][]string{
			"DEVELOPER_DIR": {"xcode-select", "-p"},
			"SDKROOT":       {"xcrun", "--show-sdk-path"},
		} {
			output, err := exec.Command(arguments[0], arguments[1:]...).Output()
			if err != nil {
				test.Fatalf("%s: %v", strings.Join(arguments, " "), err)
			}
			values[name] = strings.TrimSpace(string(output))
		}
	}
	for _, name := range unitKeyKeyed {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			set := append(slices.Clone(environment), name+"="+values[name])
			if testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, set).key(unit) == key {
				test.Fatalf("%s was set and the key held", name)
			}
		})
	}
	for _, name := range append(slices.Clone(unitKeyRefused), "CCC_OVERRIDE_OPTIONS") {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			set := append(slices.Clone(environment), name+"=/adamic-none")
			if _, err := newUnitKeying(
				unitFirst,
				common,
				unitKeyFlags,
				compiler,
				set,
				false,
			); !errors.Is(err, errUncachedUnit) {
				test.Fatalf("%s was set and the key was read: %v", name, err)
			}
		})
	}
}

// A unit's key is its own: two units compiled together, one body changed, and the other keys as it did and is taken
// from the cache without clang compiling it. Its cached object is planted with other bytes, so a hit that ran no clang
// returns those bytes; one the sampled check compiled cold returns the cold object and marks the cache off, and that
// mark is lifted and the build tried again, a few times.
// Not parallel: it clears the mark a sampled check may leave in its own cache.
func TestUnitKeyKeepsOtherUnits(test *testing.T) {
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	root := test.TempDir()
	cache := filepath.Join(root, "cache")
	common := []runtimeFile{{"units.h", []byte("#define VALUE 1\n")}}
	alpha := compilationUnit{"alpha.c", unitFirst + "int adamic_alpha(void) { return VALUE; }\n"}
	beta := compilationUnit{"beta.c", unitFirst + "int adamic_beta(void) { return VALUE + 1; }\n"}
	directory := filepath.Join(root, "units")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		test.Fatal(err)
	}
	// As buildUnitsWithLibrary does: one keying a build, and the build's units compiled at once.
	build := func(units ...compilationUnit) ([]string, *unitKeying) {
		test.Helper()
		keying := testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, environment)
		objects := make([]string, len(units))
		failures := make([]error, len(units))
		var compiles sync.WaitGroup
		for index, unit := range units {
			compiles.Go(func() {
				objects[index], failures[index] = compileUnit(
					unit,
					common,
					unitKeyFlags,
					compiler,
					keying,
					cache,
					directory,
					nil,
					false,
				)
			})
		}
		compiles.Wait()
		for _, failure := range failures {
			if failure != nil {
				test.Fatal(failure)
			}
		}
		return objects, keying
	}
	before, keying := build(alpha, beta)
	planted := []byte("beta.c's object, from the cache\n")
	if err := os.WriteFile(before[1], planted, 0o644); err != nil {
		test.Fatal(err)
	}
	changed := compilationUnit{alpha.name, unitFirst + "int adamic_alpha(void) { return VALUE + 2; }\n"}
	after, rekeyed := build(changed)
	if rekeyed.key(beta) != keying.key(beta) {
		test.Fatal("alpha.c's body changed and beta.c's key moved")
	}
	if rekeyed.key(changed) == keying.key(alpha) || after[0] == before[0] {
		test.Fatal("alpha.c's body changed and its object came from the cache")
	}
	for range 5 {
		objects, _ := build(beta)
		if contents, err := os.ReadFile(
			objects[0],
		); err == nil && objects[0] == before[1] && string(contents) == string(planted) {
			return
		}
		if err := os.Remove(filepath.Join(cache, unitCacheOffName)); err != nil && !errors.Is(err, os.ErrNotExist) {
			test.Fatal(err)
		}
	}
	test.Fatal("beta.c was compiled again on every one of five hits")
}

// A mismatch's line is noted in the build log a product test is given, where Loom's builder reads it (Node's M1), one
// line, and nothing is written when no log is named. Mutant: the note left out.
// Not parallel: test.Setenv sets the build log every compile in the process would note to.
func TestUnitMismatchIsNotedInTheBuildLog(test *testing.T) {
	mismatch := unitMismatch{
		Unit:   "probe.c",
		Key:    strings.Repeat("a", 64),
		Cached: strings.Repeat("b", 64),
		Cold:   strings.Repeat("c", 64),
	}
	log := filepath.Join(test.TempDir(), "build.log")
	writeUnitKeyFile(test, log, "build adamic-runtime 0123456789ab hit 0.01\n")
	test.Setenv(unitBuildLog, log)
	noteUnitMismatch(mismatch)
	contents, err := os.ReadFile(log)
	if err != nil {
		test.Fatal(err)
	}
	if want := "build adamic-runtime 0123456789ab hit 0.01\n" +
		unitMismatchLine(mismatch) +
		"\n"; string(contents) != want {
		test.Fatalf("the build log holds %q, want %q", contents, want)
	}
	test.Setenv(unitBuildLog, "")
	noteUnitMismatch(mismatch)
}

// A cold check compiles a hit again and compares: a hit whose bytes match passes, and one whose cached object was
// planted with other bytes builds green with the cold object, and marks the cache off naming the unit, its key and both
// digests. Mutant: the mismatch returned as an error.
// Not parallel: test.Setenv sets ADAMIC_NATIVE_UNIT_CHECK, which every compileUnit in the process reads.
func TestUnitCacheColdCheck(test *testing.T) {
	test.Setenv(unitCheckVariable, "every")
	test.Setenv(unitCacheOffVariable, "")
	compiler := unitKeyCompiler(test)
	common := []runtimeFile{{"units.h", []byte("#define VALUE 5\n")}}
	keying := testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, unitKeyEnvironment())
	unit := unitKeyProbe("VALUE")
	directory := test.TempDir()
	cache := filepath.Join(directory, "cache")
	object := compileUnitKey(test, unit, common, unitKeyFlags, compiler, keying, directory)
	if object != filepath.Join(cache, keying.key(unit), "unit.o") {
		test.Fatalf("the unit compiled to %s, not its entry", object)
	}
	if checked := compileUnitKey(test, unit, common, unitKeyFlags, compiler, keying, directory); checked != object {
		test.Fatalf("a checked hit came from %s, not %s", checked, object)
	}
	if _, off := unitCacheOff(cache); off {
		test.Fatal("a checked hit that matched marked the cache off")
	}
	cold, err := os.ReadFile(object)
	if err != nil {
		test.Fatal(err)
	}
	planted := []byte("not the object\n")
	if err := os.WriteFile(object, planted, 0o644); err != nil {
		test.Fatal(err)
	}
	// The hole is the house's, never the build's: a mismatched hit builds green with the cold object (Node's M2).
	built, err := compileUnit(unit, common, unitKeyFlags, compiler, keying, cache, directory, nil, false)
	if err != nil {
		test.Fatalf("a planted object's checked hit failed the build: %v", err)
	}
	if contents, err := os.ReadFile(built); err != nil || string(contents) != string(cold) {
		test.Fatalf("a planted object's checked hit built %s, not the cold object: %v", built, err)
	}
	want := fmt.Sprintf(
		"native: unit cache mismatch: unit probe.c key %s cached %x cold %x",
		keying.key(unit),
		sha256.Sum256(planted),
		sha256.Sum256(cold),
	)
	if line := unitMismatchLine(unitMismatch{
		Unit:   "probe.c",
		Key:    keying.key(unit),
		Cached: fmt.Sprintf("%x", sha256.Sum256(planted)),
		Cold:   fmt.Sprintf("%x", sha256.Sum256(cold)),
	}); line != want {
		test.Fatalf("the mismatch's line is %q, want %q", line, want)
	}
	contents, err := os.ReadFile(filepath.Join(cache, unitCacheOffName))
	if err != nil {
		test.Fatal(err)
	}
	var mismatch unitMismatch
	if err := json.Unmarshal(contents, &mismatch); err != nil {
		test.Fatal(err)
	}
	if mismatch.Unit != "probe.c" ||
		mismatch.Key != keying.key(unit) ||
		mismatch.Cached != fmt.Sprintf("%x", sha256.Sum256(planted)) ||
		mismatch.Cold != fmt.Sprintf("%x", sha256.Sum256(cold)) {
		test.Fatalf("the mark is %+v", mismatch)
	}
	if why, off := unitCacheOff(cache); !off || !strings.Contains(why, unitCacheOffName) {
		test.Fatalf("with the mark standing the cache is on: %q", why)
	}
}

// While a mismatch's mark stands, or ADAMIC_NATIVE_UNIT_CACHE=off, a build's units compile uncached and the cache holds
// nothing new; with neither, the same build keeps its units there. The last also shows the runtime's real headers are
// keyable here.
// Not parallel: test.Setenv sets where the build cache is written and ADAMIC_NATIVE_UNIT_CACHE.
func TestUnitCacheOffCompilesUncached(test *testing.T) {
	source := "#include \"adamic.h\"\n#include <stdio.h>\nstatic int adamic_answer(void);\nstatic int adamic_answer(void) { return 42; }\nint main(void) { printf(\"%d\\n\", adamic_answer()); return 0; }\n"
	// The runtime comes from the build cache as every build's does, before this test moves where the cache is written.
	library, err := RuntimeLibraryForSource("", source, Options{})
	if err != nil {
		test.Fatal(err)
	}
	test.Setenv(location.OutputDirectoryVariable, test.TempDir())
	test.Setenv("ADAMIC_GATE_UNCACHED", "")
	test.Setenv(unitCheckVariable, "")
	cache, err := cacheDirectory("native-units")
	if err != nil {
		test.Fatal(err)
	}
	entries := func() int {
		test.Helper()
		listed, err := os.ReadDir(cache)
		if errors.Is(err, os.ErrNotExist) {
			return 0
		}
		if err != nil {
			test.Fatal(err)
		}
		count := 0
		for _, entry := range listed {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				count++
			}
		}
		return count
	}
	build := func() {
		test.Helper()
		binary := filepath.Join(test.TempDir(), "program")
		if err := buildUnitsWithLibrary(source, binary, Options{}, library, ""); err != nil {
			test.Fatal(err)
		}
		if output, err := exec.Command(binary).CombinedOutput(); err != nil || string(output) != "42\n" {
			test.Fatalf("the program: %v\n%s", err, output)
		}
	}
	if err := markUnitCacheOff(cache, unitMismatch{Unit: "planted.c"}); err != nil {
		test.Fatal(err)
	}
	build()
	if count := entries(); count != 0 {
		test.Fatalf("with the mark standing the build kept %d units in the cache", count)
	}
	if err := os.Remove(filepath.Join(cache, unitCacheOffName)); err != nil {
		test.Fatal(err)
	}
	test.Setenv(unitCacheOffVariable, "off")
	build()
	if count := entries(); count != 0 {
		test.Fatalf("with %s=off the build kept %d units in the cache", unitCacheOffVariable, count)
	}
	test.Setenv(unitCacheOffVariable, "")
	build()
	if count := entries(); count == 0 {
		test.Fatal("with the cache on the build kept no unit in it")
	}
}

// A mismatch proves the build's cache incomplete, not one unit's: with two units' cached objects planted wrong and the
// cold check falling on one of them, every unit's object the build ends with is the one an uncached build compiles, the
// hit nobody checked included (Node's M4). Mutant: only the checked unit replaced.
func TestUnitCacheMismatchCompilesEveryUnitCold(test *testing.T) {
	test.Parallel()
	compiler := unitKeyCompiler(test)
	environment := unitKeyEnvironment()
	common := []runtimeFile{{"units.h", []byte("#define VALUE 1\n")}}
	units := []compilationUnit{
		{"alpha.c", unitFirst + "int adamic_alpha(void) { return VALUE; }\n"},
		{"beta.c", unitFirst + "int adamic_beta(void) { return VALUE + 1; }\n"},
	}
	root := test.TempDir()
	cache := filepath.Join(root, "cache")
	directory := filepath.Join(root, "units")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		test.Fatal(err)
	}
	never := testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, environment)
	never.check = func(compilationUnit) bool { return false }
	cached, err := compileUnits(units, common, unitKeyFlags, compiler, never, cache, directory, 2, nil, false)
	if err != nil {
		test.Fatal(err)
	}
	for _, object := range cached {
		if err := os.WriteFile(object, []byte("planted\n"), 0o644); err != nil {
			test.Fatal(err)
		}
	}
	keying := testUnitKeying(test, unitFirst, common, unitKeyFlags, compiler, environment)
	keying.check = func(unit compilationUnit) bool { return unit.name == "beta.c" }
	built, err := compileUnits(units, common, unitKeyFlags, compiler, keying, cache, directory, 2, nil, false)
	if err != nil {
		test.Fatalf("a mismatched build failed: %v", err)
	}
	uncachedDirectory := filepath.Join(root, "uncached")
	if err := os.MkdirAll(uncachedDirectory, 0o755); err != nil {
		test.Fatal(err)
	}
	uncached, err := compileUnits(units, common, unitKeyFlags, compiler, nil, cache, uncachedDirectory, 2, nil, false)
	if err != nil {
		test.Fatal(err)
	}
	for index, unit := range units {
		got, err := os.ReadFile(built[index])
		if err != nil {
			test.Fatal(err)
		}
		want, err := os.ReadFile(uncached[index])
		if err != nil {
			test.Fatal(err)
		}
		if string(got) != string(want) {
			test.Fatalf(
				"after beta.c's check mismatched, %s's object (%s) isn't the uncached compile's",
				unit.name,
				built[index],
			)
		}
	}
	if _, off := unitCacheOff(cache); !off {
		test.Fatal("the mismatch left no mark")
	}
}
