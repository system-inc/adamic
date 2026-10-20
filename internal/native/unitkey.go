package native

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"debug/elf"
	"debug/macho"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"syscall"
)

// A split unit's object is keyed by everything its compile reads, found before clang compiles it (#m62spx0): the unit
// by its name and bytes, every file of the directory it compiles in, the resolution of every header name the headers it
// reaches spell, clang's build as its driver decides a compile, the flags, the target, and the environment clang reads.
// A compiler change that leaves a unit's C alone then compiles nothing there. Complete before minimal: a part that could
// be narrower is kept whole, and anything the key can't account for compiles uncached, saying why (errUncachedUnit).

// errUncachedUnit is a reason a build's units compile uncached: the key can't account for something they read.
var errUncachedUnit = errors.New("native: units compile uncached")

// unitEnvironmentKeyed are the variables clang's driver or cc1 reads that can change an object's bytes, each keyed by its
// value when the compile's environment sets it: macOS's developer directory, SDK and deployment targets, the time
// `__DATE__` and `__TIME__` would read (none reaches a unit, which refuses them, but a header's could), and Apple's
// debug options, which write the command line or a prefix map into debug info.
var unitEnvironmentKeyed = []string{"DEVELOPER_DIR", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET", "IPHONEOS_DEPLOYMENT_TARGET",
	"TVOS_DEPLOYMENT_TARGET", "WATCHOS_DEPLOYMENT_TARGET", "XROS_DEPLOYMENT_TARGET", "SOURCE_DATE_EPOCH", "TZ",
	"RC_DEBUG_OPTIONS", "RC_DEBUG_PREFIX_MAP", "COMPILER_PATH"}

// unitEnvironmentRefused are the variables under which the units compile uncached when the compile's environment sets
// them: the include-path family adds search directories the resolution table doesn't walk, and CCC_OVERRIDE_OPTIONS (as
// every CCC_ setting) rewrites clang's arguments.
var unitEnvironmentRefused = []string{"CPATH", "C_INCLUDE_PATH", "CPLUS_INCLUDE_PATH", "OBJC_INCLUDE_PATH", "OBJCPLUS_INCLUDE_PATH"}

// unitEnvironmentParts are environment's settings clang reads, as key parts, or errUncachedUnit naming one the key can't
// take. environment is exactly what the compile runs under.
func unitEnvironmentParts(environment []string) ([]string, error) {
	values := map[string]string{}
	for _, setting := range environment {
		name, value, _ := strings.Cut(setting, "=")
		values[name] = value
	}
	var parts []string
	for name := range values {
		if slices.Contains(unitEnvironmentRefused, name) || strings.HasPrefix(name, "CCC_") {
			return nil, fmt.Errorf("%w: the compile's environment sets %s", errUncachedUnit, name)
		}
	}
	for _, name := range unitEnvironmentKeyed {
		if value, set := values[name]; set {
			parts = append(parts, "environment "+name+"="+value)
		}
	}
	return parts, nil
}

// unitClangFacts are what clang's build decides of a compile under some flags, read once a process for each compiler,
// flags and environment.
type unitClangFacts struct {
	once  sync.Once
	parts []string
	// search is clang's search path for `<name>` includes, in order, as `clang -E -v` lists it.
	search []searchDirectory
	err    error
}

var unitClangs sync.Map

// unitClang is clang's facts for compiler under flags and environment (unitClangFacts), read again when the program
// the path resolves to, or a config file beside it, is another file or was rewritten, so a clang swapped or configured
// under a long-lived process isn't keyed by the facts of the one before it.
func unitClang(compiler string, flags []string, environment []string) (*unitClangFacts, error) {
	identity := []string{compiler}
	if program, err := filepath.EvalSymlinks(compiler); err == nil {
		configs, _ := filepath.Glob(filepath.Join(filepath.Dir(program), "*.cfg"))
		spelled, _ := filepath.Glob(filepath.Join(filepath.Dir(compiler), "*.cfg"))
		for _, path := range append(append([]string{program}, configs...), spelled...) {
			identity = append(identity, fileIdentity(path))
		}
	}
	value, _ := unitClangs.LoadOrStore(partsKey(append(append(identity, flags...), environment...)), &unitClangFacts{})
	facts := value.(*unitClangFacts)
	facts.once.Do(func() { facts.parts, facts.search, facts.err = readUnitClang(compiler, flags, environment) })
	return facts, facts.err
}

