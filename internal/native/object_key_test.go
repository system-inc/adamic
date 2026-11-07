package native

import (
	"bufio"
	"bytes"
	"crypto/sha256"
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
// compiler and flags, linking library, and main.c keyed and compiled uncached, its recheck read here. prepare, when
// it isn't nil, is handed the compile before it's keyed, and between, when it isn't nil, runs once it's keyed and
// before it compiles.
func objectKeyBuild(
	test *testing.T,
	directory string,
	compiler string,
	source string,
	options Options,
	flags []string,
	library string,
	prepare func(compile *exec.Cmd),
	between func(keying *objectKeying),
) objectKeyed {
	test.Helper()
	jobserver, defaultJobserverError := defaultJobserver()
	if defaultJobserverError != nil {
		test.Fatal(defaultJobserverError)
	}
	if createDirectoryError := os.Mkdir(directory, 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	var keyed objectKeyed
	keyObject := func(
		compile *exec.Cmd,
		source string,
		compileFlags []string,
		program string,
		jobserver *Jobserver,
	) error {
		if prepare != nil {
			prepare(compile)
		}
		keying, newObjectKeyingError := newObjectKeying(compile, source, compileFlags, program, jobserver)
		if errors.Is(newObjectKeyingError, errorUncachedObject) {
			keyed.refusal = newObjectKeyingError
			_, compileMainError := compileMain(compile, jobserver)
			return compileMainError
		}
		if newObjectKeyingError != nil {
			return newObjectKeyingError
		}
		keyed.keying = keying
		if between != nil {
			between(keying)
		}
		if _, compileMainError := compileMain(compile, jobserver); compileMainError != nil {
			return compileMainError
		}
		keyed.differs = keying.recheck(jobserver)
		return nil
	}
	output := directory + ".program"
	if linkOneError := linkOne(
		directory,
		compiler,
		source,
		output,
		options,
		flags,
		library,
		jobserver,
		func(directory string) ([]string, error) {
			return regexpCompilerLinkFlags(source, options, directory)
		},
		keyObject,
		false,
	); linkOneError != nil {
		test.Fatal(linkOneError)
	}
	return keyed
}

// objectKeyRuntime copies the runtime archive's directory holding library into directory, with edit applied to
// adamic.h when it isn't nil, and is the copy's archive.
func objectKeyRuntime(test *testing.T, library string, directory string, edit func(header []byte) []byte) string {
	test.Helper()
	if createDirectoryError := os.Mkdir(directory, 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	entries, readDirectoryError := os.ReadDir(filepath.Dir(library))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	for _, entry := range entries {
		contents, readError := os.ReadFile(filepath.Join(filepath.Dir(library), entry.Name()))
		if readError != nil {
			test.Fatal(readError)
		}
		if entry.Name() == "adamic.h" && edit != nil {
			contents = edit(contents)
		}
		if writeError := os.WriteFile(filepath.Join(directory, entry.Name()), contents, 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return filepath.Join(directory, filepath.Base(library))
}

// objectKeyEdited is adamic.h with objectKeyEdit's byte changed.
func objectKeyEdited(test *testing.T, header []byte) []byte {
	test.Helper()
	if !bytes.Contains(header, []byte(objectKeyEdit[0])) {
		test.Fatalf("adamic.h no longer declares %q, the byte the test changes", objectKeyEdit[0])
	}
	return bytes.Replace(header, []byte(objectKeyEdit[0]), []byte(objectKeyEdit[1]), 1)
}

// objectKeyLine is a build log's line for main.c's compile, or its key's preprocessing or recheck.
var objectKeyLine = regexp.MustCompile(`^(cCompile|cPreprocess|cRecheck) main\.c (\S+) (\S+) `)

// objectKeyLogged is the key each of main.c's lines in the build log at path carries, by its kind: cCompile's, and
// cPreprocess's and cRecheck's where it was keyed; cCompile's outcome is under "outcome".
func objectKeyLogged(test *testing.T, path string) map[string][]string {
	test.Helper()
	contents, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	logged := map[string][]string{}
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	for scanner.Scan() {
		if match := objectKeyLine.FindStringSubmatch(scanner.Text()); match != nil {
			logged[match[1]] = append(logged[match[1]], match[2])
			if match[1] == "cCompile" {
				logged["outcome"] = append(logged["outcome"], match[3])
			}
		}
	}
	return logged
}

// The same program built in two build directories under two temporary directories gets the same object key, through
// Build trusted as a candidate, in a Debug build: there main.c and the runtime's headers are compiled by their absolute
// paths, so the arguments and the line markers both name the build directory, and every per-build place must be a
// placeholder in the arguments and mapped in the preprocessed bytes. A release or sanitized build names no directory a
// Debug one doesn't, so one kind is the test, two whole builds where it was six. Mutant, red: the build directory left
// in the arguments as it is.
// Not parallel: it sets the trust, ADAMIC_NATIVE_OBJECT_TRUST, with test.Setenv for the builds it runs.
func TestObjectKeyIsTheSameFromAnyDirectory(test *testing.T) {
	source := splitCorpusSource(test, "arguments.a")
	root := test.TempDir()
	test.Setenv(objectTrustVariable, "candidate")
	for _, build := range []struct {
		name    string
		options Options
	}{{"debug", Options{Debug: true}}} {
		var keys [2]string
		// The second temporary directory is deeper than the first, so a place left in the key moves it by its length
		// too.
		for index, temporary := range []string{
			filepath.Join(root, build.name, "first"),
			filepath.Join(root, build.name, "second", "deeper"),
		} {
			if createDirectoryError := os.MkdirAll(temporary, 0o755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			test.Setenv("TMPDIR", temporary)
			log := filepath.Join(temporary, "build.log")
			test.Setenv("ADAMIC_BUILD_LOG", log)
			if buildError := Build(source, filepath.Join(temporary, "program"), build.options); buildError != nil {
				test.Fatalf("%s: %v", build.name, buildError)
			}
			logged := objectKeyLogged(test, log)
			// A hit, the second build's when the first stored its object, runs no recheck.
			if len(logged["cCompile"]) != 1 || len(logged["cPreprocess"]) != 1 || len(logged["cRecheck"]) > 1 {
				test.Fatalf("%s: main.c's lines in the build log, by kind: %v", build.name, logged)
			}
			keys[index] = logged["cCompile"][0]
			if keys[index] == "-" ||
				logged["cPreprocess"][0] != keys[index] ||
				(len(logged["cRecheck"]) == 1 && logged["cRecheck"][0] != keys[index]) {
				test.Fatalf("%s: main.c's compile wasn't keyed, or its lines disagree: %v", build.name, logged)
			}
		}
		if keys[0] != keys[1] {
			test.Errorf("%s: keyed %s in one directory and %s in another", build.name, keys[0], keys[1])
		}
	}
}

// The key moves with each input the compile reads, one at a time, over one program: a flag (-O1 for -O2, a -D added),
// one byte of a runtime header, clang's build (a config file beside the program it runs by), and a variable of the
// environment clang reads (SOURCE_DATE_EPOCH). A build directory of its own and the runtime archive's directory copied
// elsewhere, the same bytes, key the same.
func TestObjectKeyMovesWithEachInput(test *testing.T) {
	test.Parallel()
	source := splitCorpusSource(test, "arguments.a")
	library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", source, Options{})
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	root := test.TempDir()
	// Every row builds from a copy of the runtime's directory under root, never the runtime cache's own: prefixMaps
	// gives a directory reached through a link a second map, for its real path, so the cache's directory, linked on
	// some boxes (the Pupa check's), keys apart from a copy for its arguments alone, which is right and isn't what a
	// row asks.
	runtime := objectKeyRuntime(test, library, filepath.Join(root, "runtime"), nil)
	copied := objectKeyRuntime(test, library, filepath.Join(root, "runtime-copied"), nil)
	edited := objectKeyRuntime(
		test,
		library,
		filepath.Join(root, "runtime-edited"),
		func(header []byte) []byte { return objectKeyEdited(test, header) },
	)
	// clang by a link beside a config file of its own, which says nothing: what the compile runs is the same program,
	// but clang reads another set of files to decide its compile.
	configured := filepath.Join(root, "configured")
	if createDirectoryError := os.Mkdir(configured, 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if symlinkError := os.Symlink(compiler, filepath.Join(configured, "clang")); symlinkError != nil {
		test.Fatal(symlinkError)
	}
	if writeError := os.WriteFile(
		filepath.Join(configured, "clang.cfg"),
		[]byte("# says nothing\n"),
		0o644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	flags := Flags(Options{})
	lowered := slices.Clone(flags)
	lowered[slices.Index(lowered, "-O2")] = "-O1"
	base := objectKeyBuild(
		test,
		filepath.Join(root, "build-base"),
		compiler,
		source,
		Options{},
		flags,
		runtime,
		nil,
		nil,
	)
	if base.keying == nil {
		test.Fatalf("the program got no key: %v", base.refusal)
	}
	for index, row := range []struct {
		name     string
		compiler string
		flags    []string
		library  string
		prepare  func(compile *exec.Cmd)
		moves    bool
	}{
		{"a build directory of its own", compiler, flags, runtime, nil, false},
		{"the runtime archive's directory copied", compiler, flags, copied, nil, false},
		{"-O1 for -O2", compiler, lowered, runtime, nil, true},
		{"a -D added", compiler, append(slices.Clone(flags), "-DADAMIC_OBJECT_KEY_MUTANT=1"), runtime, nil, true},
		{"one byte of adamic.h", compiler, flags, edited, nil, true},
		{"clang configured", filepath.Join(configured, "clang"), flags, runtime, nil, true},
		{
			"SOURCE_DATE_EPOCH set",
			compiler,
			flags,
			runtime,
			func(compile *exec.Cmd) { compile.Env = append(compile.Env, "SOURCE_DATE_EPOCH=1") },
			true,
		},
	} {
		// The rows run one at a time: each is a whole build, and seven at once beside the package's other parallel
		// tests passed a runner's memory share (internal/native#14/33 on Adam, killed past 4.1 GB where it peaked at
		// 2.5 GB).
		test.Run(row.name, func(test *testing.T) {
			keyed := objectKeyBuild(
				test,
				filepath.Join(root, fmt.Sprintf("build-%d", index)),
				row.compiler,
				source,
				Options{},
				row.flags,
				row.library,
				row.prepare,
				nil,
			)
			if keyed.keying == nil {
				test.Fatalf("no key: %v", keyed.refusal)
			}
			if moved := keyed.keying.key != base.keying.key; moved != row.moves {
				test.Errorf(
					"the key moved: %t, want %t (%s, the base's %s)\n%s",
					moved,
					row.moves,
					keyed.keying.key,
					base.keying.key,
					objectPartsMoved(base.keying.parts, keyed.keying.parts),
				)
			}
			if keyed.differs != "" {
				test.Errorf("an untouched build's recheck differs: %s", keyed.differs)
			}
		})
	}
}

// A source that reads the time it's compiled, by any of the three macros, gets no key, and the refusal names the macro;
// the build goes on and passes, and nothing is preprocessed for it. Mutant, red: the refusal dropped.
func TestObjectKeyRefusesTheTimeOfTheCompile(test *testing.T) {
	test.Parallel()
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	root := test.TempDir()
	for _, macro := range []string{"__DATE__", "__TIME__", "__TIMESTAMP__"} {
		test.Run(macro, func(test *testing.T) {
			test.Parallel()
			source := strings.Replace(objectKeySource, "return 0;", "return "+macro+"[0] == 0;", 1)
			library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", source, Options{})
			if runtimeLibraryForSourceError != nil {
				test.Fatal(runtimeLibraryForSourceError)
			}
			directory := filepath.Join(root, strings.Trim(macro, "_"))
			keyed := objectKeyBuild(test, directory, compiler, source, Options{}, Flags(Options{}), library, nil, nil)
			if keyed.keying != nil || !errors.Is(keyed.refusal, errorUncachedObject) {
				test.Fatalf("keyed: %v, refused: %v", keyed.keying != nil, keyed.refusal)
			}
			if want := "native: main.c compiles uncached: main.c reads " + macro; keyed.refusal.Error() != want {
				test.Errorf("refused as %q, want %q", keyed.refusal, want)
			}
			if _, statError := os.Stat(filepath.Join(directory, "main.i")); !errors.Is(statError, os.ErrNotExist) {
				test.Errorf("a refused source was preprocessed: %v", statError)
			}
		})
	}
}

// The recheck after the compile passes an untouched build, and fails one whose inputs moved under it: adamic.h
// rewritten with the same bytes, which only its change time tells; adamic.h with one byte changed; and a header planted
// where main.c asks for one, which changes no file the key's preprocessing read and only the preprocessed bytes tell.
// Mutants, each red: the second preprocessing skipped, and the change times not looked at again.
func TestObjectKeyRecheckSeesWhatMovedUnderTheCompile(test *testing.T) {
	test.Parallel()
	library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", objectKeySource, Options{})
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	root := test.TempDir()
	for index, row := range []struct {
		name string
		// move changes an input of the compile in its build directory, whose runtime link names a copy of the archive's
		// directory.
		move func(test *testing.T, directory string)
		// differs are what the recheck says, each a part of one of its reasons; none for an untouched build.
		differs []string
	}{
		{"untouched", nil, nil},
		{"adamic.h rewritten with the same bytes", func(test *testing.T, directory string) {
			header := filepath.Join(directory, "runtime", "adamic.h")
			contents, readError := os.ReadFile(header)
			if readError != nil {
				test.Fatal(readError)
			}
			if writeError := os.WriteFile(header, contents, 0o644); writeError != nil {
				test.Fatal(writeError)
			}
		}, []string{"runtime/adamic.h changed between main.c's key and its recheck"}},
		{
			"adamic.h with one byte changed",
			func(test *testing.T, directory string) {
				header := filepath.Join(directory, "runtime", "adamic.h")
				contents, readError := os.ReadFile(header)
				if readError != nil {
					test.Fatal(readError)
				}
				if writeError := os.WriteFile(header, objectKeyEdited(test, contents), 0o644); writeError != nil {
					test.Fatal(writeError)
				}
			},
			[]string{
				"main.c preprocessed to other bytes",
				"runtime/adamic.h changed between main.c's key and its recheck",
			},
		},
		{"a header planted", func(test *testing.T, directory string) {
			if writeError := os.WriteFile(
				filepath.Join(directory, "object_key_late.h"),
				[]byte("#define ADAMIC_OBJECT_KEY_LATE 1\n"),
				0o644,
			); writeError != nil {
				test.Fatal(writeError)
			}
		}, []string{"main.c preprocessed to other bytes"}},
	} {
		test.Run(row.name, func(test *testing.T) {
			test.Parallel()
			copied := objectKeyRuntime(test, library, filepath.Join(root, fmt.Sprintf("runtime-%d", index)), nil)
			directory := filepath.Join(root, fmt.Sprintf("build-%d", index))
			keyed := objectKeyBuild(test, directory, compiler, objectKeySource, Options{}, Flags(
				Options{},
			), copied, nil, func(keying *objectKeying) {
				if row.move != nil {
					row.move(test, directory)
				}
			})
			if keyed.keying == nil {
				test.Fatalf("no key: %v", keyed.refusal)
			}
			if len(row.differs) == 0 && keyed.differs != "" {
				test.Errorf("the recheck differs: %s", keyed.differs)
			}
			for _, differs := range row.differs {
				if !strings.Contains(keyed.differs, differs) {
					test.Errorf("the recheck said %q, without %q", keyed.differs, differs)
				}
			}
			if reasons := strings.Split(keyed.differs, "; "); len(row.differs) > 0 && len(reasons) != len(row.differs) {
				test.Errorf("the recheck said %q, want only %q", keyed.differs, row.differs)
			}
		})
	}
}

// Only a line marker's path is rewritten, and only its start, by the longest per-build place it begins with, as the
// compile's prefix maps rewrite a path in the object; its line and flags stay exactly, and every other line, a place
// another map names, and a marker naming no per-build place are as they were. Mutant, red: a marker's flags dropped.
func TestObjectKeyMapsOnlyTheLineMarkersPaths(test *testing.T) {
	test.Parallel()
	places := objectPlaces([]string{
		"-O2",
		"-ffile-prefix-map=/tmp/adamic-build-1=/adamic-build",
		"-ffile-prefix-map=/tmp/adamic-build-1/runtime-held=/adamic-runtime",
		"-ffile-prefix-map=/tools/llvm=/adamic-compiler",
		"-ffile-prefix-map=/elsewhere=/elsewhere",
		"-I",
		"/tmp/adamic-build-1/runtime",
	})
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
		test.Errorf("mapped:\n%s\nwant:\n%s", mapped, want)
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
			test.Errorf("%s is keyed as %s, want %s", argument, held, want)
		}
	}
}

// With no trust, a build keys nothing: linkOne preprocesses nothing, writing no main.i or main.d, and the build log's
// line for main.c's compile names no key, with no line for a preprocessing. Trusted as a candidate, the same build
// preprocesses twice and its compile's line carries the key, so the untrusted build's silence isn't the log's. This
// process's own trust is unset, so Build keys nothing.
// Not parallel: it names its build log with test.Setenv.
func TestObjectKeyOffRunsNoPreprocessing(test *testing.T) {
	if objectTrust() != "" {
		test.Skipf("%s is set to %q", objectTrustVariable, objectTrust())
	}
	if objectKeyer() != nil {
		test.Error("with no trust, a build keys main.c")
	}
	library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", objectKeySource, Options{})
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	jobserver, defaultJobserverError := defaultJobserver()
	if defaultJobserverError != nil {
		test.Fatal(defaultJobserverError)
	}
	root := test.TempDir()
	for _, trust := range []string{"", "anything", "candidate"} {
		var compileObject func(
			compile *exec.Cmd,
			source string,
			compileFlags []string,
			program string,
			jobserver *Jobserver,
		) error
		if cache, on := objectCacheFor(
			trust,
			filepath.Join(root, "main"),
			filepath.Join(root, "candidate"),
			filepath.Join(root, "units"),
		); on {
			compileObject = cache.compile
		}
		log := filepath.Join(root, "build-"+trust+".log")
		test.Setenv("ADAMIC_BUILD_LOG", log)
		directory := filepath.Join(root, "build-"+trust)
		if createDirectoryError := os.Mkdir(directory, 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if linkOneError := linkOne(
			directory,
			"clang",
			objectKeySource,
			directory+".program",
			Options{},
			Flags(Options{}),
			library,
			jobserver,
			nil,
			compileObject,
			false,
		); linkOneError != nil {
			test.Fatal(linkOneError)
		}
		var written []string
		for _, name := range []string{"main.i", "main.d", "main.recheck.i"} {
			if _, statError := os.Stat(filepath.Join(directory, name)); statError == nil {
				written = append(written, name)
			}
		}
		logged := objectKeyLogged(test, log)
		if trust != "candidate" && (len(written) > 0 ||
			!slices.Equal(logged["cCompile"], []string{"-"}) ||
			!slices.Equal(logged["outcome"], []string{"off"}) ||
			len(logged) != 2) {
			test.Errorf("trusted %q, the build wrote %v and logged %v", trust, written, logged)
		}
		if trust == "candidate" && (len(written) != 3 ||
			len(logged["cCompile"]) != 1 ||
			logged["cCompile"][0] == "-" ||
			len(logged["cPreprocess"]) != 1 ||
			len(logged["cRecheck"]) != 1) {
			test.Errorf("trusted as a candidate, the build wrote %v and logged %v", written, logged)
		}
	}
}

// Two sources a space apart inside one line preprocess to the same bytes, and under -g compile to other objects, whose
// line tables record columns: they key apart, released and Debug, since every file the preprocessing read is keyed by
// its bytes. Mutant, red: the files' hashes left out of the key.
func TestObjectKeyHoldsEveryColumn(test *testing.T) {
	test.Parallel()
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	root := test.TempDir()
	spaced := strings.Replace(objectKeySource, "return 0;", "return  0;", 1)
	for _, build := range []struct {
		name    string
		options Options
	}{{"release", Options{}}, {"debug", Options{Debug: true}}} {
		// One build at a time, for MovesWithEachInput's reason.
		test.Run(build.name, func(test *testing.T) {
			library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", objectKeySource, build.options)
			if runtimeLibraryForSourceError != nil {
				test.Fatal(runtimeLibraryForSourceError)
			}
			var keyed [2]objectKeyed
			for index, source := range []string{objectKeySource, spaced} {
				keyed[index] = objectKeyBuild(
					test,
					filepath.Join(root, fmt.Sprintf("%s-%d", build.name, index)),
					compiler,
					source,
					build.options,
					Flags(build.options),
					library,
					nil,
					nil,
				)
				if keyed[index].keying == nil {
					test.Fatalf("no key: %v", keyed[index].refusal)
				}
			}
			// A Debug build's markers name its own build directory, so the bytes are compared as the key maps them.
			var mapped [2][]byte
			for index := range keyed {
				mapped[index] = objectMarkersMapped(
					keyed[index].keying.preprocessed,
					objectPlaces(keyed[index].keying.compile.Args[1:]),
				)
			}
			if !bytes.Equal(mapped[0], mapped[1]) {
				test.Fatal("the two sources preprocess to other bytes, so the test shows nothing of the files' hashes")
			}
			if keyed[0].keying.key == keyed[1].keying.key {
				test.Errorf("two sources a space apart key the same: %s", keyed[0].keying.key)
			}
		})
	}
}

// objectCaches are a test's two object directories and its unit cache's, under root.
func objectCaches(test *testing.T, root string, trust string) objectCache {
	test.Helper()
	cache, on := objectCacheFor(
		trust,
		filepath.Join(root, "main"),
		filepath.Join(root, "candidate"),
		filepath.Join(root, "units"),
	)
	if !on {
		test.Fatalf("no cache for %q", trust)
	}
	return cache
}

// objectCacheBuild builds source as options say with linkOne in a build directory named name under root, main.c made by
// cache's compile, or compiled as today when cache is nil, and is the binary's bytes and the outcome main.c's compile
// line says, read from a build log of its own. Its caller sets no test parallel: the log is named by the environment.
// between, when it isn't nil, runs once main.c is keyed and before the cache is looked in.
func objectCacheBuild(
	test *testing.T,
	root string,
	name string,
	cache *objectCache,
	source string,
	options Options,
	library string,
	between func(directory string),
) ([]byte, string) {
	test.Helper()
	jobserver, defaultJobserverError := defaultJobserver()
	if defaultJobserverError != nil {
		test.Fatal(defaultJobserverError)
	}
	directory := filepath.Join(root, name)
	if createDirectoryError := os.Mkdir(directory, 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	log := directory + ".log"
	test.Setenv("ADAMIC_BUILD_LOG", log)
	var compileObject func(
		compile *exec.Cmd,
		source string,
		compileFlags []string,
		program string,
		jobserver *Jobserver,
	) error
	if cache != nil {
		compileObject = func(
			compile *exec.Cmd,
			source string,
			compileFlags []string,
			program string,
			jobserver *Jobserver,
		) error {
			keying, newObjectKeyingError := newObjectKeying(compile, source, compileFlags, program, jobserver)
			if newObjectKeyingError != nil {
				return newObjectKeyingError
			}
			if between != nil {
				between(directory)
			}
			return cache.compileKeyed(keying, compile, jobserver)
		}
	}
	output := directory + ".program"
	if linkOneError := linkOne(
		directory,
		"clang",
		source,
		output,
		options,
		Flags(options),
		library,
		jobserver,
		nil,
		compileObject,
		false,
	); linkOneError != nil {
		test.Fatal(linkOneError)
	}
	binary, readError := os.ReadFile(output)
	if readError != nil {
		test.Fatal(readError)
	}
	logged := objectKeyLogged(test, log)
	if len(logged["outcome"]) != 1 {
		test.Fatalf("%s: main.c's compile lines: %v", name, logged)
	}
	return binary, logged["outcome"][0]
}

// objectEntries are the keys a cache directory holds an object under.
func objectEntries(test *testing.T, directory string) []string {
	test.Helper()
	entries, readDirectoryError := os.ReadDir(directory)
	if errors.Is(readDirectoryError, os.ErrNotExist) {
		return nil
	}
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	var keys []string
	for _, entry := range entries {
		if _, statError := os.Stat(filepath.Join(directory, entry.Name(), "main.o")); statError == nil {
			keys = append(keys, entry.Name())
		}
	}
	return keys
}

// A hit links the binary the uncached build links, byte for byte, released, Debug and sanitized: the cache's object is
// copied to main.o and nothing compiles, and the miss before it stored exactly the object it compiled.
// Not parallel: each build names its log with test.Setenv (objectCacheBuild), which a parallel test may not.
func TestObjectCacheHitIsTheUncachedBinary(test *testing.T) {
	for _, build := range []struct {
		name    string
		options Options
	}{{"release", Options{}}, {"debug", Options{Debug: true}}, {"sanitized", Options{Sanitize: true}}} {
		test.Run(build.name, func(test *testing.T) {
			library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", objectKeySource, build.options)
			if runtimeLibraryForSourceError != nil {
				test.Fatal(runtimeLibraryForSourceError)
			}
			root := test.TempDir()
			cache := objectCaches(test, root, "candidate")
			uncached, outcome := objectCacheBuild(
				test,
				root,
				"uncached",
				nil,
				objectKeySource,
				build.options,
				library,
				nil,
			)
			if outcome != "off" {
				test.Fatalf("the uncached build's compile says %s", outcome)
			}
			for index, want := range []string{"miss", "hit"} {
				binary, outcome := objectCacheBuild(
					test,
					root,
					fmt.Sprintf("cached-%d", index),
					&cache,
					objectKeySource,
					build.options,
					library,
					nil,
				)
				if outcome != want {
					test.Errorf("build %d's compile says %s, want %s", index, outcome, want)
				}
				if !bytes.Equal(binary, uncached) {
					test.Errorf("the %s's binary isn't the uncached build's", want)
				}
			}
		})
	}
}

// Trust holds: a candidate writes its own directory and never main's; a candidate reads an object main stored; main
// never reads one a candidate stored, so a main build after a candidate's misses and stores its own. Mutants, each red:
// a candidate writing main's directory, and main reading the candidate's.
// Not parallel: each build names its log with test.Setenv (objectCacheBuild), which a parallel test may not.
func TestObjectCacheTrust(test *testing.T) {
	library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", objectKeySource, Options{})
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	root := test.TempDir()
	main, candidate := objectCaches(test, root, "main"), objectCaches(test, root, "candidate")
	if _, outcome := objectCacheBuild(
		test,
		root,
		"candidate-first",
		&candidate,
		objectKeySource,
		Options{},
		library,
		nil,
	); outcome != "miss" {
		test.Fatalf("the first candidate build says %s", outcome)
	}
	if held := objectEntries(test, filepath.Join(root, "main")); len(held) > 0 {
		test.Errorf("a candidate's build wrote main's directory: %v", held)
	}
	if held := objectEntries(test, filepath.Join(root, "candidate")); len(held) != 1 {
		test.Fatalf("the candidate's directory holds %v", held)
	}
	if _, outcome := objectCacheBuild(
		test,
		root,
		"main-after-candidate",
		&main,
		objectKeySource,
		Options{},
		library,
		nil,
	); outcome != "miss" {
		test.Errorf("main's build after a candidate's says %s: main read the candidate's object", outcome)
	}
	if held := objectEntries(test, filepath.Join(root, "main")); len(held) != 1 {
		test.Fatalf("main's directory holds %v", held)
	}
	// With main's object stored, a candidate whose own directory holds nothing reads main's.
	if renameError := os.Rename(
		filepath.Join(root, "candidate"),
		filepath.Join(root, "candidate-moved"),
	); renameError != nil {
		test.Fatal(renameError)
	}
	if _, outcome := objectCacheBuild(
		test,
		root,
		"candidate-after-main",
		&candidate,
		objectKeySource,
		Options{},
		library,
		nil,
	); outcome != "hit" {
		test.Errorf("a candidate's build says %s with main's object stored", outcome)
	}
}

// A recheck that fails stores nothing: adamic.h rewritten with the same bytes after main.c is keyed and before it
// compiles leaves the build green and the cache empty, and the next build misses again. Mutant, red: storing on a
// failed recheck.
// Not parallel: each build names its log with test.Setenv (objectCacheBuild), which a parallel test may not.
func TestObjectCacheRecheckFailureStoresNothing(test *testing.T) {
	library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", objectKeySource, Options{})
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	root := test.TempDir()
	copied := objectKeyRuntime(test, library, filepath.Join(root, "runtime"), nil)
	cache := objectCaches(test, root, "candidate")
	if _, outcome := objectCacheBuild(test, root, "moved", &cache, objectKeySource, Options{}, copied, func(
		directory string,
	) {
		header := filepath.Join(directory, "runtime", "adamic.h")
		contents, readError := os.ReadFile(header)
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(header, contents, 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}); outcome != "miss" {
		test.Fatalf("the build says %s", outcome)
	}
	if held := objectEntries(test, filepath.Join(root, "candidate")); len(held) > 0 {
		test.Errorf("a build whose recheck failed stored %v", held)
	}
	if _, outcome := objectCacheBuild(
		test,
		root,
		"again",
		&cache,
		objectKeySource,
		Options{},
		copied,
		nil,
	); outcome != "miss" {
		test.Errorf("the build after it says %s", outcome)
	}
}

// A wrong object planted under the right key in main's directory is caught where ADAMIC_NATIVE_UNIT_CHECK asks every
// hit compiled: the build links the cold object, the uncached build's binary, says the unit cache's mismatch line
// naming main.c and its key, and marks the unit cache off, which turns the object cache off too. Mutant, red: the
// forced check skipped.
// Not parallel: each build names its log with test.Setenv (objectCacheBuild), which a parallel test may not.
func TestObjectCacheForcedCheckCatchesAPlantedObject(test *testing.T) {
	library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", objectKeySource, Options{})
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	root := test.TempDir()
	cache := objectCaches(test, root, "main")
	uncached, _ := objectCacheBuild(test, root, "uncached", nil, objectKeySource, Options{}, library, nil)
	if _, outcome := objectCacheBuild(
		test,
		root,
		"stored",
		&cache,
		objectKeySource,
		Options{},
		library,
		nil,
	); outcome != "miss" {
		test.Fatalf("the first build says %s", outcome)
	}
	held := objectEntries(test, filepath.Join(root, "main"))
	if len(held) != 1 {
		test.Fatalf("main's directory holds %v", held)
	}
	// Planted whole, its sum beside it, so the sum check passes it and only the forced check can catch it.
	planted := []byte("not an object\n")
	if writeError := os.WriteFile(filepath.Join(root, "main", held[0], "main.o"), planted, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	if writeError := os.WriteFile(
		filepath.Join(root, "main", held[0], objectSumName),
		fmt.Appendf(nil, "%x\n", sha256.Sum256(planted)),
		0o644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	test.Setenv(unitCheckVariable, "every")
	binary, outcome := objectCacheBuild(test, root, "checked", &cache, objectKeySource, Options{}, library, nil)
	if outcome != "miss" {
		test.Errorf("the checked build says %s", outcome)
	}
	if !bytes.Equal(binary, uncached) {
		test.Error("the checked build didn't link the cold object")
	}
	log, readError := os.ReadFile(filepath.Join(root, "checked.log"))
	if readError != nil {
		test.Fatal(readError)
	}
	if want := regexp.MustCompile(`(?m)^` +
		regexp.QuoteMeta(UnitMismatchPrefix) +
		`unit main\.c key ` +
		held[0] +
		` cached [0-9a-f]{64} cold [0-9a-f]{64}$`); !want.Match(log) {
		test.Errorf("the build log says no mismatch line:\n%s", log)
	}
	if why, off := unitCacheOff(filepath.Join(root, "units")); !off {
		test.Error("the mismatch didn't mark the unit cache off")
	} else if !strings.Contains(why, "main.c") {
		test.Errorf("the mark says %s", why)
	}
}

// An entry that isn't the bytes its sum names is never linked: an object truncated, one with a byte flipped, and one
// whose sum file is gone each compile cold as a miss, link the uncached build's binary, and are replaced by the object
// that compile stores, so the build after hits. A crash or a disk mid-write leaves such an entry. Mutant, red: the sum
// check dropped, which links the torn object.
// Not parallel: each build names its log with test.Setenv (objectCacheBuild), which a parallel test may not.
func TestObjectCacheNeverLinksATornEntry(test *testing.T) {
	library, runtimeLibraryForSourceError := RuntimeLibraryForSource("", objectKeySource, Options{})
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	root := test.TempDir()
	cache := objectCaches(test, root, "main")
	uncached, _ := objectCacheBuild(test, root, "uncached", nil, objectKeySource, Options{}, library, nil)
	if _, outcome := objectCacheBuild(
		test,
		root,
		"stored",
		&cache,
		objectKeySource,
		Options{},
		library,
		nil,
	); outcome != "miss" {
		test.Fatalf("the first build says %s", outcome)
	}
	held := objectEntries(test, filepath.Join(root, "main"))
	if len(held) != 1 {
		test.Fatalf("main's directory holds %v", held)
	}
	entry := filepath.Join(root, "main", held[0])
	for index, tear := range []struct {
		name string
		tear func(object []byte)
	}{
		{"truncated", func(object []byte) {
			if writeError := os.WriteFile(
				filepath.Join(entry, "main.o"),
				object[:len(object)/2],
				0o644,
			); writeError != nil {
				test.Fatal(writeError)
			}
		}},
		{"a byte flipped", func(object []byte) {
			flipped := slices.Clone(object)
			flipped[len(flipped)/2] ^= 0xff
			if writeError := os.WriteFile(filepath.Join(entry, "main.o"), flipped, 0o644); writeError != nil {
				test.Fatal(writeError)
			}
		}},
		{"its sum gone", func([]byte) {
			if removeError := os.Remove(filepath.Join(entry, objectSumName)); removeError != nil {
				test.Fatal(removeError)
			}
		}},
	} {
		object, readError := os.ReadFile(filepath.Join(entry, "main.o"))
		if readError != nil {
			test.Fatal(readError)
		}
		tear.tear(object)
		binary, outcome := objectCacheBuild(
			test,
			root,
			fmt.Sprintf("torn-%d", index),
			&cache,
			objectKeySource,
			Options{},
			library,
			nil,
		)
		if outcome != "miss" {
			test.Errorf("%s: the build says %s", tear.name, outcome)
		}
		if !bytes.Equal(binary, uncached) {
			test.Errorf("%s: the build didn't link the cold object", tear.name)
		}
		if _, _, whole := objectEntry(entry); !whole {
			test.Errorf("%s: the torn entry wasn't replaced", tear.name)
		}
		if _, outcome := objectCacheBuild(
			test,
			root,
			fmt.Sprintf("after-%d", index),
			&cache,
			objectKeySource,
			Options{},
			library,
			nil,
		); outcome != "hit" {
			test.Errorf("%s: the build after its replacement says %s", tear.name, outcome)
		}
	}
}

// objectPartsMoved says which of two keys' parts are in one and not the other, each part on a line of its own.
func objectPartsMoved(base []string, other []string) string {
	var moved strings.Builder
	for _, part := range base {
		if !slices.Contains(other, part) {
			fmt.Fprintf(&moved, "  only the base's: %s\n", part)
		}
	}
	for _, part := range other {
		if !slices.Contains(base, part) {
			fmt.Fprintf(&moved, "  only this one's: %s\n", part)
		}
	}
	return moved.String()
}
