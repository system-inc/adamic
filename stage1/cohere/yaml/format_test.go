package yaml

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
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
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): source}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	goBinary := filepath.Join(t.TempDir(), "go-format")
	run(t, filepath.Join(root, "cohere"), nil, "go", "build", "-overlay", path, "-o", goBinary, "./command/formatter_comparison")
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

// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units); fixed filenames in ADAMIC_YAML_ARTIFACTS; fixed /tmp/stage1-yaml-format-* diagnostic files.
func TestFormatterMatchesGo(t *testing.T) {
	cases, files, count := formatCases(t)
	expected := goFormat(t, cases)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "format")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	nativeOut := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, "--cases", cases)
	runner := filepath.Join(root, "oracle/node.mjs")
	node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, "--cases", cases)
	emitted := filepath.Join(t.TempDir(), "format.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	backend := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, "--cases", cases)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", nativeOut}, {"Node", node}, {"emitted JavaScript", backend}} {
		if !bytes.Equal(side.out, expected) {
			os.WriteFile("/tmp/stage1-yaml-format-expected.txt", expected, 0644)
			os.WriteFile("/tmp/stage1-yaml-format-actual.txt", side.out, 0644)
			t.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	external := run(t, "", nil, "node", "testdata/format_library.mjs", library, "--cases", cases)
	expectedLines := bytes.Split(bytes.TrimSuffix(expected, []byte("\n")), []byte("\n"))
	actualLines := bytes.Split(bytes.TrimSuffix(external, []byte("\n")), []byte("\n"))
	inputs, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	inputLines := strings.Split(strings.TrimSuffix(string(inputs), "\n"), "\n")
	if len(actualLines) != len(expectedLines) {
		t.Fatalf("library answered %d cases, Go %d", len(actualLines), len(expectedLines))
	}
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
	differences := 0
	for index, actual := range actualLines {
		wanted := expectedLines[index]
		if bytes.Equal(actual, wanted) {
			continue
		}
		input := unescapeCase(inputLines[index])
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
	if differences != 42 {
		t.Fatalf("expected exactly 42 independently proved bundled-library differences, got %d", differences)
	}
	t.Logf("original Prettier: %d matching cases, %d proved binary/minification differences", count-differences, differences)

	t.Logf("%d repository files, %d cases, %d format answer bytes identical on Go/native/Node/emitted JavaScript", files, count, len(expected))
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