// readUnitClang reads clang's build: its program by the sha256 of what it holds, the config files beside it, its
// version as it prints it without the place it's installed, its resource directory's headers and sanitizer lists, the
// command its driver hands cc1 for a probe compile under flags (the target, its CPU and features, and every default the
// driver adds), and its search path.
func readUnitClang(compiler string, flags []string, environment []string) ([]string, []searchDirectory, error) {
	program, err := filepath.EvalSymlinks(compiler)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: resolving clang: %v", errUncachedUnit, err)
	}
	digest, err := fileDigest(program)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: reading clang: %v", errUncachedUnit, err)
	}
	parts := []string{"clang " + digest}
	libraries, err := programLibraries(program)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: reading clang's shared libraries: %v", errUncachedUnit, err)
	}
	for _, library := range libraries {
		// An LLVM or clang library holds code generation the program's own digest doesn't: the house's clang links none.
		if lower := strings.ToLower(library); strings.Contains(lower, "llvm") || strings.Contains(lower, "clang") {
			return nil, nil, fmt.Errorf("%w: clang links %s, whose code the key doesn't read", errUncachedUnit, library)
		}
		parts = append(parts, "library "+library)
	}
	// clang reads <name>.cfg and <triple>-<name>.cfg beside its program, and each can add flags.
	configs, err := filepath.Glob(filepath.Join(filepath.Dir(program), "*.cfg"))
	if err != nil {
		return nil, nil, err
	}
	if spelled := filepath.Dir(compiler); spelled != filepath.Dir(program) {
		more, err := filepath.Glob(filepath.Join(spelled, "*.cfg"))
		if err != nil {
			return nil, nil, err
		}
		configs = append(configs, more...)
	}
	for _, config := range configs {
		digest, err := fileDigest(config)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: reading clang's config %s: %v", errUncachedUnit, config, err)
		}
		parts = append(parts, "config "+config+" "+digest)
	}
	// Every run is from inside the probe's directory, so the compilation directory the driver hands cc1 is the probe's,
	// written <probe> below, and not wherever this process runs.
	probe, err := os.MkdirTemp("", "adamic-units-clang-")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(probe)
	run := func(arguments ...string) ([]byte, error) {
		command := exec.Command(compiler, arguments...)
		command.Dir, command.Env = probe, environment
		output, err := command.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("%w: clang %s: %v\n%s", errUncachedUnit, strings.Join(arguments, " "), err, output)
		}
		return output, nil
	}
	version, err := run("--version")
	if err != nil {
		return nil, nil, err
	}
	var versionLines []string
	for _, line := range strings.Split(string(version), "\n") {
		if !strings.HasPrefix(line, "InstalledDir:") && line != "" {
			versionLines = append(versionLines, line)
		}
	}
	parts = append(parts, "version "+strings.Join(versionLines, "\n"))
	resource, err := run("-print-resource-dir")
	if err != nil {
		return nil, nil, err
	}
	resourceDirectory := strings.TrimSpace(string(resource))
	// The resource directory's include tree holds clang's own headers, and its share tree the sanitizers' ignore lists a
	// sanitized compile reads; its lib tree is only linked.
	for _, tree := range []string{"include", "share"} {
		lines, err := treeDigests(filepath.Join(resourceDirectory, tree))
		if err != nil {
			return nil, nil, fmt.Errorf("%w: reading clang's resource %s: %v", errUncachedUnit, tree, err)
		}
		for _, line := range lines {
			parts = append(parts, "resource "+tree+"/"+line)
		}
	}
	if err := os.WriteFile(filepath.Join(probe, "probe.c"), nil, 0o644); err != nil {
		return nil, nil, err
	}
	driver, err := run(append(slices.Clone(flags), "-###", "-c", "probe.c", "-o", "probe.o")...)
	if err != nil {
		return nil, nil, err
	}
	for _, line := range strings.Split(string(driver), "\n") {
		// The cc1 line and the others are quoted arguments; the version and thread lines are the --version part's.
		if !strings.HasPrefix(line, " \"") {
			continue
		}
		line = strings.ReplaceAll(line, probe, "<probe>")
		if real, err := filepath.EvalSymlinks(probe); err == nil {
			line = strings.ReplaceAll(line, real, "<probe>")
		}
		parts = append(parts, "driver"+line)
	}
	verbose, err := run(append(slices.Clone(flags), "-E", "-v", "-x", "c", os.DevNull)...)
	if err != nil {
		return nil, nil, err
	}
	search, err := searchPath(verbose)
	if err != nil {
		return nil, nil, err
	}
	for _, directory := range search {
		parts = append(parts, fmt.Sprintf("search %s framework=%t", directory.path, directory.framework))
	}
	return parts, search, nil
}

