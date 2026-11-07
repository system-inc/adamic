package native

import (
	"bytes"
	"context"
	"debug/dwarf"
	"debug/elf"
	"debug/macho"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/toolchain"
)

func debugProgram(test *testing.T, path string, debug bool) (*ir.Program, error) {
	test.Helper()
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		return nil, loadError
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return lower.LowerWithOptions(ctx, loaded, lower.Options{Debug: debug})
}

// Strip metadata recursively, including statements held inside expressions.
func stripSourceLines(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return value
		}
		result := reflect.New(value.Type()).Elem()
		result.Set(stripSourceLines(value.Elem()))
		return result
	case reflect.Pointer:
		if value.IsNil() {
			return value
		}
		result := reflect.New(value.Type().Elem())
		result.Elem().Set(stripSourceLines(value.Elem()))
		return result
	case reflect.Struct:
		result := reflect.New(value.Type()).Elem()
		result.Set(value)
		for i := 0; i < value.NumField(); i++ {
			if result.Field(i).CanSet() {
				result.Field(i).Set(stripSourceLines(value.Field(i)))
			}
		}
		return result
	case reflect.Slice:
		if value.IsNil() {
			return value
		}
		result := reflect.MakeSlice(value.Type(), 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			if _, ok := value.Index(i).Interface().(ir.SourceLine); ok {
				continue
			}
			result = reflect.Append(result, stripSourceLines(value.Index(i)))
		}
		return result
	default:
		return value
	}
}

func debugPendingProgram(test *testing.T, path string) bool {
	test.Helper()
	_, statError := os.Stat(strings.TrimSuffix(path, filepath.Ext(path)) + ".pending")
	if statError == nil {
		return true
	}
	if !errors.Is(statError, os.ErrNotExist) {
		test.Fatal(statError)
	}
	return false
}

func debugRefusedProgram(path string) bool {
	directory := filepath.Dir(path)
	return filepath.Base(directory) == "refused" && filepath.Base(filepath.Dir(directory)) == "review"
}

