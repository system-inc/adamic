package native

import (
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"

	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"github.com/system-inc/adamic/bridge/tsgo/spec"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/toolchain"
)

func UsesTSGo(program *ir.Program) bool {
	for _, function := range program.Functions {
		if _, found := spec.Name(function.Name); found {
			return true
		}
	}
	return false
}

// TSGoC replaces only the compiler-generated library bodies. Kept here so
// this unit does not edit the shared emitter. Signatures, calls, snapshots and
// cleanup remain the ordinary emitter's output. The library contracts borrow
// inputs, mutate no Adamic value and return fresh values, as the IR bodies say.
// A changed emitter layout is an error, never a guessed replacement.
func TSGoC(program *ir.Program) (string, error) {
	source := C(program)
	emitter := &emitter{program: program, regions: planRegions(program)}
	for index, function := range program.Functions {
		name, found := spec.Name(function.Name)
		if !found {
			continue
		}
		for _, inRegion := range emitter.regionVariants(index) {
			emitter.inRegion = inRegion
			signature := "static " + emitter.signature(index) + " {\n"
			if strings.Count(source, signature) != 1 {
				return "", fmt.Errorf("native: missing unique %s library body", name)
			}
			start := strings.Index(source, signature) + len(signature)
			end := strings.Index(source[start:], "\n}\n")
			if end < 0 {
				return "", fmt.Errorf("native: missing end of %s library body", name)
			}
			arguments := []string{}
			if inRegion {
				arguments = append(arguments, "region")
			}
			for _, parameter := range function.Parameters {
				if !program.Locals[parameter].Borrowed && program.Locals[parameter].Type.IsReference() {
					return "", fmt.Errorf("native: library parameter is not borrowed")
				}
				arguments = append(arguments, emitter.localName(parameter))
			}
			runtimeName := map[string]string{"tsgoProgram": "program", "tsgoQuery": "query", "tsgoRelease": "release", "tsgoTypeParts": "type_parts", "tsgoInspect": "inspect"}[name]
			if inRegion {
				if name != "tsgoQuery" {
					return "", fmt.Errorf("native: unexpected region version of %s", name)
				}
				runtimeName += "_in"
			}
			body := "\tADAMIC_CHECK_STACK_NAMED(" + cString(function.Name) + ");\n\t"
			if function.Returns != 0 {
				body += "return "
			}
			body += "adamic_tsgo_" + runtimeName + "(" + strings.Join(arguments, ", ") + ");"
			source = source[:start] + body + source[start+end:]
		}
	}
	return "#include \"tsgo_runtime.h\"\n" + source, nil
}

// BuildTSGo is the native build with an explicitly selected checker archive.
// The public ABI header is embedded beside the archive's source, never copied
// from cgo's generated implementation header.
func BuildTSGo(source, output, archive string, options Options) error {
	archive, err := checkerArchivePath(archive)
	if err != nil {
		return err
	}
	directory, err := os.MkdirTemp("", "adamic-tsgo-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0o644); err != nil {
		return err
	}
	// clang is given no absolute path to read, as Build gives it none: it compiles each unit by its name from inside the
	// directory, so a sanitized binary's globals don't name it (landable-16c's tree, Oct 11: bridge-test-wrong-native,
	// -healthy-region and -native-asan named /tmp/adamic-tsgo-<random>/adamic.c and every runtime unit, random per build).
	units := []string{"main.c"}
	entries, err := runtime.ReadDir("runtime")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		contents, err := runtime.ReadFile("runtime/" + entry.Name())
		if err != nil {
			return err
		}
		path := filepath.Join(directory, entry.Name())
		if err := os.WriteFile(path, contents, 0o644); err != nil {
			return err
		}
		if strings.HasSuffix(path, ".c") {
			units = append(units, entry.Name())
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "tsgo.h"), bridge.Header, 0o644); err != nil {
		return err
	}
	if err := os.Symlink(archive, filepath.Join(directory, checkerArchive)); err != nil {
		return err
	}
	// The output is where the binary is written, never read into it (macOS's ld names it in the code signature by its
	// base name alone), so it may be absolute; it must be, since clang runs from the build directory.
	output, err = filepath.Abs(output)
	if err != nil {
		return err
	}
	arguments := append(append(linkFlags(tsgoFlags(source, options), options), prefixMaps(directory, "/adamic-build")...), "-o", output)
	arguments = append(arguments, units...)
	arguments = append(append(arguments, checkerArchive, "-lm", "-lpthread", "-ldl"), objectPrefixFlags()...)
	command := toolchain.CommandIn(directory, "clang", arguments...)
	var combined []byte
	if err := withCompilerToken(func() (err error) { combined, err = command.CombinedOutput(); return err }); err != nil {
		return fmt.Errorf("native: clang with checker archive failed: %w\n%s", err, combined)
	}
	return nil
}

// checkerArchive is the checker archive's name on a link line: a link to it in the directory clang links from, so the
// line names no place on this machine, as Build's names the runtime archive through a link (#m27xj48: on macOS a
// checker build's debug map named the archive by its absolute path, so the binary differed with where it was built).
const checkerArchive = "checker.a"

// checkerArchivePath is archive, a path from the working directory, as the absolute path a link to it in a build
// directory needs; only that link ever holds it.
func checkerArchivePath(archive string) (string, error) {
	archive, err := filepath.Abs(archive)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(archive); err != nil {
		return "", fmt.Errorf("native: checker archive: %w", err)
	}
	return archive, nil
}

// objectPrefixFlags take the link's working directory off the paths its debug map records. On macOS ld records each
// object it links that carries debug info by path, for a debugger to find its DWARF there (the stabs' N_OSO), an
// archive's member by the archive's path, made absolute against the working directory when it isn't: the checker
// archive's go.o carries Go's DWARF, so checker.a(go.o) would read as the build directory's path, random per build.
// -oso_prefix takes the working directory back off. An ELF linker copies the DWARF in and records no input's path.
func objectPrefixFlags() []string {
	if goruntime.GOOS == "darwin" {
		return []string{"-Wl,-oso_prefix,."}
	}
	return nil
}

// tsgoFlags compiles the checker-archive build. tsgo_runtime.h, included first, brings adamic.h in
// before main.c's own #defines, so the program's features come as flags, as they do for the runtime
// library.
func tsgoFlags(source string, options Options) []string {
	return append(sourceFlags(source, options), "-DADAMIC_TSGO")
}