// A searchDirectory is one directory of clang's `<name>` search path: a framework directory (macOS's, marked
// "(framework directory)") resolves `<Foo/Bar.h>` to Foo.framework/Headers/Bar.h or Foo.framework/PrivateHeaders/Bar.h
// under it, and any other directory to Foo/Bar.h.
type searchDirectory struct {
	path      string
	framework bool
}

// candidates are the places clang looks for name in directory.
func (directory searchDirectory) candidates(name string) []string {
	if !directory.framework {
		return []string{filepath.Join(directory.path, name)}
	}
	framework, rest, nested := strings.Cut(name, "/")
	if !nested {
		return nil
	}
	return []string{filepath.Join(directory.path, framework+".framework", "Headers", rest), filepath.Join(directory.path, framework+".framework", "PrivateHeaders", rest)}
}

// searchPath is the `<name>` search path `clang -E -v` lists, in order, framework directories included. A quote search
// directory beyond the including file's own, which only -iquote adds, refuses: the resolution table doesn't walk it.
func searchPath(verbose []byte) ([]searchDirectory, error) {
	var search []searchDirectory
	state := ""
	scanner := bufio.NewScanner(bytes.NewReader(verbose))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "#include \"...\" search starts here:":
			state = "quote"
		case line == "#include <...> search starts here:":
			state = "angle"
		case line == "End of search list.":
			state = "done"
		case strings.HasPrefix(line, " ") && state == "quote":
			return nil, fmt.Errorf("%w: clang searches %s for quoted includes", errUncachedUnit, strings.TrimSpace(line))
		case strings.HasPrefix(line, " ") && state == "angle":
			directory, mark, _ := strings.Cut(strings.TrimSpace(line), " (")
			search = append(search, searchDirectory{path: filepath.Clean(directory), framework: mark == "framework directory)"})
		}
	}
	if state != "done" || len(search) == 0 {
		return nil, fmt.Errorf("%w: clang -E -v listed no search path", errUncachedUnit)
	}
	return search, nil
}

// unitIncludes finds every header name a header asks for: an #include, #include_next or #import directive at a line's
// start, or a __has_include or __has_include_next anywhere, each by a quoted or bracketed name. A form with anything else
// there, a macro, refuses. One in a comment is read too, which only adds candidates.
var unitIncludes = regexp.MustCompile(`(?m)^[ \t]*#[ \t]*(include|include_next|import)\b[ \t]*(.?)|__has_include(_next)?\s*\(\s*(.?)`)

// unitDefinesName is the start of a line that defines or undefines the macro named next.
var unitDefinesName = regexp.MustCompile(`^[ \t]*#[ \t]*(define|undef)[ \t]+$`)

// unitNameOperand is one quoted or bracketed header name, as an include spells it.
var unitNameOperand = regexp.MustCompile(`^(?:"([^"\n]+)"|<([^>\n]+)>)`)

