package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"github.com/system-inc/adamic/bridge/tsgo/spec"
	"github.com/system-inc/adamic/internal/ir"
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
			body := "\tADAMIC_CHECK_STACK();\n\t"
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
	archive, err := filepath.Abs(archive)
	if err != nil {
		return err
	}
	if _, err := os.Stat(archive); err != nil {
		return fmt.Errorf("native: checker archive: %w", err)
	}
	directory, err := os.MkdirTemp("", "adamic-tsgo-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0o644); err != nil {
		return err
	}
	units := []string{filepath.Join(directory, "main.c")}
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
			units = append(units, path)
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "tsgo.h"), bridge.Header, 0o644); err != nil {
		return err
	}
	arguments := append(tsgoFlags(source, options), "-o", output)
	arguments = append(arguments, units...)
	arguments = append(arguments, archive, "-lm", "-lpthread", "-ldl")
	combined, err := exec.Command("clang", arguments...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("native: clang with checker archive failed: %w\n%s", err, combined)
	}
	return nil
}

// tsgoFlags compiles the checker-archive build. tsgo_runtime.h, included first, brings adamic.h in
// before main.c's own #defines, so the program's features come as flags, as they do for the runtime
// library.
func tsgoFlags(source string, options Options) []string {
	return append(sourceFlags(source, options), "-DADAMIC_TSGO")
}
