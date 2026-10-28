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

func debugProgram(t *testing.T, path string, debug bool) (*ir.Program, error) {
	t.Helper()
	loaded, err := load.Load([]string{path})
	if err != nil {
		return nil, err
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

func debugPendingProgram(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(strings.TrimSuffix(path, filepath.Ext(path)) + ".pending")
	if err == nil {
		return true
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	return false
}

func debugRefusedProgram(path string) bool {
	directory := filepath.Dir(path)
	return filepath.Base(directory) == "refused" && filepath.Base(filepath.Dir(directory)) == "review"
}

func reviewPrograms(t *testing.T) []string {
	t.Helper()
	paths := []string{}
	err := filepath.WalkDir("../oracle/testdata/review", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && (strings.HasSuffix(path, ".a") || strings.HasSuffix(path, ".ts")) {
			if !debugPendingProgram(t, path) {
				paths = append(paths, path)
			}
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return paths
}

func debugCorpus(t *testing.T) []string {
	t.Helper()
	paths := reviewPrograms(t)
	graph, err := filepath.Glob("../oracle/testdata/graph_regions*.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range graph {
		if !debugPendingProgram(t, path) {
			paths = append(paths, path)
		}
	}
	return paths
}

func runDebugStrippedIR(t *testing.T, shard int) {
	classAllocators := 0
	for _, path := range debugFamilyPaths(debugCorpus(t), shard, true) {
		t.Run(path, func(t *testing.T) {
			release, err := debugProgram(t, path, false)
			debug, debugErr := debugProgram(t, path, true)
			if debugRefusedProgram(path) {
				if err == nil || debugErr == nil {
					t.Fatal("review/refused program unexpectedly lowered")
				}
				requireDebugDiagnostic(t, err)
				requireDebugDiagnostic(t, debugErr)
				return
			}
			if err != nil {
				t.Fatalf("agree/smoke program failed to lower: %v", err)
			}
			if debugErr != nil {
				t.Fatal(debugErr)
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
			if !reflect.DeepEqual(releasePlan.classObjects, debugPlan.classObjects) || !reflect.DeepEqual(releasePlan.fresh, debugPlan.fresh) {
				t.Fatal("debug metadata changed region eligibility")
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
						t.Fatalf("IR differs at JSON byte %d: release %s debug %s", i, a[i:end], b[i:end2])
					}
				}
				t.Fatalf("IR differs: JSON lengths %d/%d equal=%t", len(a), len(b), bytes.Equal(a, b))
			}
		})
	}
	if shard == -1 && classAllocators == 0 {
		t.Fatal("corpus contains no region-eligible class allocator")
	}
	t.Logf("stripped IR equal; %d region-eligible class allocator observations", classAllocators)
}

func TestDebugTwoDirectories(t *testing.T) {
	t.Parallel()
	for _, mode := range []struct {
		name    string
		options Options
	}{
		{"release", Options{Release: true}}, {"sanitized", Options{Sanitize: true}}, {"debug", Options{Debug: true}},
	} {
		t.Run(mode.name, func(t *testing.T) {
			var firstBinary []byte
			var first string
			var firstKeys []string
			for i := 0; i < 2; i++ {
				path := filepath.Join(t.TempDir(), "fixture.a")
				if err := os.Mkdir(filepath.Join(filepath.Dir(path), "lib"), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(filepath.Dir(path), "lib", "value.a"), []byte("export function f(n: number): number { return n + 1; }\n"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("import { f } from './lib/value.a';\nconsole.log((f(4)).toString());\n"), 0644); err != nil {
					t.Fatal(err)
				}
				options := mode.options
				options.DebugDirectory = filepath.Dir(path)
				program, err := debugProgram(t, path, options.Debug)
				if err != nil {
					t.Fatal(err)
				}
				source := CWithOptions(program, options)
				binary := filepath.Join(filepath.Dir(path), "program")
				if err := Build(source, binary, options); err != nil {
					t.Fatal(err)
				}
				if options.Debug && !strings.Contains(source, `#line 1 "lib/value.a"`) {
					t.Fatal("missing entry-relative imported marker")
				}
				if options.Debug && !strings.Contains(source, `#line 2 "fixture.a"`) {
					t.Fatal("missing relative entry marker")
				}
				contents, err := os.ReadFile(binary)
				if err != nil {
					t.Fatal(err)
				}
				if i == 0 {
					firstBinary = contents
				} else if options.Debug {
					if bytes.Equal(firstBinary, contents) {
						t.Fatal("debug binaries must differ by entry directory")
					}
				} else if !bytes.Equal(firstBinary, contents) {
					t.Fatal("release/sanitized binaries must match")
				}
				if i == 0 {
					first = source
				} else if source != first {
					t.Fatal("C differs between directories")
				}
				// Splitting is deterministic too, including each unit's resolved reset.
				header, units, err := splitC(source)
				if err != nil {
					t.Fatal(err)
				}
				header2, units2, err := splitC(first)
				if err != nil {
					t.Fatal(err)
				}
				if header != header2 || !reflect.DeepEqual(units, units2) {
					t.Fatal("split unit bytes differ")
				}
				common, err := readRuntime(runtime, "runtime")
				if err != nil {
					t.Fatal(err)
				}
				common = append(common, runtimeFile{name: "units.h", contents: []byte(header)})
				keying := testUnitKeying(t, unitFirst, common, append(SourceFlags(source, options), programDebugFlags(options)...), unitKeyCompiler(t), unitKeyEnvironment())
				var keys []string
				for _, unit := range units {
					keys = append(keys, keying.key(unit))
				}
				if i == 0 {
					firstKeys = keys
				} else if options.Debug {
					if reflect.DeepEqual(firstKeys, keys) {
						t.Fatal("debug keys must differ by entry directory")
					}
				} else if !reflect.DeepEqual(firstKeys, keys) {
					t.Fatal("release/sanitized keys must match")
				}

			}
			// Debug binaries and unit keys embed the absolute entry directory, as Go
			// does without -trimpath. C bytes remain identical in every mode.
			if mode.options.Debug {
				t.Log("C identical; debug binaries and keys differ because they embed the absolute entry directory")
			} else {
				t.Log("C, binaries and keys identical")
			}
		})
	}
}

func TestDebugCounterType(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "counter.a")
	if err := os.WriteFile(path, []byte("function sum(n: number): number { let total = 0; for (let i = 0; i < 10; i++) { total += i; } return total + n; }\nconsole.log((sum(2)).toString());\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, debug := range []bool{false, true} {
		program, err := debugProgram(t, path, debug)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for i, local := range program.Locals {
			if local.Name == "i" {
				found = true
				if !local.Counter {
					t.Fatalf("debug=%t: counter lost integer proof", debug)
				}
				e := emitter{program: program}
				if !strings.Contains(CWithOptions(program, Options{Debug: debug}), "int64_t "+e.localName(i)) {
					t.Fatal("no emitted integer counter")
				}
			}
		}
		if !found {
			t.Fatal("missing counter")
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

func debugByteIdentityPaths(t *testing.T) []string {
	t.Helper()
	paths := map[string]bool{}
	for _, path := range reviewPrograms(t) {
		paths[path] = true
	}
	smoke, err := os.ReadFile("../../cloud/fast-gate/smoke.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(smoke), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && !strings.HasPrefix(fields[0], "#") {
			path := filepath.Join("../..", fields[len(fields)-1])
			if !debugPendingProgram(t, path) {
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

func requireDebugDiagnostic(t *testing.T, err error) {
	t.Helper()
	var refused *lower.Refused
	var notYet *lower.NotYet
	if !errors.As(err, &refused) && !errors.As(err, &notYet) {
		t.Fatalf("non-compiling program must return Refused or NotYet, got %T: %v", err, err)
	}
	if strings.TrimSpace(err.Error()) == "" {
		t.Fatal("non-compiling program has an empty diagnostic")
	}
	t.Logf("typed lowering diagnostic: %v", err)
}

// Review's refused family proves its diagnostic claim. Every other program
// must lower and prove the C and IR claims; pending programs were not enumerated.
func checkDebugByteIdentity(t *testing.T, path string) bool {
	t.Helper()
	program, err := debugProgram(t, path, false)
	if debugRefusedProgram(path) {
		if err == nil {
			t.Fatal("review/refused program unexpectedly lowered")
		}
		requireDebugDiagnostic(t, err)
		return false
	}
	if err != nil {
		t.Fatalf("agree/smoke program failed to lower: %v", err)
	}
	if sourceLineIn(reflect.ValueOf(program.Main)) {
		t.Fatal("non-debug main contains SourceLine")
	}
	for _, function := range program.Functions {
		if sourceLineIn(reflect.ValueOf(function.Body)) {
			t.Fatalf("non-debug function %s contains SourceLine", function.Name)
		}
	}
	plain, options := C(program), CWithOptions(program, Options{})
	if plain != options {
		t.Fatal("C and CWithOptions without Debug differ")
	}
	for _, source := range []string{plain, options} {
		if strings.Contains(source, "#line") || strings.Contains(source, "/* adamic source reset */") {
			t.Fatal("non-debug C contains source metadata")
		}
	}
	return true
}

func runDebugByteIdentity(t *testing.T, shard int) {
	compiled, refused := 0, 0
	for _, path := range debugFamilyPaths(debugByteIdentityPaths(t), shard, false) {
		t.Run(path, func(t *testing.T) {
			if checkDebugByteIdentity(t, path) {
				compiled++
			} else {
				refused++
			}
		})
	}
	t.Logf("%d current review/smoke programs compiled: default C identical, no directives/reset tags, no SourceLine IR; %d returned typed lowering diagnostics", compiled, refused)
}

func TestDebugLineTable(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(debugLineProduct(t, false), "program")
	tool, err := toolchain.Path("llvm-dwarfdump")
	if err != nil {
		clang, clangErr := toolchain.Path("clang")
		if clangErr == nil {
			clang, clangErr = filepath.EvalSymlinks(clang)
		}
		if clangErr != nil {
			t.Fatal(clangErr)
		}
		tool = filepath.Join(filepath.Dir(clang), "llvm-dwarfdump")
		if _, statErr := os.Stat(tool); statErr != nil {
			t.Fatal("toolchain has no llvm-dwarfdump: stop line-table proof")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, tool, "--debug-line", binary).CombinedOutput()
	if err != nil {
		t.Fatalf("dwarfdump: %v: %s", err, output)
	}
	if !bytes.Contains(output, []byte("fixture.a")) {
		t.Fatal("dwarfdump has no fixture.a")
	}
	var dwarfData *dwarf.Data
	if object, openErr := elf.Open(binary); openErr == nil {
		defer object.Close()
		dwarfData, err = object.DWARF()
	} else {
		object, openErr := macho.Open(binary)
		if openErr != nil {
			t.Fatal(openErr)
		}
		defer object.Close()
		dwarfData, err = object.DWARF()
	}
	if err != nil {
		t.Fatal(err)
	}
	var symbols []struct {
		Name        string
		Value, Size uint64
	}
	functions := dwarfData.Reader()
	for {
		entry, err := functions.Next()
		if err != nil {
			t.Fatal(err)
		}
		if entry == nil {
			break
		}
		if entry.Tag != dwarf.TagSubprogram {
			continue
		}
		name, _ := entry.Val(dwarf.AttrName).(string)
		ranges, err := dwarfData.Ranges(entry)
		if err != nil {
			t.Fatal(err)
		}
		for _, span := range ranges {
			symbols = append(symbols, struct {
				Name        string
				Value, Size uint64
			}{name, span[0], span[1] - span[0]})
		}
	}
	marked, async := false, false
	for _, symbol := range symbols {
		syncSubject := strings.Contains(symbol.Name, "marked") && strings.HasPrefix(symbol.Name, "adamic_function_")
		asyncSubject := strings.Contains(symbol.Name, "async_resume") && strings.Contains(symbol.Name, "later")
		if !syncSubject && !asyncSubject {
			continue
		}
		rows := 0
		right := map[int]bool{}
		reader := dwarfData.Reader()
		for {
			unit, err := reader.Next()
			if err != nil {
				t.Fatal(err)
			}
			if unit == nil {
				break
			}
			if unit.Tag != dwarf.TagCompileUnit {
				continue
			}
			lines, err := dwarfData.LineReader(unit)
			if err != nil {
				t.Fatal(err)
			}
			if lines == nil {
				continue
			}
			var row dwarf.LineEntry
			for {
				err := lines.Next(&row)
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				if row.EndSequence || row.Address < symbol.Value || row.Address >= symbol.Value+symbol.Size {
					continue
				}
				rows++
				file := filepath.Base(row.File.Name)
				if syncSubject {
					if file == "fixture.a" {
						right[row.Line] = true
					}
				} else if strings.HasSuffix(file, ".c") {
					right[row.Line] = true
				} else if !strings.HasSuffix(file, ".h") {
					t.Fatalf("async frame maps to %s:%d", file, row.Line)
				}
			}
		}
		if rows == 0 {
			t.Fatalf("no rows for %s", symbol.Name)
		}
		if syncSubject {
			// Inlined formatting may use runtime rows; the branch and return retain source rows.
			if !right[3] || !right[4] {
				t.Fatalf("marked rows %v, want 3,4", right)
			}
			marked = true
		} else {
			if len(right) == 0 {
				t.Fatal("async function has no C rows")
			}
			async = true
		}
	}
	if !marked || !async {
		t.Fatalf("subjects: marked=%t async=%t", marked, async)
	}
	t.Log("llvm-dwarfdump: build --debug marked fixture.a lines 3,4; later adamic_async_resume frame: .c")
}

func debugRunBinary(t *testing.T, binary string) mainPartsRun {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(command.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "LSAN_OPTIONS=exitcode=99")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	if command.ProcessState == nil {
		t.Fatal(err)
	}
	return mainPartsRun{stdout.String(), stderr.String(), command.ProcessState.ExitCode()}
}

func runDebugReviewOutputAndOwnership(t *testing.T, shard int) {
	exercised := 0
	var paths []string
	for _, path := range reviewPrograms(t) {
		if filepath.Dir(path) == "../oracle/testdata/review/agree" {
			paths = append(paths, path)
		}
	}
	for _, path := range debugFamilyPaths(paths, shard, false) {
		t.Run(filepath.Base(path), func(t *testing.T) {

			want := debugRunBinary(t, filepath.Join(debugReviewProduct(t, path, 0, false), "program"))
			got := debugRunBinary(t, filepath.Join(debugReviewProduct(t, path, 1, false), "program"))
			if got != want {
				t.Fatalf("debug=%+v release=%+v", got, want)
			}
			sanitized := debugRunBinary(t, filepath.Join(debugReviewProduct(t, path, 2, false), "program"))
			if sanitized != want {
				t.Fatalf("debug ASan/UBSan/LSan=%+v release=%+v", sanitized, want)
			}
			exercised++
		})
	}
	t.Logf("%d agree programs exercised: identical output/status, ASan, UBSan and LeakSanitizer", exercised)
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
 if (n > 0) return (await later(n - 1)) + n;
 return n + 2;
}
console.log((marked(3)).toString());
console.log((await later(2)).toString());
`