// unitEmbeds are the forms that read or ask after a file the resolution table doesn't follow: #embed and __has_embed.
var unitEmbeds = regexp.MustCompile(`(?m)^[ \t]*#[ \t]*embed\b|__has_embed\s*\(`)

// unitTime are the predefined macros whose value is when the compile ran.
var unitTime = regexp.MustCompile(`__(DATE|TIME|TIMESTAMP)__`)

// unitHeaderParts are the resolution table of a build's units (#m62spx0): every file clang opens for a probe holding
// only first, the line every unit begins with (`#include "units.h"`), compiled under flags in a directory holding common (units.h and the
// runtime's headers) and nothing else; and for every header name any header in that closure spells, by an #include, an
// #include_next, a __has_include or a __has_include_next, that name's candidate in the including header's own
// directory, when it spells it quoted, and in every search directory, each absent or by the sha256 of what it holds. A
// header planted, removed or changed under any name a header asks for, in any place clang would look, moves the key,
// whether clang found it, found it missing, or took the next. The probe is the build's one clang run before its units
// compile, and common's own files are keyed by what they hold elsewhere (unitKeyParts).
func unitHeaderParts(first string, common []runtimeFile, flags []string, compiler string, environment []string, search []searchDirectory) ([]string, error) {
	probe, err := os.MkdirTemp("", "adamic-units-headers-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(probe)
	for _, file := range common {
		if err := os.WriteFile(filepath.Join(probe, file.name), file.contents, 0o644); err != nil {
			return nil, err
		}
	}
	if err := os.WriteFile(filepath.Join(probe, "probe.c"), []byte(first), 0o644); err != nil {
		return nil, err
	}
	command := exec.Command(compiler, append(slices.Clone(flags), "-M", "probe.c")...)
	command.Dir, command.Env = probe, environment
	rules, err := command.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return nil, fmt.Errorf("%w: listing the units' headers: %v\n%s", errUncachedUnit, err, exit.Stderr)
		}
		return nil, fmt.Errorf("%w: listing the units' headers: %v", errUncachedUnit, err)
	}
	resolvedProbe, err := filepath.EvalSymlinks(probe)
	if err != nil {
		return nil, err
	}
	place := func(name string) string {
		if !filepath.IsAbs(name) {
			name = filepath.Join(probe, name)
		}
		return filepath.Clean(name)
	}
	ownName := func(name string) (string, bool) {
		for _, directory := range []string{probe, resolvedProbe} {
			if relative, err := filepath.Rel(directory, name); err == nil && !strings.HasPrefix(relative, "..") {
				return relative, true
			}
		}
		return "", false
	}
	var parts []string
	closure := map[string]bool{}
	for _, name := range dependencies(rules) {
		closure[place(name)] = true
	}
	candidates := map[string]bool{}
	for _, header := range slices.Sorted(maps.Keys(closure)) {
		relative, own := ownName(header)
		if own && relative == "probe.c" {
			continue
		}
		contents, err := os.ReadFile(header)
		if err != nil {
			return nil, fmt.Errorf("%w: reading header %s: %v", errUncachedUnit, header, err)
		}
		if unitTime.Match(contents) {
			return nil, fmt.Errorf("%w: %s reads the time of the compile", errUncachedUnit, header)
		}
		if unitEmbeds.Match(contents) {
			return nil, fmt.Errorf("%w: %s reads a file by a form the key can't follow (#embed or __has_embed)", errUncachedUnit, header)
		}
		if !own {
			parts = append(parts, fmt.Sprintf("system %s %x", header, sha256.Sum256(contents)))
		}
		for _, match := range unitIncludes.FindAllSubmatchIndex(contents, -1) {
			// A directive's operand is its second group; a __has_include's, its fourth.
			operand := match[4]
			if match[2] < 0 {
				operand = match[8]
				// A header that defines or undefines the name itself, as macOS's sys/cdefs.h does for a compiler without
				// it (`#define __has_include(x) 0`), asks after no header there.
				if unitDefinesName.Match(contents[bytes.LastIndexByte(contents[:match[0]], '\n')+1 : match[0]]) {
					continue
				}
			}
			name := unitNameOperand.FindSubmatch(contents[operand:])
			if name == nil {
				return nil, fmt.Errorf("%w: %s spells an include clang resolves by a macro, which the key can't follow", errUncachedUnit, header)
			}
			if quoted := string(name[1]); quoted != "" {
				candidates[filepath.Join(filepath.Dir(header), quoted)] = true
				for _, directory := range search {
					for _, candidate := range directory.candidates(quoted) {
						candidates[candidate] = true
					}
				}
				continue
			}
			for _, directory := range search {
				for _, candidate := range directory.candidates(string(name[2])) {
					candidates[candidate] = true
				}
			}
		}
	}
	for _, candidate := range slices.Sorted(maps.Keys(candidates)) {
		if relative, own := ownName(candidate); own {
			// The units' own directory holds only common, keyed by name and contents; a name it doesn't hold is absent.
			parts = append(parts, "own "+relative+" "+ownDigest(common, relative))
			continue
		}
		info, err := os.Stat(candidate)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			parts = append(parts, "candidate "+candidate+" absent")
		case err != nil:
			return nil, fmt.Errorf("%w: looking for %s: %v", errUncachedUnit, candidate, err)
		case info.IsDir():
			parts = append(parts, "candidate "+candidate+" directory")
		default:
			digest, err := fileDigest(candidate)
			if err != nil {
				return nil, fmt.Errorf("%w: reading %s: %v", errUncachedUnit, candidate, err)
			}
			parts = append(parts, "candidate "+candidate+" "+digest)
		}
	}
	return parts, nil
}

