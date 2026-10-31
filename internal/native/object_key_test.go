package native

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// objectKeySource is a program whose main.c includes a header only once it's there, so a header planted between the key
// and the recheck changes what main.c preprocesses to without changing any file the key's preprocessing read.
const objectKeySource = `#include "adamic.h"
#if __has_include("object_key_late.h")
#include "object_key_late.h"
#endif
int main(int argc, char **argv) {
	adamic_start(argc, argv);
	return 0;
}
`

// objectKeyEdit is one byte of adamic.h changed, in a declaration main.c's preprocessing keeps.
var objectKeyEdit = [2]string{"void adamic_start(int count,", "void adamic_start(int counu,"}

// objectKeyed is what keying one build's main.c read: its keying, or the refusal it got, and what its recheck said.
type objectKeyed struct {
	keying  *objectKeying
	refusal error
	differs string
}

// objectKeyBuild builds source as options say with linkOne, as buildOneWithRuntime does, in directory, made here, with
// compiler and flags, linking library, and main.c keyed as objectKey keys it, read here rather than said. prepare, when
// it isn't nil, is handed the compile before it's keyed, and between, when it isn't nil, runs once it's keyed and before
// it compiles.
func objectKeyBuild(t *testing.T, directory string, compiler string, source string, options Options, flags []string, library string, prepare func(compile *exec.Cmd), between func(keying *objectKeying)) objectKeyed {
	t.Helper()
	jobserver, err := defaultJobserver()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	var keyed objectKeyed
	keyObject := func(compile *exec.Cmd, source string, compileFlags []string, program string, jobserver *Jobserver) (string, func(), error) {
		if prepare != nil {
			prepare(compile)
		}
		keying, err := newObjectKeying(compile, source, compileFlags, program, jobserver)
		if errors.Is(err, errUncachedObject) {
			keyed.refusal = err
			return "", nil, nil
		}
		if err != nil {
			return "", nil, err
		}
		keyed.keying = keying
		if between != nil {
			between(keying)
		}
		return keying.key, func() { keyed.differs = keying.recheck(jobserver) }, nil
	}
	output := directory + ".program"
	if err := linkOne(directory, compiler, source, output, options, flags, library, jobserver, func(directory string) ([]string, error) {
		return regexpCompilerLinkFlags(source, options, directory)
	}, keyObject, false); err != nil {
		t.Fatal(err)
	}
	return keyed
}

