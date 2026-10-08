package yaml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func goFormat(t *testing.T, cases string) []byte {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs("testdata/format_go.go")
	if err != nil {
		t.Fatal(err)
	}
	goBinary := yamlProduct(t, yamlProductInputs{
		Name: "go-format", Files: []string{source, filepath.Join(root, "cohere")},
		Flags: []string{"build", "-overlay", "./command/formatter_comparison"}, Toolchain: "go",
	}, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): source}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		return yamlBuildCommand(filepath.Join(root, "cohere"), "go", "build", "-overlay", path, "-o", filepath.Join(dir, "go-format"), "./command/formatter_comparison")
	})
	goBinary = filepath.Join(goBinary, "go-format")
	expected := run(t, "", nil, goBinary, "--cases", cases)
	if artifacts := os.Getenv("ADAMIC_YAML_ARTIFACTS"); artifacts != "" {
		if err := os.MkdirAll(artifacts, 0755); err != nil {
			t.Fatal(err)
		}
		executable, err := os.ReadFile(goBinary)
		if err != nil {
			t.Fatal(err)
		}
		inputs, err := os.ReadFile(cases)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range []struct {
			name string
			data []byte
			mode os.FileMode
		}{{"go-format", executable, 0755}, {"format-cases.txt", inputs, 0644}, {"format-expected.txt", expected, 0644}} {
			if err := os.WriteFile(filepath.Join(artifacts, file.name), file.data, file.mode); err != nil {
				t.Fatal(err)
			}
		}
	}
	return expected
}

func formatCases(t *testing.T) (string, int, int) {
	path, files, count := composeCases(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var extra strings.Builder
	add := func(text string) { extra.WriteString("0\t" + escape.Replace(text) + "\n"); count++ }
	for _, unit := range []string{"a", "中", "😀", "é", "👨‍👩‍👦"} {
		for width := 34; width <= 86; width += 2 {
			value := strings.Repeat(unit, width)
			for _, text := range []string{"key: [" + value + ", tail]\n", "{first: " + value + ", second: [a, b]}\n", "[" + value + ", {x: y}, end]\n", "\"" + value + "\": [a, b]\n"} {
				add(text)
			}
		}
	}
	for _, header := range []string{"|", "|-", "|+", "|2", "|2-", "|+2", ">", ">-", ">+", ">2", ">2-", ">+2"} {
		for _, body := range []string{"", "\n", "  a\n", "  a\n\n", "  a\n\n\n", "  a\n    b\n  c\n", "  a  \n  b\t\n"} {
			for _, prefix := range []string{"", "key: ", "- "} {
				add(prefix + header + " # header\n" + body)
			}
		}
	}
	for _, text := range []string{"# prettier-ignore\nkey:   [a,b]\n", "---\n# prettier-ignore\na:    {b: c}\n", "a:\n  # prettier-ignore\n  -   {b:  c}\n", "? [a,b]\n: {c: d}\n", "key: # colon\n  value\n", "a: 1\n\n# end\n", "[a,\n\n# middle\nb]\n", "&a\n# middle\n{b: c}\n", "!local &a\n# middle\nvalue\n", "a: 'x\\y'\n", "a: \"x\\\"y\"\n", "a: 'x\ny'\n", "a: \"x\\n y\"\n", "\ufeffa:  1\r\n", "\ufeff  ", "  \t\n", "", "---\n...\n---\n"} {
		add(text)
	}
	target := filepath.Join(t.TempDir(), "format-cases.txt")
	if err := os.WriteFile(target, append(data, []byte(extra.String())...), 0644); err != nil {
		t.Fatal(err)
	}
	return target, files, count
}

const testFormatterMatchesGoShards = 20
const testFileDriverShards = 50
const testFormatterMutantsShards = 6

// TestFormatterMatchesGo runs deterministic ranges of 512 complete corpus
// cases in parallel. ADAMIC_TEST_SHARD=i/n (zero-based i) selects stable unit
// indices modulo n; unset runs all. All four comparison sides run on every unit.
func TestFormatterMatchesGo(t *testing.T) {
	cases, files, count := formatCases(t)
	expected := goFormat(t, cases)
	binary, emitted, entry, runner := yamlFormatterProducts(t)
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to yaml@2.9.0 and prettier@3.9.6")
	}
	inputs, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	inputLines := strings.Split(strings.TrimSuffix(string(inputs), "\n"), "\n")
	expectedLines := bytes.Split(bytes.TrimSuffix(expected, []byte("\n")), []byte("\n"))
	if len(inputLines) != count || len(expectedLines) != count {
		t.Fatalf("case enumeration=%d, inputs=%d, Go answers=%d", count, len(inputLines), len(expectedLines))
	}
	units, err := yamlCaseUnits(count, yamlFormatterCaseUnitSize)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != testFormatterMatchesGoShards {
		t.Fatalf("enumerated %d shards, declared testFormatterMatchesGoShards=%d", len(units), testFormatterMatchesGoShards)
	}
	selected := yamlSelectedUnits(t, len(units))
	binaryInputs, prototypeInputs := yamlFormatterLibraryExceptions()
	totalKnown := 0
	for _, input := range inputLines {
		text := unescapeCase(input)
		if binaryInputs[text] || prototypeInputs[text] {
			totalKnown++
		}
	}
	if totalKnown != 42 {
		t.Fatalf("expected exactly 42 independently proved library difference cases, enumerated %d", totalKnown)
	}
	t.Logf("union: %d unique case ids across %d units; %d repository files; 42 proved library difference cases", count, len(units), files)
	var prepared []struct {
		unit   yamlCaseUnit
		path   string
		wanted []byte
		known  int
	}
	for index, unit := range units {
		if !selected[index] {
			continue
		}
		shardInputs := inputLines[unit.start:unit.end]
		path := filepath.Join(t.TempDir(), "cases.txt")
		if err := os.WriteFile(path, []byte(strings.Join(shardInputs, "\n")+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		wanted := append(bytes.Join(expectedLines[unit.start:unit.end], []byte("\n")), '\n')
		known := 0
		for _, input := range shardInputs {
			text := unescapeCase(input)
			if binaryInputs[text] || prototypeInputs[text] {
				known++
			}
		}
		prepared = append(prepared, struct {
			unit   yamlCaseUnit
			path   string
			wanted []byte
			known  int
		}{unit, path, wanted, known})
	}
	for _, product := range prepared {
		unit, path, wanted, known := product.unit, product.path, product.wanted, product.known
		t.Run(unit.name, func(t *testing.T) {
			t.Parallel()
			defer yamlUnitClock(t)()
			t.Logf("content: case ids %d..%d (%d cases)", unit.start, unit.end-1, unit.end-unit.start)
			for _, side := range []struct {
				name string
				out  []byte
			}{
				{"native ASan/UBSan/LSan", run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, "--cases", path)},
				{"Node", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, "--cases", path)},
				{"emitted JavaScript", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, "--cases", path)},
			} {
				if err := yamlCompareCaseOutputs(unit.name, side.name, side.out, wanted); err != nil {
					t.Fatal(err)
				}
			}
			external := run(t, "", nil, "node", "testdata/format_library.mjs", library, "--cases", path)
			yamlCheckFormatterLibrary(t, wanted, external, inputLines[unit.start:unit.end], unit.start, known)
		})
	}
}