// ownDigest is the sha256 of the file named relative among common, or "absent".
func ownDigest(common []runtimeFile, relative string) string {
	for _, file := range common {
		if file.name == filepath.ToSlash(relative) {
			return fmt.Sprintf("%x", sha256.Sum256(file.contents))
		}
	}
	return "absent"
}

// unitSourceRefusal says why a unit's own source can't be keyed by its bytes alone: a first line other than first, the
// one the build's headers were probed by, a directive past it, or a macro that reads the time or asks after a header.
// "" when it can.
func unitSourceRefusal(unit compilationUnit, first string) string {
	body, begins := strings.CutPrefix(unit.source, first)
	if !begins {
		return fmt.Sprintf("%s doesn't begin with the line its build's headers were probed by", unit.name)
	}
	tokens, err := cTokens(body)
	if err != nil {
		return err.Error()
	}
	for _, token := range tokens {
		if strings.HasPrefix(token.text, "#") {
			return fmt.Sprintf("%s holds a directive past its first line", unit.name)
		}
		if unitTime.MatchString(token.text) || strings.HasPrefix(token.text, "__has_include") || strings.HasPrefix(token.text, "__has_embed") {
			return fmt.Sprintf("%s reads %s", unit.name, token.text)
		}
	}
	return ""
}

// unitKeying is everything a build's units are keyed by but each unit's own name and bytes, and the environment they
// compile under, read once a build before any of them compiles: the platform, clang's build (unitClang), the
// environment clang reads, the flags, how a unit is compiled, the resolution table (unitHeaderParts), and every file of
// the directory a unit compiles in, by name and contents.
type unitKeying struct {
	parts []string
	// first is the line every unit begins with, the probe's whole source.
	first       string
	environment []string
}