func reviewPrograms(test *testing.T) []string {
	test.Helper()
	paths := []string{}
	walkError := filepath.WalkDir("../oracle/testdata/review", func(path string, d fs.DirEntry, pathError error) error {
		if pathError == nil && !d.IsDir() && (strings.HasSuffix(path, ".a") || strings.HasSuffix(path, ".ts")) {
			if !debugPendingProgram(test, path) {
				paths = append(paths, path)
			}
		}
		return pathError
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	return paths
}

func debugCorpus(test *testing.T) []string {
	test.Helper()
	paths := reviewPrograms(test)
	graph, globError := filepath.Glob("../oracle/testdata/graph_regions*.a")
	if globError != nil {
		test.Fatal(globError)
	}
	for _, path := range graph {
		if !debugPendingProgram(test, path) {
			paths = append(paths, path)
		}
	}
	return paths
}

func runDebugStrippedIR(test *testing.T, shard int) {
	classAllocators := 0
	for _, path := range debugFamilyPaths(debugCorpus(test), shard, true) {
		test.Run(path, func(test *testing.T) {
			release, debugProgramError := debugProgram(test, path, false)
			debug, debugError := debugProgram(test, path, true)
			if debugRefusedProgram(path) {
				if debugProgramError == nil || debugError == nil {
					test.Fatal("review/refused program unexpectedly lowered")
				}
				requireDebugDiagnostic(test, debugProgramError)
				requireDebugDiagnostic(test, debugError)
				return
			}
			if debugProgramError != nil {
				test.Fatalf("agree/smoke program failed to lower: %v", debugProgramError)
			}
			if debugError != nil {
				test.Fatal(debugError)
			}
			// Module boundaries count all physical statements. Translate the counts before stripping.
			offset := len(debug.Main)
			for _, module := range debug.MainModules {
				offset -= module.Statements
			}
			for i := range debug.MainModules {
				end := offset + debug.MainModules[i].Statements
				for _, statement := range debug.Main[offset:end] {
					if _, ok := statement.(ir.SourceLine); ok {
						debug.MainModules[i].Statements--
					}
				}
				offset = end
			}
			releasePlan, debugPlan := planRegions(release), planRegions(debug)
			classAllocators += len(releasePlan.classObjects)
			if !reflect.DeepEqual(releasePlan.classObjects, debugPlan.classObjects) ||
				!reflect.DeepEqual(releasePlan.fresh, debugPlan.fresh) {
				test.Fatal("debug metadata changed region eligibility")
			}
			stripped := stripSourceLines(reflect.ValueOf(debug)).Interface().(*ir.Program)
			if !equalSourceIR(reflect.ValueOf(release), reflect.ValueOf(stripped)) {
				a, _ := json.Marshal(release)
				b, _ := json.Marshal(stripped)
				for i := 0; i < len(a) && i < len(b); i++ {
					if a[i] != b[i] {
						end := i + 180
						if end > len(a) {
							end = len(a)
						}
						end2 := i + 180
						if end2 > len(b) {
							end2 = len(b)
						}
						test.Fatalf("IR differs at JSON byte %d: release %s debug %s", i, a[i:end], b[i:end2])
					}
				}
				test.Fatalf("IR differs: JSON lengths %d/%d equal=%t", len(a), len(b), bytes.Equal(a, b))
			}
		})
	}
	if shard == -1 && classAllocators == 0 {
		test.Fatal("corpus contains no region-eligible class allocator")
	}
	test.Logf("stripped IR equal; %d region-eligible class allocator observations", classAllocators)
}

func TestDebugTwoDirectories(test *testing.T) {
	test.Parallel()
	for _, mode := range []struct {
		name    string
		options Options
	}{
		{"release", Options{Release: true}}, {"sanitized", Options{Sanitize: true}}, {"debug", Options{Debug: true}},
	} {
		test.Run(mode.name, func(test *testing.T) {
			var firstBinary []byte
			var first string
			var firstKeys []string
			for i := 0; i < 2; i++ {
				path := filepath.Join(test.TempDir(), "fixture.a")
				if createDirectoryError := os.Mkdir(
					filepath.Join(filepath.Dir(path), "lib"),
					0755,
				); createDirectoryError != nil {
					test.Fatal(createDirectoryError)
				}
				if writeError := os.WriteFile(
					filepath.Join(filepath.Dir(path), "lib", "value.a"),
					[]byte("export function f(n: number): number { return n + 1; }\n"),
					0644,
				); writeError != nil {
					test.Fatal(writeError)
				}
				if writeError := os.WriteFile(
					path,
					[]byte("import { f } from './lib/value.a';\nconsole.log((f(4)).toString());\n"),
					0644,
				); writeError != nil {
					test.Fatal(writeError)
				}
				options := mode.options
				options.DebugDirectory = filepath.Dir(path)
				program, debugProgramError := debugProgram(test, path, options.Debug)
				if debugProgramError != nil {
					test.Fatal(debugProgramError)
				}
				source := CWithOptions(program, options)
				binary := filepath.Join(filepath.Dir(path), "program")
				if buildWithRuntimeError := BuildWithRuntime(
					source,
					binary,
					options,
					debugArchiveProduct(test, options, false),
				); buildWithRuntimeError != nil {
					test.Fatal(buildWithRuntimeError)
				}
				if options.Debug && !strings.Contains(source, `#line 1 "lib/value.a"`) {
					test.Fatal("missing entry-relative imported marker")
				}
				if options.Debug && !strings.Contains(source, `#line 2 "fixture.a"`) {
					test.Fatal("missing relative entry marker")
				}
				contents, readError := os.ReadFile(binary)
				if readError != nil {
					test.Fatal(readError)
				}
				if i == 0 {
					firstBinary = contents
				} else if options.Debug {
					if bytes.Equal(firstBinary, contents) {
						test.Fatal("debug binaries must differ by entry directory")
					}
				} else if !bytes.Equal(firstBinary, contents) {
					test.Fatal("release/sanitized binaries must match")
				}
				if i == 0 {
					first = source
				} else if source != first {
					test.Fatal("C differs between directories")
				}
				// Splitting is deterministic too, including each unit's resolved reset.
				header, units, splitCError := splitC(source)
				if splitCError != nil {
					test.Fatal(splitCError)
				}
				header2, units2, splitCError := splitC(first)
				if splitCError != nil {
					test.Fatal(splitCError)
				}
				if header != header2 || !reflect.DeepEqual(units, units2) {
					test.Fatal("split unit bytes differ")
				}
				common, readRuntimeError := readRuntime(runtime, "runtime")
				if readRuntimeError != nil {
					test.Fatal(readRuntimeError)
				}
				common = append(common, runtimeFile{name: "units.h", contents: []byte(header)})
				keying, newUnitKeyingError := newUnitKeying(
					unitFirst,
					common,
					append(SourceFlags(source, options), programDebugFlags(options)...),
					unitKeyCompiler(test),
					unitKeyEnvironment(),
					options.Sanitize || options.ThreadSanitize,
				)
				if newUnitKeyingError != nil {
					test.Fatal(newUnitKeyingError)
				}
				var keys []string
				for _, unit := range units {
					keys = append(keys, keying.key(unit))
				}
				if i == 0 {
					firstKeys = keys
				} else if options.Debug {
					if reflect.DeepEqual(firstKeys, keys) {
						test.Fatal("debug keys must differ by entry directory")
					}
				} else if !reflect.DeepEqual(firstKeys, keys) {
					test.Fatal("release/sanitized keys must match")
				}

			}
			// Debug binaries and unit keys embed the absolute entry directory, as Go
			// does without -trimpath. C bytes remain identical in every mode.
			if mode.options.Debug {
				test.Log("C identical; debug binaries and keys differ because they embed the absolute entry directory")
			} else {
				test.Log("C, binaries and keys identical")
			}
		})
	}
}

func TestDebugCounterType(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "counter.a")
	if writeError := os.WriteFile(path, []byte(
		"function sum(n: number): number { let total = 0; for (let i = 0; i < 10; i++) { total += i; } return total + n; }\nconsole.log((sum(2)).toString());\n",
	), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	for _, debug := range []bool{false, true} {
		program, debugProgramError := debugProgram(test, path, debug)
		if debugProgramError != nil {
			test.Fatal(debugProgramError)
		}
		found := false
		for i, local := range program.Locals {
			if local.Name == "i" {
				found = true
				if !local.Counter {
					test.Fatalf("debug=%t: counter lost integer proof", debug)
				}
				e := emitter{program: program}
				if !strings.Contains(CWithOptions(program, Options{Debug: debug}), "int64_t "+e.localName(i)) {
					test.Fatal("no emitted integer counter")
				}
			}
		}
		if !found {
			test.Fatal("missing counter")
		}
	}
}

// sourceLineIn follows nested statements inside expressions as well as blocks.
func sourceLineIn(value reflect.Value) bool {
	if !value.IsValid() {
		return false
	}
	if value.Type() == reflect.TypeFor[ir.SourceLine]() {
		return true
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		return !value.IsNil() && sourceLineIn(value.Elem())
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if sourceLineIn(value.Field(i)) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if sourceLineIn(value.Index(i)) {
				return true
			}
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			if sourceLineIn(key) || sourceLineIn(value.MapIndex(key)) {
				return true
			}
		}
	}
	return false
}

func debugByteIdentityPaths(test *testing.T) []string {
	test.Helper()
	paths := map[string]bool{}
	for _, path := range reviewPrograms(test) {
		paths[path] = true
	}
	smoke, readError := os.ReadFile("../../cloud/fast-gate/smoke.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	for _, line := range strings.Split(string(smoke), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && !strings.HasPrefix(fields[0], "#") {
			path := filepath.Join("../..", fields[len(fields)-1])
			if !debugPendingProgram(test, path) {
				paths[path] = true
			}
		}
	}
	ordered := make([]string, 0, len(paths))
	for path := range paths {
		ordered = append(ordered, path)
	}
	sort.Strings(ordered)
	return ordered
}

