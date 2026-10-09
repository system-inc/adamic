package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."
const compilerCommit = "050880ce59e30b356b686bd3144efe24f875ebc8"

var portFiles = []string{"characters.ts", "tokens.ts", "scanner.ts", "main.ts"}

type execution struct {
	output   []byte
	duration time.Duration
}

// Guarded output is captured in a file so the large corpus need not stay in memory while running.
func execute(t *testing.T, directory, name string, args ...string) execution {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = childguard.Run(command, childguard.Options{})
	duration := time.Since(started)
	if err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, duration}
}

func goOracle(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/oracle.go")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(root, "adamic_scanner_oracle.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "oracle")
	execute(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}

func buildPort(t *testing.T, directory string, sanitize bool) string {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "scanner")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitize}); err != nil {
		t.Fatal(err)
	}
	return binary
}

func copyPort(t *testing.T, file, from, to string) string {
	t.Helper()
	directory := t.TempDir()
	for _, name := range portFiles {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		if name == file {
			if strings.Count(source, from) != 1 {
				t.Fatalf("mutant must change exactly one site in %s: %q", file, from)
			}
			source = strings.Replace(source, from, to, 1)
		}
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func node(t *testing.T, directory, manifest string, count bool) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", manifest}
	if count {
		args = append(args, "--count")
	}
	return execute(t, "", "node", args...)
}

func difference(got, want []byte) string {
	if bytes.Equal(got, want) {
		return ""
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) || i < len(b); i++ {
		left, right := "<EOF>", "<EOF>"
		if i < len(a) {
			left = a[i]
		}
		if i < len(b) {
			right = b[i]
		}
		if left != right {
			return fmt.Sprintf("line %d: port %q, Go %q", i+1, left, right)
		}
	}
	return "different bytes"
}

func compilerSource(t *testing.T) string {
	t.Helper()
	directory := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if directory == "" {
		directory = filepath.Join(t.TempDir(), "typescript")
		// Clone messages are expected on stderr, so this setup command has its own log.
		command := exec.Command("git", "clone", "--quiet", "--depth", "1", "--branch", "v6.0.3", "https://github.com/microsoft/TypeScript.git", directory)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("clone pinned corpus: %v\n%s", err, output)
		}
	}
	// An environment override must be the same pinned checkout, not merely a directory with .ts files.
	command := exec.Command("git", "-C", directory, "rev-parse", "HEAD")
	output, err := command.Output()
	if err != nil || strings.TrimSpace(string(output)) != compilerCommit {
		t.Fatalf("TypeScript corpus must be v6.0.3 at %s, got %q (%v)", compilerCommit, output, err)
	}
	return filepath.Join(directory, "src/compiler")
}

func sourceFiles(t *testing.T, directory string) []string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(directory) == filepath.Join(root, "stage1") || filepath.Clean(directory) == filepath.Clean(filepath.Join(repository, "stage1")) {
		return corpusfiles.Repository(t, root, []string{"stage1"}, []string{"*.ts"})
	}
	return corpusfiles.Upstream(t, filepath.Dir(filepath.Dir(directory)), compilerCommit, []string{"src/compiler"}, []string{"*.ts"})
}

type corpus struct {
	all, edges, compiler string
	files, generated     int
}