// newUnitKeying reads a build's unitKeying, or errUncachedUnit naming what the key can't account for.
func newUnitKeying(first string, common []runtimeFile, flags []string, compiler string, environment []string) (*unitKeying, error) {
	for _, flag := range flags {
		// Clang modules read module maps and a module cache the resolution table doesn't walk.
		if strings.HasPrefix(flag, "-fmodule") || strings.HasPrefix(flag, "-fimplicit-module") {
			return nil, fmt.Errorf("%w: the flags use clang modules (%s)", errUncachedUnit, flag)
		}
	}
	environmentParts, err := unitEnvironmentParts(environment)
	if err != nil {
		return nil, err
	}
	clang, err := unitClang(compiler, flags, environment)
	if err != nil {
		return nil, err
	}
	headers, err := unitHeaderParts(first, common, flags, compiler, environment, clang.search)
	if err != nil {
		return nil, err
	}
	// v4: keyed before clang runs, by the unit's bytes and its headers' resolution, never by its preprocessed bytes.
	parts := []string{"adamic-units-v4", "os " + goruntime.GOOS, "arch " + goruntime.GOARCH}
	parts = append(parts, clang.parts...)
	parts = append(parts, environmentParts...)
	for _, flag := range flags {
		parts = append(parts, "flag "+flag)
	}
	parts = append(parts, "compile "+strings.Join(unitCompileArguments("<unit>", "<directory>"), " "))
	parts = append(parts, headers...)
	files := slices.Clone(common)
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	for _, file := range files {
		parts = append(parts, fmt.Sprintf("file %s %x", file.name, sha256.Sum256(file.contents)))
	}
	return &unitKeying{parts: parts, first: first, environment: environment}, nil
}

// key is unit's key under keying: its parts and the unit by name and contents.
func (keying *unitKeying) key(unit compilationUnit) string {
	return partsKey(append(slices.Clone(keying.parts), fmt.Sprintf("unit %s %x", unit.name, sha256.Sum256([]byte(unit.source)))))
}

// unitCompileArguments are what follow a unit's flags when it compiles: the directory it compiles in mapped to one name
// in everything the object records of a path, and the unit compiled to unit.o there.
func unitCompileArguments(name string, directory string) []string {
	return append(prefixMaps(directory, "/adamic-units"), "-c", name, "-o", "unit.o")
}

// fileDigest is the sha256 of what path holds, in hex.
func fileDigest(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(contents)), nil
}

// treeDigests is every regular file under root, by its path from root and the sha256 of what it holds, sorted; a link
// is followed, and a tree that isn't there is none.
func treeDigests(root string) ([]string, error) {
	var lines []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) && path == root {
			return filepath.SkipDir
		}
		if err != nil || entry.IsDir() {
			return err
		}
		digest, err := fileDigest(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		lines = append(lines, filepath.ToSlash(relative)+" "+digest)
		return nil
	})
	sort.Strings(lines)
	return lines, err
}

// programLibraries are the shared libraries program names for the dynamic loader, as its ELF or Mach-O header lists
// them, sorted; none for a static program, and an error for one it can't read as either.
func programLibraries(program string) ([]string, error) {
	var libraries []string
	if file, err := elf.Open(program); err == nil {
		defer file.Close()
		libraries, err = file.ImportedLibraries()
		if err != nil {
			return nil, err
		}
	} else if file, err := macho.Open(program); err == nil {
		defer file.Close()
		libraries, err = file.ImportedLibraries()
		if err != nil {
			return nil, err
		}
	} else if fat, err := macho.OpenFat(program); err == nil {
		defer fat.Close()
		for _, architecture := range fat.Arches {
			more, err := architecture.ImportedLibraries()
			if err != nil {
				return nil, err
			}
			libraries = append(libraries, more...)
		}
	} else {
		return nil, fmt.Errorf("%s is neither ELF nor Mach-O", program)
	}
	slices.Sort(libraries)
	return slices.Compact(libraries), nil
}

// fileIdentity names the file at path as the system holds it, its place, size, modification time, device and inode, so
// one rewritten or replaced there names another; a file that can't be read is named so.
func fileIdentity(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return path + " " + err.Error()
	}
	identity := fmt.Sprintf("%s %d %d", path, info.Size(), info.ModTime().UnixNano())
	if system, ok := info.Sys().(*syscall.Stat_t); ok {
		identity += fmt.Sprintf(" %d %d", system.Dev, system.Ino)
	}
	return identity
}