func requireDebugDiagnostic(test *testing.T, sourceError error) {
	test.Helper()
	var refused *lower.Refused
	var notYet *lower.NotYet
	if !errors.As(sourceError, &refused) && !errors.As(sourceError, &notYet) {
		test.Fatalf("non-compiling program must return Refused or NotYet, got %T: %v", sourceError, sourceError)
	}
	if strings.TrimSpace(sourceError.Error()) == "" {
		test.Fatal("non-compiling program has an empty diagnostic")
	}
	test.Logf("typed lowering diagnostic: %v", sourceError)
}

// Review's refused family proves its diagnostic claim. Every other program
// must lower and prove the C and IR claims; pending programs were not enumerated.
func checkDebugByteIdentity(test *testing.T, path string) bool {
	test.Helper()
	program, debugProgramError := debugProgram(test, path, false)
	if debugRefusedProgram(path) {
		if debugProgramError == nil {
			test.Fatal("review/refused program unexpectedly lowered")
		}
		requireDebugDiagnostic(test, debugProgramError)
		return false
	}
	if debugProgramError != nil {
		test.Fatalf("agree/smoke program failed to lower: %v", debugProgramError)
	}
	if sourceLineIn(reflect.ValueOf(program.Main)) {
		test.Fatal("non-debug main contains SourceLine")
	}
	for _, function := range program.Functions {
		if sourceLineIn(reflect.ValueOf(function.Body)) {
			test.Fatalf("non-debug function %s contains SourceLine", function.Name)
		}
	}
	plain, options := C(program), CWithOptions(program, Options{})
	if plain != options {
		test.Fatal("C and CWithOptions without Debug differ")
	}
	for _, source := range []string{plain, options} {
		if strings.Contains(source, "#line") || strings.Contains(source, "/* adamic source reset */") {
			test.Fatal("non-debug C contains source metadata")
		}
	}
	return true
}