func yamlFormatterLibraryExceptions() (map[string]bool, map[string]bool) {
	binaryInputs := map[string]bool{}
	prototypeInputs := map[string]bool{}
	for _, version := range []string{"", "%YAML 1.1\n---\n", "%YAML 1.2\n---\n"} {
		for _, value := range []string{"2000-01-01", "<<"} {
			binaryInputs[version+"!!binary "+value+"\n"] = true
		}
		for _, tag := range []string{"!<toString>", "!<constructor>", "!<__proto__>"} {
			for _, value := range []string{"word", "2000-01-01", "null", "<<"} {
				prototypeInputs[version+tag+" "+value+"\n"] = true
			}
		}
	}
	return binaryInputs, prototypeInputs
}

func yamlCheckFormatterLibrary(t *testing.T, expected, external []byte, inputLines []string, offset, wantDifferences int) {
	t.Helper()
	expectedLines := bytes.Split(bytes.TrimSuffix(expected, []byte("\n")), []byte("\n"))
	actualLines := bytes.Split(bytes.TrimSuffix(external, []byte("\n")), []byte("\n"))
	if len(actualLines) != len(expectedLines) {
		t.Fatalf("library answered %d cases, Go %d", len(actualLines), len(expectedLines))
	}
	binaryInputs, prototypeInputs := yamlFormatterLibraryExceptions()
	differences := 0
	for localIndex, actual := range actualLines {
		index := offset + localIndex
		wanted := expectedLines[localIndex]
		if bytes.Equal(actual, wanted) {
			continue
		}
		input := unescapeCase(inputLines[localIndex])
		knownBinary := binaryInputs[input]
		knownMinification := prototypeInputs[input] && string(actual) == "error\tc.resolve is not a function" && string(wanted) == "error\ttag.resolve is not a function"
		if !(knownBinary && string(actual) == "error\tInvalid character" && bytes.HasPrefix(wanted, []byte("ok\t"))) && !knownMinification {
			os.WriteFile("/tmp/stage1-yaml-format-library.txt", external, 0644)
			os.WriteFile("/tmp/stage1-yaml-format-expected.txt", expected, 0644)
			t.Fatalf("new library difference, case %d input %q: %s", index, input, firstDifference(actual, wanted))
		}
		differences++
		t.Logf("recorded original Prettier difference, case %d: %q versus Go %q", index, actual, wanted)
	}
	if differences != wantDifferences {
		t.Fatalf("expected exactly %d independently proved bundled-library differences, got %d", wantDifferences, differences)
	}
	t.Logf("original Prettier: %d matching cases, %d proved binary/minification differences", len(inputLines)-differences, differences)
}