func askedCorpus(t *testing.T) corpus {
	t.Helper()
	directory := t.TempDir()
	var all, edges, compiler strings.Builder
	compilerFiles := sourceFiles(t, compilerSource(t))
	for _, path := range compilerFiles {
		fmt.Fprintf(&compiler, "scan\t%s\n", path)
	}
	all.WriteString(compiler.String())
	stageFiles := sourceFiles(t, filepath.Join(repository, "stage1"))
	for _, path := range stageFiles {
		fmt.Fprintf(&all, "scan\t%s\n", path)
	}
	generated := 0
	add := func(mode, text string) {
		path := filepath.Join(directory, fmt.Sprintf("edge-%05d.ts", generated))
		generated++
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&edges, "%s\t%s\n", mode, path)
	}
	tokens, err := os.ReadFile(filepath.Join(repository, "cohere/TypeScript/tsc/internal/scanner/scanner.go"))
	if err != nil {
		t.Fatal(err)
	}
	// The oracle's table determines coverage, independently of the port's table.
	spellings := regexp.MustCompile(`"([^"\\]*(?:\\.[^"\\]*)*)":\s+ast.Kind(\w+)`).FindAllStringSubmatch(string(tokens), -1)
	if len(spellings) < 100 {
		t.Fatal("oracle token table was not found")
	}
	for _, fields := range spellings {
		spelling, err := strconv.Unquote(`"` + fields[1] + `"`)
		if err != nil {
			t.Fatal(err)
		}
		add("scan", spelling)
		if strings.Contains(fields[2], "Token") {
			add("greater", spelling)
		}
	}
	for _, literal := range []string{"0", "9007199254740993", "1.25", ".5", "1.", "1e3", "1E+3", "1e-9", "1e", "1e+", "01", "077", "088", "0_1", "1_2", "1__2", "1_", "_1", "1._2", "1.2_", "1e_2", "1e2_", "0x", "0XfF", "0x_1", "0x1__2", "0x1_", "0b", "0b_1", "0b102", "0o", "0o8", "0o7_7", "0o_7", "0o7__7", "1n", "1_2n", "0xFFn", "1.1n", "1e2n", "123abc", "0x1foo", "1_n", "01n", "0" + strings.Repeat("7", 40)} {
		add("scan", literal)
	}
	for _, prefix := range []string{"", "0x", "0X", "0o", "0O", "0b", "0B"} {
		for length := 1; length <= 100; length++ {
			for _, digit := range []string{"0", "1", "7", "9", "f"} {
				literal := prefix + strings.Repeat(digit, length)
				add("scan", literal)
				add("scan", literal+"n")
			}
		}
	}
	for _, literal := range []string{"1", "12", "123", "1.2", "1e2", "1e+2", "0xf", "0o7", "0b1"} {
		for position := 0; position <= len(literal); position++ {
			for _, separator := range []string{"_", "__", "___"} {
				add("scan", literal[:position]+separator+literal[position:])
			}
		}
	}
	escapes := []string{`\é`, `\😀`, "\\\u2028", "\\\u2029", `\x00`, `\xFF`, `\x0`, `\xGG`, `\u0000`, `\u00ff`, `\uD800`, `\uD83D\uDE00`, `\u{d800}\u{dc00}`, `\u{10ffff}`, `\u{110000}`, `\u{}`, `\u{0`, `\u{123q}`, `\u123`, `\u000G`, `\u{FFFFFFFFFFFFFFFFFFFFFFFF}`}
	for code := rune(0); code < 128; code++ {
		escapes = append(escapes, "\\"+string(code))
	}
	for number := 0; number < 512; number++ {
		escapes = append(escapes, fmt.Sprintf(`\%03o`, number))
	}
	for _, escape := range escapes {
		for _, quote := range []string{`"`, `'`, "`"} {
			add("scan", quote+escape+quote)
			add("scan", quote+escape)
		}
	}
	for _, text := range []string{"", "é", "日本語", "𐐀", "a\u200c", "a\u200d", "a\u0300", "\u0300a", `\u0061`, `a\u0062`, `\u{10400}`, `\uD801\uDC00`, `a\u{}`, `a\u0020`, `a\u{110000}`, `\u{0}`, "#name", `#\u0061`, "#", "#1", "#!x\nx", "x #!x", "\ufffd x", "\ufeff a", "\u0085 a", "a\u2028b", "a\u2029b", "/** @deprecated @link x */ a", "/** @deprecatedx */ a", "/*x", "//x", "<<<<<<< ours\na\n=======\nb\n>>>>>>> theirs", "`a${b}c${d}e`", "`a\r\nb`"} {
		add("scan", text)
	}
	// Every Unicode identifier range's two ends, one before and one after, and one stride.
	table, err := os.ReadFile(filepath.Join(repository, "cohere/TypeScript/tsc/internal/stringutil/identifier_parts_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	ranges := regexp.MustCompile(`\{(0x[0-9A-Fa-f]+), (0x[0-9A-Fa-f]+), (\d+)\}`).FindAllStringSubmatch(string(table), -1)
	if len(ranges) < 1000 {
		t.Fatal("oracle Unicode ranges were not found")
	}
	for _, match := range ranges {
		fields := match[1:]
		low, err := strconv.ParseInt(fields[0], 0, 32)
		if err != nil {
			t.Fatal(err)
		}
		high, err := strconv.ParseInt(fields[1], 0, 32)
		if err != nil {
			t.Fatal(err)
		}
		stride, err := strconv.Atoi(strings.TrimSuffix(fields[2], ","))
		if err != nil {
			t.Fatal(err)
		}
		for _, point := range []int64{low - 1, low, low + int64(stride), high, high + 1} {
			if point < 0 || point > 0x10ffff || point >= 0xd800 && point <= 0xdfff {
				continue
			}
			character := string(rune(point))
			add("scan", character+" a"+character+" "+fmt.Sprintf(`\u{%x} a\u{%x}`, point, point))
		}
	}
	for _, text := range []string{`"😀𐐀"`, `'😀𐐀'`, `"😀𐐀`, "`😀𐐀`", "`😀𐐀", "/*😀𐐀*/x", "/*😀𐐀", "//😀𐐀\nx", "//😀\u2028x", "/*😀\u2029*/x", "#!😀\nx"} {
		add("scan", text)
	}
	add("template", "`😀${x}𐐀`")
	for _, text := range []string{"`a${b}c${d}e`", "`a${b}\\uZZZZ`", "`a${b}\\1`"} {
		add("template", text)
	}
	for _, text := range []string{`/abc/g`, `/a\/b/i`, `/[/]/g`, `/[[]/v`, `/a`, `/a ;`, `/a) ; x`, `/[a`, "/a\nnext", "/a\u2028b/g", `/=abc/`, `/x/é`, `/\u{d800}/u`, `/(?<x>a)/`, `/a/vu`, `/x/gg`, `a/b/c`, "//x\n/ok/"} {
		add("regex", text)
		add("scan", text)
	}
	for _, text := range []string{" hello\nworld <x> { }", "\n  \t", "x>y}z", "é😀", "<x>words</x>", "\n\n<x>"} {
		add("jsx", text)
	}
	all.WriteString(edges.String())
	write := func(name, text string) string {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	t.Logf("%d compiler files, %d stage1 files, %d generated inputs", len(compilerFiles), len(stageFiles), generated)
	return corpus{write("all.txt", all.String()), write("edges.txt", edges.String()), write("compiler.txt", compiler.String()), len(compilerFiles) + len(stageFiles), generated}
}

func TestScannerAgreesWithTypescriptGo(t *testing.T) {
	t.Parallel()
	asked := askedCorpus(t)
	oracle := goOracle(t)
	directory := copyPort(t, "", "", "")
	binary := buildPort(t, directory, true)
	want := execute(t, "", oracle, "--manifest", asked.all)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, asked.all, false)},
		{"native under ASan, UBSan and LeakSanitizer", execute(t, "", binary, "--manifest", asked.all)},
	} {
		if diff := difference(side.run.output, want.output); diff != "" {
			t.Errorf("%s: %s", side.name, diff)
		}
	}
	if !t.Failed() {
		t.Logf("%d answer bytes identical", len(want.output))
	}
	edgeWant := execute(t, "", oracle, "--manifest", asked.edges)
	mutants := []struct{ name, file, from, to string }{
		{"source scanned as Identifier", "tokens.ts", `['source', 'SourceKeyword']`, `['source', 'Identifier']`},
		{"punctuator != scanned as ==", "tokens.ts", `['!=', 'ExclamationEqualsToken']`, `['!=', 'EqualsEqualsToken']`},
		{"invalid decimal separator accepted", "scanner.ts", "this.error(previous ? 6189 : 6188, this.pos, 1);", "if (base !== 10) { this.error(previous ? 6189 : 6188, this.pos, 1); } else { this.flags &= ~16384; }"},
		{"regex rescan skipped", "main.ts", "scanner.rescanSlash();", "scanner.code();"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			mutated := copyPort(t, mutant.file, mutant.from, mutant.to)
			nativeMutant := buildPort(t, mutated, true)
			for _, side := range []struct {
				name string
				run  execution
			}{
				{"Node", node(t, mutated, asked.edges, false)},
				{"native", execute(t, "", nativeMutant, "--manifest", asked.edges)},
			} {
				diff := difference(side.run.output, edgeWant.output)
				if diff == "" {
					t.Errorf("%s: mutant survives comparison", side.name)
				} else {
					t.Logf("%s caught by byte comparison: %s", side.name, diff)
				}
			}
		})
	}
}