func runDebugByteIdentity(test *testing.T, shard int) {
	compiled, refused := 0, 0
	for _, path := range debugFamilyPaths(debugByteIdentityPaths(test), shard, false) {
		test.Run(path, func(test *testing.T) {
			if checkDebugByteIdentity(test, path) {
				compiled++
			} else {
				refused++
			}
		})
	}
	test.Logf(
		"%d current review/smoke programs compiled: default C identical, no directives/reset tags, no SourceLine IR; %d returned typed lowering diagnostics",
		compiled,
		refused,
	)
}

func TestDebugLineTable(test *testing.T) {
	test.Parallel()
	binary := filepath.Join(debugLineProduct(test, false), "program")
	tool, pathError := toolchain.Path("llvm-dwarfdump")
	if pathError != nil {
		clang, clangError := toolchain.Path("clang")
		if clangError == nil {
			clang, clangError = filepath.EvalSymlinks(clang)
		}
		if clangError != nil {
			test.Fatal(clangError)
		}
		tool = filepath.Join(filepath.Dir(clang), "llvm-dwarfdump")
		if _, statError := os.Stat(tool); statError != nil {
			test.Fatal("toolchain has no llvm-dwarfdump: stop line-table proof")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	output, combinedOutputError := exec.CommandContext(ctx, tool, "--debug-line", binary).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("dwarfdump: %v: %s", combinedOutputError, output)
	}
	if !bytes.Contains(output, []byte("fixture.a")) {
		test.Fatal("dwarfdump has no fixture.a")
	}
	var dwarfData *dwarf.Data
	var dwarfError error
	if object, openError := elf.Open(binary); openError == nil {
		defer object.Close()
		dwarfData, dwarfError = object.DWARF()
	} else {
		object, openError := macho.Open(binary)
		if openError != nil {
			test.Fatal(openError)
		}
		defer object.Close()
		dwarfData, dwarfError = object.DWARF()
	}
	if dwarfError != nil {
		test.Fatal(dwarfError)
	}
	var symbols []struct {
		Name        string
		Value, Size uint64
	}
	functions := dwarfData.Reader()
	for {
		entry, nextError := functions.Next()
		if nextError != nil {
			test.Fatal(nextError)
		}
		if entry == nil {
			break
		}
		if entry.Tag != dwarf.TagSubprogram {
			continue
		}
		name, _ := entry.Val(dwarf.AttrName).(string)
		ranges, rangesError := dwarfData.Ranges(entry)
		if rangesError != nil {
			test.Fatal(rangesError)
		}
		for _, span := range ranges {
			symbols = append(symbols, struct {
				Name        string
				Value, Size uint64
			}{name, span[0], span[1] - span[0]})
		}
	}
	marked, async, unmarked := false, false, false
	for _, symbol := range symbols {
		syncSubject := strings.Contains(symbol.Name, "marked") && strings.HasPrefix(symbol.Name, "adamic_function_")
		asyncSubject := strings.Contains(symbol.Name, "async_resume") && strings.Contains(symbol.Name, "later")
		unmarkedSubject := strings.Contains(symbol.Name, "async_children") && strings.Contains(symbol.Name, "after")
		if !syncSubject && !asyncSubject && !unmarkedSubject {
			continue
		}
		rows := 0
		right := map[int]bool{}
		reader := dwarfData.Reader()
		for {
			unit, nextError := reader.Next()
			if nextError != nil {
				test.Fatal(nextError)
			}
			if unit == nil {
				break
			}
			if unit.Tag != dwarf.TagCompileUnit {
				continue
			}
			lines, lineReaderError := dwarfData.LineReader(unit)
			if lineReaderError != nil {
				test.Fatal(lineReaderError)
			}
			if lines == nil {
				continue
			}
			var row dwarf.LineEntry
			for {
				nextError := lines.Next(&row)
				if nextError == io.EOF {
					break
				}
				if nextError != nil {
					test.Fatal(nextError)
				}
				if row.EndSequence || row.Address < symbol.Value || row.Address >= symbol.Value+symbol.Size {
					continue
				}
				rows++
				file := filepath.Base(row.File.Name)
				if syncSubject || asyncSubject {
					if file == "fixture.a" {
						right[row.Line] = true
					}
				} else if strings.HasSuffix(file, ".c") {
					right[row.Line] = true
				} else if !strings.HasSuffix(file, ".h") {
					test.Fatalf("unmarked async children helper maps to %s:%d", file, row.Line)
				}
			}
		}
		if rows == 0 {
			test.Fatalf("no rows for %s", symbol.Name)
		}
		if syncSubject {
			// Inlined formatting may use runtime rows; the branch and return retain source rows.
			if !right[3] || !right[4] {
				test.Fatalf("marked rows %v, want 3,4", right)
			}
			marked = true
		} else if asyncSubject {
			for _, line := range []int{7, 8, 10, 12} {
				if !right[line] {
					test.Fatalf("async rows %v, want fixture.a line %d before/after await and at throw", right, line)
				}
			}
			async = true
		} else {
			if len(right) == 0 {
				test.Fatal("following async children helper has no C rows")
			}
			unmarked = true
		}
	}
	if !marked || !async || !unmarked {
		test.Fatalf("subjects: marked=%t async=%t unmarked=%t", marked, async, unmarked)
	}
	test.Log(
		"llvm-dwarfdump: sync fixture.a lines 3,4; async fixture.a lines 7,8,10,12; following adamic_async_children_after helper: C",
	)
}

func debugRunBinary(test *testing.T, binary string) mainPartsRun {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(command.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "LSAN_OPTIONS=exitcode=99")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runError := command.Run()
	if ctx.Err() != nil {
		test.Fatal(ctx.Err())
	}
	if command.ProcessState == nil {
		test.Fatal(runError)
	}
	return mainPartsRun{stdout.String(), stderr.String(), command.ProcessState.ExitCode()}
}

func runDebugReviewOutputAndOwnership(test *testing.T, shard int) {
	exercised := 0
	var paths []string
	for _, path := range reviewPrograms(test) {
		if filepath.Dir(path) == "../oracle/testdata/review/agree" {
			paths = append(paths, path)
		}
	}
	for _, path := range debugFamilyPaths(paths, shard, false) {
		test.Run(filepath.Base(path), func(test *testing.T) {

			want := debugRunBinary(test, filepath.Join(debugReviewProduct(test, path, 0, false), "program"))
			got := debugRunBinary(test, filepath.Join(debugReviewProduct(test, path, 1, false), "program"))
			if got != want {
				test.Fatalf("debug=%+v release=%+v", got, want)
			}
			sanitized := debugRunBinary(test, filepath.Join(debugReviewProduct(test, path, 2, false), "program"))
			if sanitized != want {
				test.Fatalf("debug ASan/UBSan/LSan=%+v release=%+v", sanitized, want)
			}
			exercised++
		})
	}
	test.Logf("%d agree programs exercised: identical output/status, ASan, UBSan and LeakSanitizer", exercised)
}

// DeepEqual considers NaN unequal to itself. IR constants compare by their IEEE bits.
func equalSourceIR(a, b reflect.Value) bool {
	if !a.IsValid() || !b.IsValid() {
		return a.IsValid() == b.IsValid()
	}
	if a.Type() != b.Type() {
		return false
	}
	switch a.Kind() {
	case reflect.Float64, reflect.Float32:
		return math.Float64bits(a.Float()) == math.Float64bits(b.Float())
	case reflect.Pointer, reflect.Interface:
		if a.IsNil() || b.IsNil() {
			return a.IsNil() == b.IsNil()
		}
		return equalSourceIR(a.Elem(), b.Elem())
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if !equalSourceIR(a.Field(i), b.Field(i)) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if a.IsNil() != b.IsNil() || a.Len() != b.Len() {
			return false
		}
		for i := 0; i < a.Len(); i++ {
			if !equalSourceIR(a.Index(i), b.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map:
		if a.IsNil() != b.IsNil() || a.Len() != b.Len() {
			return false
		}
		for _, key := range a.MapKeys() {
			if !equalSourceIR(a.MapIndex(key), b.MapIndex(key)) {
				return false
			}
		}
		return true
	case reflect.Bool:
		return a.Bool() == b.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() == b.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return a.Uint() == b.Uint()
	case reflect.String:
		return a.String() == b.String()
	default:
		return reflect.DeepEqual(a.Interface(), b.Interface())
	}
}

const debugLineFixture = `function marked(n: number): number {
 console.log(n.toString());
 if (n > 0) return marked(n - 1) + n;
 return n;
}
async function later(n: number): Promise<number> {
 const before = n + 2;
 const resumed = await Promise.resolve(before);
 if (resumed > 0) {
  return resumed + 1;
 }
 throw new Error('negative');
}
async function after(n: number): Promise<number> {
 return n;
}
console.log((marked(3)).toString());
console.log((await later(2)).toString());
try { await later(-3); } catch { console.log('caught'); }
console.log((await after(9)).toString());
`