// objectKeyRuntime copies the runtime archive's directory holding library into directory, with edit applied to
// adamic.h when it isn't nil, and is the copy's archive.
func objectKeyRuntime(t *testing.T, library string, directory string, edit func(header []byte) []byte) string {
	t.Helper()
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Dir(library))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		contents, err := os.ReadFile(filepath.Join(filepath.Dir(library), entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == "adamic.h" && edit != nil {
			contents = edit(contents)
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(directory, filepath.Base(library))
}

// objectKeyEdited is adamic.h with objectKeyEdit's byte changed.
func objectKeyEdited(t *testing.T, header []byte) []byte {
	t.Helper()
	if !bytes.Contains(header, []byte(objectKeyEdit[0])) {
		t.Fatalf("adamic.h no longer declares %q, the byte the test changes", objectKeyEdit[0])
	}
	return bytes.Replace(header, []byte(objectKeyEdit[0]), []byte(objectKeyEdit[1]), 1)
}

// objectKeyLine is a build log's line for main.c's compile, or its key's preprocessing or recheck.
var objectKeyLine = regexp.MustCompile(`^(cCompile|cPreprocess|cRecheck) main\.c (\S+) (\S+) `)

// objectKeyLogged is the key each of main.c's lines in the build log at path carries, by its kind: cCompile's, and
// cPreprocess's and cRecheck's where it was keyed.
func objectKeyLogged(t *testing.T, path string) map[string][]string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	logged := map[string][]string{}
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	for scanner.Scan() {
		if match := objectKeyLine.FindStringSubmatch(scanner.Text()); match != nil {
			if match[3] != "off" {
				t.Errorf("main.c's %s line says %s; nothing is cached by its key", match[1], match[3])
			}
			logged[match[1]] = append(logged[match[1]], match[2])
		}
	}
	return logged
}

// The same program built in two build directories under two temporary directories gets the same object key, released,
// Debug (where main.c and the runtime's headers are compiled by their absolute paths, so its line markers name the
// build directory) and sanitized, through Build with ADAMIC_NATIVE_OBJECT_KEY=on: every per-build place is a placeholder
// in the arguments and mapped in the preprocessed bytes. The three builds key apart. Mutant, red: the build directory
// left in the arguments as it is.
func TestObjectKeyIsTheSameFromAnyDirectory(t *testing.T) {
	source := splitCorpusSource(t, "arguments.a")
	root := t.TempDir()
	t.Setenv(objectKeyVariable, "on")
	seen := map[string]string{}
	for _, build := range []struct {
		name    string
		options Options
	}{{"release", Options{}}, {"debug", Options{Debug: true}}, {"sanitized", Options{Sanitize: true}}} {
		var keys [2]string
		// The second temporary directory is deeper than the first, so a place left in the key moves it by its length too.
		for index, temporary := range []string{filepath.Join(root, build.name, "first"), filepath.Join(root, build.name, "second", "deeper")} {
			if err := os.MkdirAll(temporary, 0o755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("TMPDIR", temporary)
			log := filepath.Join(temporary, "build.log")
			t.Setenv("ADAMIC_BUILD_LOG", log)
			if err := Build(source, filepath.Join(temporary, "program"), build.options); err != nil {
				t.Fatalf("%s: %v", build.name, err)
			}
			logged := objectKeyLogged(t, log)
			if len(logged["cCompile"]) != 1 || len(logged["cPreprocess"]) != 1 || len(logged["cRecheck"]) != 1 {
				t.Fatalf("%s: main.c's lines in the build log, by kind: %v", build.name, logged)
			}
			keys[index] = logged["cCompile"][0]
			if keys[index] == "-" || logged["cPreprocess"][0] != keys[index] || logged["cRecheck"][0] != keys[index] {
				t.Fatalf("%s: main.c's compile wasn't keyed, or its lines disagree: %v", build.name, logged)
			}
		}
		if keys[0] != keys[1] {
			t.Errorf("%s: keyed %s in one directory and %s in another", build.name, keys[0], keys[1])
		}
		if other, keyed := seen[keys[0]]; keyed {
			t.Errorf("%s and %s key the same: %s", other, build.name, keys[0])
		}
		seen[keys[0]] = build.name
	}
}

// The key moves with each input the compile reads, one at a time, over one program: a flag (-O1 for -O2, a -D added),
// one byte of a runtime header, clang's build (a config file beside the program it runs by), and a variable of the
// environment clang reads (SOURCE_DATE_EPOCH). A build directory of its own and the runtime archive's directory copied
// elsewhere, the same bytes, key the same.
func TestObjectKeyMovesWithEachInput(t *testing.T) {
	t.Parallel()
	source := splitCorpusSource(t, "arguments.a")
	library, err := RuntimeLibraryForSource("", source, Options{})
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	copied := objectKeyRuntime(t, library, filepath.Join(root, "runtime-copied"), nil)
	edited := objectKeyRuntime(t, library, filepath.Join(root, "runtime-edited"), func(header []byte) []byte { return objectKeyEdited(t, header) })
	// clang by a link beside a config file of its own, which says nothing: what the compile runs is the same program, but
	// clang reads another set of files to decide its compile.
	configured := filepath.Join(root, "configured")
	if err := os.Mkdir(configured, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(compiler, filepath.Join(configured, "clang")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configured, "clang.cfg"), []byte("# says nothing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	flags := Flags(Options{})
	lowered := slices.Clone(flags)
	lowered[slices.Index(lowered, "-O2")] = "-O1"
	base := objectKeyBuild(t, filepath.Join(root, "build-base"), compiler, source, Options{}, flags, library, nil, nil)
	if base.keying == nil {
		t.Fatalf("the program got no key: %v", base.refusal)
	}
	for index, row := range []struct {
		name     string
		compiler string
		flags    []string
		library  string
		prepare  func(compile *exec.Cmd)
		moves    bool
	}{
		{"a build directory of its own", compiler, flags, library, nil, false},
		{"the runtime archive's directory copied", compiler, flags, copied, nil, false},
		{"-O1 for -O2", compiler, lowered, library, nil, true},
		{"a -D added", compiler, append(slices.Clone(flags), "-DADAMIC_OBJECT_KEY_MUTANT=1"), library, nil, true},
		{"one byte of adamic.h", compiler, flags, edited, nil, true},
		{"clang configured", filepath.Join(configured, "clang"), flags, library, nil, true},
		{"SOURCE_DATE_EPOCH set", compiler, flags, library, func(compile *exec.Cmd) { compile.Env = append(compile.Env, "SOURCE_DATE_EPOCH=1") }, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			keyed := objectKeyBuild(t, filepath.Join(root, fmt.Sprintf("build-%d", index)), row.compiler, source, Options{}, row.flags, row.library, row.prepare, nil)
			if keyed.keying == nil {
				t.Fatalf("no key: %v", keyed.refusal)
			}
			if moved := keyed.keying.key != base.keying.key; moved != row.moves {
				t.Errorf("the key moved: %t, want %t (%s, the base's %s)", moved, row.moves, keyed.keying.key, base.keying.key)
			}
			if keyed.differs != "" {
				t.Errorf("an untouched build's recheck differs: %s", keyed.differs)
			}
		})
	}
}

// A source that reads the time it's compiled, by any of the three macros, gets no key, and the refusal names the macro;
// the build goes on and passes, and nothing is preprocessed for it. Mutant, red: the refusal dropped.
func TestObjectKeyRefusesTheTimeOfTheCompile(t *testing.T) {
	t.Parallel()
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, macro := range []string{"__DATE__", "__TIME__", "__TIMESTAMP__"} {
		t.Run(macro, func(t *testing.T) {
			t.Parallel()
			source := strings.Replace(objectKeySource, "return 0;", "return "+macro+"[0] == 0;", 1)
			library, err := RuntimeLibraryForSource("", source, Options{})
			if err != nil {
				t.Fatal(err)
			}
			directory := filepath.Join(root, strings.Trim(macro, "_"))
			keyed := objectKeyBuild(t, directory, compiler, source, Options{}, Flags(Options{}), library, nil, nil)
			if keyed.keying != nil || !errors.Is(keyed.refusal, errUncachedObject) {
				t.Fatalf("keyed: %v, refused: %v", keyed.keying != nil, keyed.refusal)
			}
			if want := "native: main.c compiles uncached: main.c reads " + macro; keyed.refusal.Error() != want {
				t.Errorf("refused as %q, want %q", keyed.refusal, want)
			}
			if _, err := os.Stat(filepath.Join(directory, "main.i")); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("a refused source was preprocessed: %v", err)
			}
		})
	}
}

// The recheck after the compile passes an untouched build, and fails one whose inputs moved under it: adamic.h rewritten
// with the same bytes, which only its change time tells; adamic.h with one byte changed; and a header planted where
// main.c asks for one, which changes no file the key's preprocessing read and only the preprocessed bytes tell. Mutants,
// each red: the second preprocessing skipped, and the change times not looked at again.
func TestObjectKeyRecheckSeesWhatMovedUnderTheCompile(t *testing.T) {
	t.Parallel()
	library, err := RuntimeLibraryForSource("", objectKeySource, Options{})
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for index, row := range []struct {
		name string
		// move changes an input of the compile in its build directory, whose runtime link names a copy of the archive's
		// directory.
		move func(t *testing.T, directory string)
		// differs are what the recheck says, each a part of one of its reasons; none for an untouched build.
		differs []string
	}{
		{"untouched", nil, nil},
		{"adamic.h rewritten with the same bytes", func(t *testing.T, directory string) {
			header := filepath.Join(directory, "runtime", "adamic.h")
			contents, err := os.ReadFile(header)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(header, contents, 0o644); err != nil {
				t.Fatal(err)
			}
		}, []string{"runtime/adamic.h changed between main.c's key and its recheck"}},
		{"adamic.h with one byte changed", func(t *testing.T, directory string) {
			header := filepath.Join(directory, "runtime", "adamic.h")
			contents, err := os.ReadFile(header)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(header, objectKeyEdited(t, contents), 0o644); err != nil {
				t.Fatal(err)
			}
		}, []string{"main.c preprocessed to other bytes", "runtime/adamic.h changed between main.c's key and its recheck"}},
		{"a header planted", func(t *testing.T, directory string) {
			if err := os.WriteFile(filepath.Join(directory, "object_key_late.h"), []byte("#define ADAMIC_OBJECT_KEY_LATE 1\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, []string{"main.c preprocessed to other bytes"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			copied := objectKeyRuntime(t, library, filepath.Join(root, fmt.Sprintf("runtime-%d", index)), nil)
			directory := filepath.Join(root, fmt.Sprintf("build-%d", index))
			keyed := objectKeyBuild(t, directory, compiler, objectKeySource, Options{}, Flags(Options{}), copied, nil, func(keying *objectKeying) {
				if row.move != nil {
					row.move(t, directory)
				}
			})
			if keyed.keying == nil {
				t.Fatalf("no key: %v", keyed.refusal)
			}
			if len(row.differs) == 0 && keyed.differs != "" {
				t.Errorf("the recheck differs: %s", keyed.differs)
			}
			for _, differs := range row.differs {
				if !strings.Contains(keyed.differs, differs) {
					t.Errorf("the recheck said %q, without %q", keyed.differs, differs)
				}
			}
			if reasons := strings.Split(keyed.differs, "; "); len(row.differs) > 0 && len(reasons) != len(row.differs) {
				t.Errorf("the recheck said %q, want only %q", keyed.differs, row.differs)
			}
		})
	}
}

// Only a line marker's path is rewritten, and only its start, by the longest per-build place it begins with, as the
// compile's prefix maps rewrite a path in the object; its line and flags stay exactly, and every other line, a place
// another map names, and a marker naming no per-build place are as they were. Mutant, red: a marker's flags dropped.
func TestObjectKeyMapsOnlyTheLineMarkersPaths(t *testing.T) {
	t.Parallel()
	places := objectPlaces([]string{"-O2", "-ffile-prefix-map=/tmp/adamic-build-1=/adamic-build", "-ffile-prefix-map=/tmp/adamic-build-1/runtime-held=/adamic-runtime",
		"-ffile-prefix-map=/tools/llvm=/adamic-compiler", "-ffile-prefix-map=/elsewhere=/elsewhere", "-I", "/tmp/adamic-build-1/runtime"})
	preprocessed := strings.Join([]string{
		`# 1 "/tmp/adamic-build-1/main.c"`,
		`# 1 "<built-in>" 1`,
		`# 410 "<built-in>" 3`,
		`# 1 "<command line>" 1`,
		`# 1 "/tmp/adamic-build-1/runtime/adamic.h" 1`,
		`# 1 "/tools/llvm/lib/clang/20/include/stddef.h" 1 3 4`,
		`# 28 "/usr/include/stdio.h" 2 3 4`,
		`# 7 "/tmp/adamic-build-1/runtime-held/adamic.h" 2`,
		`# 3 "runtime/adamic.h" 2`,
		`# 12 "main.c"`,
		`# 4 "/elsewhere/x.h" 1`,
		`const char *place = "/tmp/adamic-build-1/main.c";`,
		`#pragma clang diagnostic push`,
		`#   1 "/tmp/adamic-build-1/main.c"`,
		``,
	}, "\n")
	want := strings.Join([]string{
		`# 1 "/adamic-build/main.c"`,
		`# 1 "<built-in>" 1`,
		`# 410 "<built-in>" 3`,
		`# 1 "<command line>" 1`,
		`# 1 "/adamic-build/runtime/adamic.h" 1`,
		`# 1 "/adamic-compiler/lib/clang/20/include/stddef.h" 1 3 4`,
		`# 28 "/usr/include/stdio.h" 2 3 4`,
		`# 7 "/adamic-runtime/adamic.h" 2`,
		`# 3 "runtime/adamic.h" 2`,
		`# 12 "main.c"`,
		`# 4 "/elsewhere/x.h" 1`,
		`const char *place = "/tmp/adamic-build-1/main.c";`,
		`#pragma clang diagnostic push`,
		`#   1 "/tmp/adamic-build-1/main.c"`,
		``,
	}, "\n")
	if mapped := string(objectMarkersMapped([]byte(preprocessed), places)); mapped != want {
		t.Errorf("mapped:\n%s\nwant:\n%s", mapped, want)
	}
	// The arguments write each place as its placeholder, the longest first, and leave a place another map names.
	placeheld := objectPlaceheld(places)
	for argument, want := range map[string]string{
		"-ffile-prefix-map=/tmp/adamic-build-1=/adamic-build":                "-ffile-prefix-map=<build>=/adamic-build",
		"-ffile-prefix-map=/tmp/adamic-build-1/runtime-held=/adamic-runtime": "-ffile-prefix-map=<runtime>=/adamic-runtime",
		"/tmp/adamic-build-1/runtime":                                        "<build>/runtime",
		"-ffile-prefix-map=/tools/llvm=/adamic-compiler":                     "-ffile-prefix-map=<compiler>=/adamic-compiler",
		"-ffile-prefix-map=/elsewhere=/elsewhere":                            "-ffile-prefix-map=/elsewhere=/elsewhere",
	} {
		if held := placeheld.Replace(argument); held != want {
			t.Errorf("%s is keyed as %s, want %s", argument, held, want)
		}
	}
}

// With ADAMIC_NATIVE_OBJECT_KEY unset, a build keys nothing: linkOne preprocesses nothing, writing no main.i or main.d,
// and the build log's line for main.c's compile names no key, with no line for a preprocessing. Set to on, the same build
// preprocesses twice and its compile's line carries the key, so the off build's silence isn't the log's.
func TestObjectKeyOffRunsNoPreprocessing(t *testing.T) {
	library, err := RuntimeLibraryForSource("", objectKeySource, Options{})
	if err != nil {
		t.Fatal(err)
	}
	jobserver, err := defaultJobserver()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, setting := range []string{"", "on"} {
		t.Setenv(objectKeyVariable, setting)
		log := filepath.Join(root, "build-"+setting+".log")
		t.Setenv("ADAMIC_BUILD_LOG", log)
		directory := filepath.Join(root, "build-"+setting)
		if err := os.Mkdir(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := linkOne(directory, "clang", objectKeySource, directory+".program", Options{}, Flags(Options{}), library, jobserver, nil, objectKeyer(), false); err != nil {
			t.Fatal(err)
		}
		var written []string
		for _, name := range []string{"main.i", "main.d", "main.recheck.i"} {
			if _, err := os.Stat(filepath.Join(directory, name)); err == nil {
				written = append(written, name)
			}
		}
		logged := objectKeyLogged(t, log)
		if setting == "" && (len(written) > 0 || !slices.Equal(logged["cCompile"], []string{"-"}) || len(logged) != 1) {
			t.Errorf("off, the build wrote %v and logged %v", written, logged)
		}
		if setting == "on" && (len(written) != 3 || len(logged["cCompile"]) != 1 || logged["cCompile"][0] == "-" || len(logged["cPreprocess"]) != 1 || len(logged["cRecheck"]) != 1) {
			t.Errorf("on, the build wrote %v and logged %v", written, logged)
		}
	}
}

// Two sources a space apart inside one line preprocess to the same bytes, and under -g compile to other objects, whose
// line tables record columns: they key apart, released and Debug, since every file the preprocessing read is keyed by
// its bytes. Mutant, red: the files' hashes left out of the key.
func TestObjectKeyHoldsEveryColumn(t *testing.T) {
	t.Parallel()
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	spaced := strings.Replace(objectKeySource, "return 0;", "return  0;", 1)
	for _, build := range []struct {
		name    string
		options Options
	}{{"release", Options{}}, {"debug", Options{Debug: true}}} {
		t.Run(build.name, func(t *testing.T) {
			t.Parallel()
			library, err := RuntimeLibraryForSource("", objectKeySource, build.options)
			if err != nil {
				t.Fatal(err)
			}
			var keyed [2]objectKeyed
			for index, source := range []string{objectKeySource, spaced} {
				keyed[index] = objectKeyBuild(t, filepath.Join(root, fmt.Sprintf("%s-%d", build.name, index)), compiler, source, build.options, Flags(build.options), library, nil, nil)
				if keyed[index].keying == nil {
					t.Fatalf("no key: %v", keyed[index].refusal)
				}
			}
			// A Debug build's markers name its own build directory, so the bytes are compared as the key maps them.
			var mapped [2][]byte
			for index := range keyed {
				mapped[index] = objectMarkersMapped(keyed[index].keying.preprocessed, objectPlaces(keyed[index].keying.compile.Args[1:]))
			}
			if !bytes.Equal(mapped[0], mapped[1]) {
				t.Fatal("the two sources preprocess to other bytes, so the test shows nothing of the files' hashes")
			}
			if keyed[0].keying.key == keyed[1].keying.key {
				t.Errorf("two sources a space apart key the same: %s", keyed[0].keying.key)
			}
		})
	}
}