func unescapeCase(line string) string {
	parts := strings.SplitN(line, "\t", 2)
	return strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t").Replace(parts[1])
}

func TestBundledParserDifference(t *testing.T) {
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	actual := run(t, "", nil, "node", "gaps/bundledParser.mjs", library)
	expected := "yaml: accepted\nprettier: Invalid character\nyaml: accepted\nprettier: Invalid character\nyaml: tag.resolve is not a function\nprettier: c.resolve is not a function\n"
	if string(actual) != expected {
		t.Fatalf("original library gap changed: %q", actual)
	}
	t.Logf("%s", actual)
}

// TestFileDriver runs one parallel unit per corpus file/control, preserving all
// native sanitizer/leak and both JavaScript stdout comparisons on every unit.
// ADAMIC_TEST_SHARD=i/n (zero-based i) selects stable indices modulo n; unset runs all.
func TestFileDriver(t *testing.T) {
	cases, files, _ := formatCases(t)
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")[:files]
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	for _, text := range []string{"", "  \t\n", "\ufeff", "\ufeffa: b\n", "|+", "|+\n", "|+ # header\n", ">+", "a: |+", "a: |+\n", "a: |+\n  b\n\n", "{a: b, c: [x,y]}\n", "---\n...\n", "key: 'x\\y'\n"} {
		inputs = append(inputs, "0\t"+escape.Replace(text))
	}
	path := filepath.Join(t.TempDir(), "driver-cases.txt")
	if err := os.WriteFile(path, []byte(strings.Join(inputs, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	expected := bytes.Split(bytes.TrimSuffix(goFormat(t, path), []byte("\n")), []byte("\n"))
	binary, emitted, entry, runner := yamlFormatterProducts(t)
	units, err := yamlCaseUnits(len(inputs), yamlFileDriverCaseUnitSize)
	if err != nil {
		t.Fatal(err)
	}
	if len(expected) != len(inputs) {
		t.Fatalf("Go answered %d driver cases, want %d", len(expected), len(inputs))
	}
	if len(units) != testFileDriverShards {
		t.Fatalf("enumerated %d shards, declared testFileDriverShards=%d", len(units), testFileDriverShards)
	}
	selected := yamlSelectedUnits(t, len(units))
	t.Logf("union: %d unique driver case ids across %d units (%d repository files, %d controls)", len(inputs), len(units), files, len(inputs)-files)
	for unitIndex, unit := range units {
		if !selected[unitIndex] {
			continue
		}
		index := unit.start
		input := inputs[index]
		t.Run(unit.name, func(t *testing.T) {
			t.Parallel()
			defer yamlUnitClock(t)()
			t.Logf("content: file/control case id %d", index)

			answer := string(expected[index])
			if !strings.HasPrefix(answer, "ok\t") {
				t.Fatalf("driver control %d invalid: %s", index, answer)
			}
			wanted := []byte(unescapeCase("0\t" + strings.TrimPrefix(answer, "ok\t")))
			file := filepath.Join(t.TempDir(), "input.yaml")
			if err := os.WriteFile(file, []byte(unescapeCase(input)), 0644); err != nil {
				t.Fatal(err)
			}
			for _, side := range []struct {
				name string
				out  []byte
			}{{"native", run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, file)}, {"Node", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, file)}, {"emitted JavaScript", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, file)}} {
				if err := yamlCompareCaseOutputs(unit.name, side.name, side.out, wanted); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

var formatterMutants = []struct{ name, file, from, to string }{
	{"root final newline lost", "printer.ts", "const hard = !(", "const hard = false && !("},
	{"colon separation lost", "printer.ts", "this.layout.text(': '), printedValue", "this.layout.text(':'), printedValue"},
	{"flow trailing comma lost", "printer.ts", "this.layout.ifBreak(this.layout.text(','), this.empty, -1)", "this.layout.ifBreak(this.layout.text(''), this.empty, -1)"},
	{"batch backslash scan misses escapes", "main.ts", "text.charCodeAt(index) !== 92", "text.charCodeAt(index) !== 13"},
	{"emoji first unit range loses endpoint", "width.ts", "code <= last", "code < last"},
	{"emoji surrogate slot shifted", "width.ts", "code - 0x100000 + 0xd800", "code - 0x100000 + 0xd900"},
}

// TestFormatterMutants runs one parallel unit per mutant. ADAMIC_TEST_SHARD=i/n
// (zero-based i) selects units whose stable index modulo n is i; unset runs all.
// Build products are prepared once before any parallel test logic starts.
func TestFormatterMutants(t *testing.T) {
	cases, _, count := formatCases(t)
	expected := goFormat(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	mutants := formatterMutants
	if len(mutants) != testFormatterMutantsShards {
		t.Fatalf("enumerated %d shards, declared testFormatterMutantsShards=%d", len(mutants), testFormatterMutantsShards)
	}
	// Enumerate the unsplit Cartesian product independently of the unit plan.
	var enumeration []string
	for _, mutant := range mutants {
		for index := 0; index < count; index++ {
			enumeration = append(enumeration, fmt.Sprintf("%s/case-%05d", mutant.name, index))
		}
	}
	units := make([]yamlUnit, len(mutants))
	for index, mutant := range mutants {
		units[index].name = yamlShardName(index, len(mutants))
		for caseIndex := 0; caseIndex < count; caseIndex++ {
			units[index].ids = append(units[index].ids, fmt.Sprintf("%s/case-%05d", mutant.name, caseIndex))
		}
	}
	if err := yamlValidateUnion(enumeration, units); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	if rows := len(strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")); rows != count {
		t.Fatalf("enumerated %d cases but input has %d rows", count, rows)
	}
	t.Logf("union: %d unique mutant/case ids, %d mutants x %d complete corpus cases", len(enumeration), len(mutants), count)
	selected := yamlSelectedUnits(t, len(units))
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := filepath.Glob("*.ts")
	if err != nil {
		t.Fatal(err)
	}
	var prepared []struct {
		unit                  yamlUnit
		mutant, entry, binary string
	}
	for index, mutant := range mutants {
		if !selected[index] {
			continue
		}
		unit := units[index]
		sources := yamlProduct(t, yamlProductInputs{
			Name: unit.name + "-sources", Files: entries, Flags: []string{mutant.file, mutant.from, mutant.to}, Toolchain: "source-overlay",
		}, func(dir string) error {
			for _, file := range entries {
				source, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				if file == mutant.file {
					if strings.Count(string(source), mutant.from) != 1 {
						return fmt.Errorf("mutation site must occur once")
					}
					source = []byte(strings.Replace(string(source), mutant.from, mutant.to, 1))
				}
				if err := os.WriteFile(filepath.Join(dir, file), source, 0644); err != nil {
					return err
				}
			}
			return nil
		})
		entry := filepath.Join(sources, "main.ts")
		lowered := yamlProduct(t, yamlProductInputs{
			Name: unit.name + "-lowered", Files: []string{sources, filepath.Join(root, "internal"), filepath.Join(root, "cohere"), filepath.Join(root, "go.mod")}, Flags: []string{"main.ts", "C"}, Toolchain: "adamic",
		}, func(dir string) error {
			program, err := load.Load([]string{entry})
			if err != nil {
				return err
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(dir, "mutant.c"), []byte(native.C(lowered)), 0644)
		})
		options := native.Options{}
		product := yamlProduct(t, yamlProductInputs{
			Name: unit.name + "-native", Files: []string{filepath.Join(lowered, "mutant.c"), filepath.Join(root, "internal/native")}, Flags: native.Flags(options), Toolchain: "clang",
		}, func(dir string) error {
			code, err := os.ReadFile(filepath.Join(lowered, "mutant.c"))
			if err != nil {
				return err
			}
			return native.Build(string(code), filepath.Join(dir, "mutant"), options)
		})
		binary := filepath.Join(product, "mutant")
		prepared = append(prepared, struct {
			unit                  yamlUnit
			mutant, entry, binary string
		}{unit, mutant.name, entry, binary})
	}
	for _, product := range prepared {
		unit, entry, binary := product.unit, product.entry, product.binary
		t.Run(unit.name, func(t *testing.T) {
			t.Parallel()
			defer yamlUnitClock(t)()
			t.Logf("content: mutant %q, case ids 0..%d (%d cases)", product.mutant, count-1, count)
			nativeOut := run(t, "", nil, binary, "--cases", cases)
			node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, "--cases", cases)
			for _, side := range []struct {
				name string
				out  []byte
			}{{"native", nativeOut}, {"Node", node}} {
				if err := yamlRejectSurvivor(unit.name, side.name, side.out, expected); err != nil {
					t.Fatal(err)
				}
				t.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(side.out, expected))
			}
		})
	}
}
