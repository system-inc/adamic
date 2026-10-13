package yaml

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func goFormat(t *testing.T, cases string) []byte {
	t.Helper()
	goBinary := goFormatOracle(t)
	expected := run(t, "", nil, goBinary, "--cases", cases)
	if artifacts := os.Getenv("ADAMIC_YAML_ARTIFACTS"); artifacts != "" {
		yamlArtifacts.Lock()
		defer yamlArtifacts.Unlock()
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

// goFormatOracle is formatter_comparison over testdata/format_go.go, a product built ahead (#5qykzj5).
func goFormatOracle(t *testing.T) string {
	t.Helper()
	return formatterComparison(t, "go-format", "format_go.go")
}

func TestProduct_YAMLFormatGo(t *testing.T) {
	t.Parallel()
	goFormatOracle(t)
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

// formatterMatchesGo compares the formatter's port build (port_products_test.go) with Go over formatCases, in four
// shards of about ten seconds each alone on one core.
var formatterMatchesGo = yamlMatchGo{family: "format", shards: 4, cases: formatCases, oracle: goFormatOracle, arguments: []string{"--cases"}, build: yamlFormatBuild, entry: "main.ts"}

func TestFormatterMatchesGo_000(t *testing.T) { t.Parallel(); formatterMatchesGoShard(t, 0) }
func TestFormatterMatchesGo_001(t *testing.T) { t.Parallel(); formatterMatchesGoShard(t, 1) }
func TestFormatterMatchesGo_002(t *testing.T) { t.Parallel(); formatterMatchesGoShard(t, 2) }
func TestFormatterMatchesGo_003(t *testing.T) { t.Parallel(); formatterMatchesGoShard(t, 3) }

// TestFormatterMatchesGoUnion holds every case to one shard, answers the whole corpus with Go once (leaving it in
// ADAMIC_YAML_ARTIFACTS, when set, as the whole test did), and counts the 42 cases whose bundled-library differences
// the shards prove between them.
func TestFormatterMatchesGoUnion(t *testing.T) {
	t.Parallel()
	cases := formatterMatchesGo.union(t)
	path := filepath.Join(t.TempDir(), "format-cases.txt")
	if err := os.WriteFile(path, []byte(strings.Join(cases, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	expected := goFormat(t, path)
	if answers := len(bytes.Split(bytes.TrimSuffix(expected, []byte("\n")), []byte("\n"))); answers != len(cases) {
		t.Fatalf("Go answered %d of %d cases", answers, len(cases))
	}
	binaryInputs, prototypeInputs := formatterLibraryKnown()
	known := 0
	for _, line := range cases {
		input := unescapeCase(line)
		if binaryInputs[input] || prototypeInputs[input] {
			known++
		}
	}
	if known != 42 {
		t.Fatalf("expected exactly 42 cases with independently proved bundled-library differences, got %d", known)
	}
	t.Logf("%d cases, %d of them with proved bundled-library differences", len(cases), known)
}

func TestFormatterMatchesGoPlantedFailure(t *testing.T) {
	t.Parallel()
	formatterMatchesGo.planted(t)
}

// formatterLibraryKnown are the inputs whose original Prettier answers are proved to differ from Go's: binary values
// the bundled parser rejects, and prototype-named tags its minified build names c rather than tag.
func formatterLibraryKnown() (map[string]bool, map[string]bool) {
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

func formatterMatchesGoShard(t *testing.T, shard int) {
	t.Helper()
	result := formatterMatchesGo.run(t, shard)
	expected := result.expected
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", result.native}, {"Node", result.node}, {"emitted JavaScript", result.emitted}} {
		if !bytes.Equal(side.out, expected) {
			os.WriteFile(formatterMatchesGo.diagnostic(shard, "expected"), expected, 0644)
			os.WriteFile(formatterMatchesGo.diagnostic(shard, "actual"), side.out, 0644)
			t.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	external := run(t, "", nil, "node", "testdata/format_library.mjs", library, "--cases", result.path)
	expectedLines := bytes.Split(bytes.TrimSuffix(expected, []byte("\n")), []byte("\n"))
	actualLines := bytes.Split(bytes.TrimSuffix(external, []byte("\n")), []byte("\n"))
	if len(expectedLines) != len(result.indices) {
		t.Fatalf("Go answered %d of %d cases", len(expectedLines), len(result.indices))
	}
	if len(actualLines) != len(expectedLines) {
		t.Fatalf("library answered %d cases, Go %d", len(actualLines), len(expectedLines))
	}
	binaryInputs, prototypeInputs := formatterLibraryKnown()
	known := 0
	for _, index := range result.indices {
		input := unescapeCase(result.all[index])
		if binaryInputs[input] || prototypeInputs[input] {
			known++
		}
	}
	differences := 0
	for line, actual := range actualLines {
		wanted := expectedLines[line]
		if bytes.Equal(actual, wanted) {
			continue
		}
		index := result.indices[line]
		input := unescapeCase(result.all[index])
		knownBinary := binaryInputs[input]
		knownMinification := prototypeInputs[input] && string(actual) == "error\tc.resolve is not a function" && string(wanted) == "error\ttag.resolve is not a function"
		if !(knownBinary && string(actual) == "error\tInvalid character" && bytes.HasPrefix(wanted, []byte("ok\t"))) && !knownMinification {
			os.WriteFile(formatterMatchesGo.diagnostic(shard, "library"), external, 0644)
			os.WriteFile(formatterMatchesGo.diagnostic(shard, "expected"), expected, 0644)
			t.Fatalf("new library difference, case %d input %q: %s", index, input, firstDifference(actual, wanted))
		}
		differences++
		t.Logf("recorded original Prettier difference, case %d: %q versus Go %q", index, actual, wanted)
	}
	// Every case with a proved difference must still differ: TestFormatterMatchesGoUnion counts 42 over the corpus.
	if differences != known {
		t.Fatalf("expected exactly %d independently proved bundled-library differences in this shard, got %d", known, differences)
	}
	t.Logf("original Prettier: %d matching cases, %d proved binary/minification differences", len(result.indices)-differences, differences)

	t.Logf("%d repository files, %d of %d cases, %d format answer bytes identical on Go/native/Node/emitted JavaScript", result.files, len(result.indices), len(result.all), len(expected))
}

func unescapeCase(line string) string {
	parts := strings.SplitN(line, "\t", 2)
	return strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t").Replace(parts[1])
}

func TestBundledParserDifference(t *testing.T) {
	t.Parallel()
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

// Not parallel: prepares and publishes the shared file-driver products before parallel shards.
func TestFileDriver_Setup(t *testing.T) {
	if fileDriverShared != nil {
		t.Log("shared setup completed before test execution")
		return
	}
	fileDriverShared = fileDriverSetup(t)
	fileDriverPublish(t, fileDriverShared)
}