// Not parallel: best-of-five scanner throughput timing requires exclusive CPU use.
func TestPerformance(t *testing.T) {
	if os.Getenv("ADAMIC_SCANNER_BENCH") == "" {
		t.Skip("set ADAMIC_SCANNER_BENCH=1 for best-of-five throughput")
	}
	asked := askedCorpus(t)
	oracle := goOracle(t)
	directory := copyPort(t, "", "", "")
	binary := buildPort(t, directory, false)
	want := execute(t, "", oracle, "--manifest", asked.compiler, "--count").output
	tokens, err := strconv.Atoi(strings.TrimSpace(string(want)))
	if err != nil {
		t.Fatal(err)
	}
	best := map[string]time.Duration{}
	for round := 0; round < 5; round++ {
		for _, side := range []struct {
			name string
			run  func() execution
		}{
			{"native", func() execution { return execute(t, "", binary, "--manifest", asked.compiler, "--count") }},
			{"Go", func() execution { return execute(t, "", oracle, "--manifest", asked.compiler, "--count") }},
			{"Node", func() execution { return node(t, directory, asked.compiler, true) }},
		} {
			result := side.run()
			if !bytes.Equal(result.output, want) {
				t.Fatalf("%s count %q, Go %q", side.name, result.output, want)
			}
			if best[side.name] == 0 || result.duration < best[side.name] {
				best[side.name] = result.duration
			}
			t.Logf("round %d %s %.6fs", round+1, side.name, result.duration.Seconds())
		}
	}
	for _, name := range []string{"native", "Go", "Node"} {
		t.Logf("%s best of 5: %d tokens / %.6fs = %.0f tokens/s", name, tokens, best[name].Seconds(), float64(tokens)/best[name].Seconds())
	}
}
