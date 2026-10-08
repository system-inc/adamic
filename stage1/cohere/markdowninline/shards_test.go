package markdowninline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Census protocol: literal count, uniform zero-padded shard names. Full corpus
// enumeration must match this constant; sampling keeps the same shard slots.
const testMarkdownInlineShards = 1287
const inlineModes = "wefnspctrukvhijlboq"

// Same shape as internal/buildcache.Inputs; replace this local metadata type
// when that helper lands. There is deliberately no package-local disk cache.
type inlineInputs struct {
	Name                    string
	Files, Flags, Toolchain []string
}

func inlineEnvironment() []string {
	var result []string
	for _, name := range []string{"PATH", "HOME", "XDG_CACHE_HOME", "GOENV", "GOFLAGS", "GOTOOLCHAIN", "GOROOT", "GOPATH", "GOWORK", "GOCACHE", "GOMODCACHE", "CGO_ENABLED", "GOOS", "GOARCH", "GOAMD64", "GOARM64", "GOEXPERIMENT", "GO111MODULE", "GOPROXY", "GONOPROXY", "GONOSUMDB", "GOPRIVATE", "GOSUMDB", "GOVCS", "GODEBUG", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "CC", "CXX", "LIBRARY_PATH", "CPATH", "C_INCLUDE_PATH", "CPLUS_INCLUDE_PATH", "LANG", "LC_ALL", "SOURCE_DATE_EPOCH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET", "LD_LIBRARY_PATH", "DYLD_LIBRARY_PATH", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED", "WASI_SYSROOT", "ADAMIC_RUNTIME_CACHE", "ADAMIC_CHECKER_CACHE", "TMPDIR", "NODE_OPTIONS"} {
		result = append(result, name+"="+os.Getenv(name))
	}
	return result
}
func inlineVersion(name string, args ...string) string {
	out, err := exec.Command(name, args...).CombinedOutput()
	return fmt.Sprintf("%s %v: %s (%v)", name, args, strings.TrimSpace(string(out)), err)
}
func inlineBuild(t *testing.T, in inlineInputs, build func(string) error) string {
	t.Helper()
	dir := t.TempDir()
	start := time.Now()
	if err := build(dir); err != nil {
		t.Fatalf("build %s: %v", in.Name, err)
	}
	t.Logf("build census: %s %.6fs inputs=%+v", in.Name, time.Since(start).Seconds(), in)
	return dir
}
func inlineWrite(dir, name string, data []byte) error {
	return os.WriteFile(filepath.Join(dir, name), data, 0644)
}
func buildInlineGoBridge(dir string) error {
	root, err := filepath.Abs(repository)
	if err != nil {
		return err
	}
	bridge, err := filepath.Abs("testdata/bridge.go")
	if err != nil {
		return err
	}
	driver, err := filepath.Abs("testdata/go_driver.go")
	if err != nil {
		return err
	}
	cohere := filepath.Join(root, "cohere")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(cohere, "internal/format/markdown/adamic_stage_one.go"): bridge, filepath.Join(cohere, "cmd/adamic_stage_one/main.go"): driver}})
	if err != nil {
		return err
	}
	if err = inlineWrite(dir, "overlay.json", data); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-overlay="+filepath.Join(dir, "overlay.json"), "-o", filepath.Join(dir, "go-printer"), filepath.Join(cohere, "cmd/adamic_stage_one/main.go"))
	cmd.Dir = cohere
	out, err := combinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}
func buildInlineProgram(main, dir string) error {
	p, err := load.Load([]string{main})
	if err != nil {
		return err
	}
	program, err := lower.Lower(context.Background(), p)
	if err != nil {
		return err
	}
	if err = inlineWrite(dir, "program.c", []byte(native.C(program))); err != nil {
		return err
	}
	return inlineWrite(dir, "program.mjs", []byte(javascript.JavaScript(program)))
}
func buildInlineLowered(dir string) error {
	main, err := filepath.Abs("main.ts")
	if err != nil {
		return err
	}
	return buildInlineProgram(main, dir)
}

type inlineSource string

func (source inlineSource) buildNative(dir string) error    { return source.buildBinary(dir, false) }
func (source inlineSource) buildSanitized(dir string) error { return source.buildBinary(dir, true) }
func (source inlineSource) buildBinary(dir string, sanitize bool) error {
	data, err := os.ReadFile(filepath.Join(string(source), "program.c"))
	if err != nil {
		return err
	}
	return native.Build(string(data), filepath.Join(dir, "port"), native.Options{Sanitize: sanitize})
}
func (source inlineSource) buildNode(dir string) error {
	script, err := filepath.Abs("testdata/build_node.mjs")
	if err != nil {
		return err
	}
	cmd := exec.Command("node", "--disable-warning=ExperimentalWarning", script, string(source), dir)
	out, err := combinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

type inlineMutation struct{ name, file, from, to string }

func (m inlineMutation) buildLowered(dir string) error {
	for _, name := range []string{"main.ts", "inline.ts", "classes.ts"} {
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		if name == m.file {
			if strings.Count(string(data), m.from) != 1 {
				return fmt.Errorf("%s must change one site", m.name)
			}
			data = []byte(strings.Replace(string(data), m.from, m.to, 1))
		}
		if err = inlineWrite(dir, name, data); err != nil {
			return err
		}
	}
	return buildInlineProgram(filepath.Join(dir, "main.ts"), dir)
}
func inlineGoInputs() inlineInputs {
	return inlineInputs{"markdowninline go bridge", []string{"stage1/cohere/markdowninline/testdata/bridge.go", "stage1/cohere/markdowninline/testdata/go_driver.go", "stage1/cohere/markdowninline/shards_test.go", "cohere", "go.mod", "go.work"}, append([]string{"go build", "-overlay", "-o"}, inlineEnvironment()...), []string{runtime.Version(), inlineVersion("go", "version")}}
}
func inlineProgramInputs(name string) inlineInputs {
	return inlineInputs{name, []string{"stage1/cohere/markdowninline", "internal", "oracle", "cohere", "go.mod", "go.work"}, inlineEnvironment(), []string{runtime.Version(), inlineVersion("go", "version")}}
}
func inlineNativeInputs(name string, sanitize bool) inlineInputs {
	in := inlineProgramInputs(name)
	in.Flags = append(native.Flags(native.Options{Sanitize: sanitize}), in.Flags...)
	in.Toolchain = append(in.Toolchain, inlineVersion("clang", "--version"))
	return in
}
func inlineNodeInputs(name string) inlineInputs {
	in := inlineProgramInputs(name)
	in.Flags = append([]string{"--disable-warning=ExperimentalWarning", "stripTypeScriptTypes mode=transform"}, in.Flags...)
	in.Toolchain = append(in.Toolchain, inlineVersion("node", "--version"))
	return in
}

type inlineShard struct {
	content string
	ids     []int
}

func inlineRange(name string, first, end, modeFirst, modeEnd int) inlineShard {
	s := inlineShard{content: fmt.Sprintf("%s texts [%d,%d) modes [%d,%d)", name, first, end, modeFirst, modeEnd)}
	for i := first; i < end; i++ {
		for m := modeFirst; m < modeEnd; m++ {
			s.ids = append(s.ids, i*len(inlineModes)+m)
		}
	}
	return s
}
func enumerateInlineShards(texts []string, files, generatedEnd, unicodeEnd int) []inlineShard {
	var shards []inlineShard
	// Contiguous corpus runs: at most eight files and 32 KiB of source.
	// Large singleton texts retain their exact content, split by mode instead.
	add := func(name string, first, end int) {
		if end == first+1 && len(texts[first]) > 32*1024 {
			for m := range len(inlineModes) {
				shards = append(shards, inlineRange(name, first, end, m, m+1))
			}
		} else {
			shards = append(shards, inlineRange(name, first, end, 0, len(inlineModes)))
		}
	}
	for first := 0; first < files; {
		end := first + 1
		size := len(texts[first])
		for end < files && end-first < 8 && size+len(texts[end]) <= 32*1024 {
			size += len(texts[end])
			end++
		}
		add("corpus", first, end)
		first = end
	}
	for first := files; first < generatedEnd; {
		end := first + 1
		if len(texts[first]) <= 32*1024 {
			for end < generatedEnd && end-first < 128 && len(texts[end]) <= 32*1024 {
				end++
			}
		}
		add("generated", first, end)
		first = end
	}
	add("unicode", generatedEnd, unicodeEnd)
	add("astral", unicodeEnd, len(texts))
	return shards
}
func inlineUnion(shards []inlineShard, expected map[int]bool) error {
	seen := map[int]bool{}
	count := 0
	for i, s := range shards {
		for _, id := range s.ids {
			count++
			if !expected[id] {
				return fmt.Errorf("shard-%04d foreign case %d", i, id)
			}
			if seen[id] {
				return fmt.Errorf("shard-%04d repeated case %d", i, id)
			}
			seen[id] = true
		}
	}
	if count != len(expected) || len(seen) != len(expected) {
		return fmt.Errorf("union has %d cases/%d ids, want %d", count, len(seen), len(expected))
	}
	for id := range expected {
		if !seen[id] {
			return fmt.Errorf("missing case %d", id)
		}
	}
	return nil
}
func inlineBox() (int, int, error) {
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return 0, 1, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("ADAMIC_TEST_SHARD must be zero-based i/n")
	}
	i, e := strconv.Atoi(parts[0])
	n, f := strconv.Atoi(parts[1])
	if e != nil || f != nil || n < 1 || i < 0 || i >= n {
		return 0, 0, fmt.Errorf("invalid ADAMIC_TEST_SHARD=%q", value)
	}
	return i, n, nil
}

var inlineEncoder = strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

func inlineInput(texts []string, ids []int) []byte {
	var input strings.Builder
	for _, id := range ids {
		input.WriteByte(inlineModes[id%len(inlineModes)])
		input.WriteString(inlineEncoder.Replace(texts[id/len(inlineModes)]))
		input.WriteByte('\n')
	}
	return []byte(input.String())
}
func inlineDifference(name string, ids []int, got, want []byte) error {
	if bytes.Equal(got, want) {
		return nil
	}
	offset := 0
	for offset < len(got) && offset < len(want) && got[offset] == want[offset] {
		offset++
	}
	row := bytes.Count(want[:min(offset, len(want))], []byte("\n"))
	id := -1
	if row < len(ids) {
		id = ids[row]
	}
	return fmt.Errorf("%s disagreement: case %d mode %c byte %d lengths %d/%d", name, id, inlineModes[max(0, id)%len(inlineModes)], offset, len(got), len(want))
}
func inlineNativeRun(t *testing.T, sanitized, fast string, args ...string) run {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		return execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, args...)
	case "darwin":
		r := execute(t, nil, sanitized, args...)
		check := execute(t, nil, "leaks", append([]string{"--atExit", "--", fast}, args...)...)
		if check.exitCode != 0 {
			t.Fatalf("leaks: %s %s", check.stdout, check.stderr)
		}
		return r
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
		return run{}
	}
}
func runInlineShards(t *testing.T, paths, texts []string, files, generatedEnd, unicodeEnd int, selected map[string]bool) {
	if len(texts)*len(inlineModes) != 115045 {
		t.Fatalf("full enumeration: %d cases, want 115045", len(texts)*len(inlineModes))
	}
	shards := enumerateInlineShards(texts, files, generatedEnd, unicodeEnd)
	if len(shards) != testMarkdownInlineShards {
		t.Fatalf("enumerated %d shards; census constant is %d", len(shards), testMarkdownInlineShards)
	}
	expected := map[int]bool{}
	for text := range texts {
		if text >= files || selected[paths[text]] {
			for mode := range len(inlineModes) {
				expected[text*len(inlineModes)+mode] = true
			}
		}
	}
	for i := range shards {
		var ids []int
		for _, id := range shards[i].ids {
			if expected[id] {
				ids = append(ids, id)
			}
		}
		shards[i].ids = ids
	}
	if err := inlineUnion(shards, expected); err != nil {
		t.Fatal(err)
	}
	t.Logf("counted union: %d cases, each id exactly once; %d shards", len(expected), len(shards))
	box, boxes, err := inlineBox()
	if err != nil {
		t.Fatal(err)
	}
	goDir := inlineBuild(t, inlineGoInputs(), buildInlineGoBridge)
	goBinary := filepath.Join(goDir, "go-printer")
	source, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	nodeDir := inlineBuild(t, inlineNodeInputs("markdowninline Node oracle"), inlineSource(source).buildNode)
	lowerDir := inlineBuild(t, inlineProgramInputs("markdowninline lowered program"), buildInlineLowered)
	nativeDir := inlineBuild(t, inlineNativeInputs("markdowninline native", false), inlineSource(lowerDir).buildNative)
	sanitizedDir := inlineBuild(t, inlineNativeInputs("markdowninline sanitized", true), inlineSource(lowerDir).buildSanitized)
	fast, sanitized := filepath.Join(nativeDir, "port"), filepath.Join(sanitizedDir, "port")
	planted := inlineMutation{"one-case disagreement", "main.ts", "console.log(encode(formatLeaf(line.slice(0, 1), decode(line.slice(1)))));", "console.log(line === 'w😀😀😀😀' ? 'planted disagreement' : encode(formatLeaf(line.slice(0, 1), decode(line.slice(1)))));"}
	mutantDir := inlineBuild(t, inlineProgramInputs("markdowninline planted lowered"), planted.buildLowered)
	mutantNativeDir := inlineBuild(t, inlineNativeInputs("markdowninline planted sanitized", true), inlineSource(mutantDir).buildSanitized)
	mutantFast := fast
	if runtime.GOOS == "darwin" {
		mutantFast = filepath.Join(inlineBuild(t, inlineNativeInputs("markdowninline planted native", false), inlineSource(mutantDir).buildNative), "port")
	}
	target := -1
	for text := range texts {
		if texts[text] == "😀😀😀😀" {
			if target != -1 {
				t.Fatal("planted text repeated")
			}
			target = text * len(inlineModes)
		}
	}
	if target < 0 {
		t.Fatal("planted case absent")
	}
	targetOwners := 0
	for _, s := range shards {
		for _, id := range s.ids {
			if id == target {
				targetOwners++
			}
		}
	}
	if targetOwners != 1 {
		t.Fatalf("planted case has %d owners", targetOwners)
	}
	// Build the original mutants once too; their small generated witness corpus
	// avoids sending every large repository text through known-wrong programs.
	type mutantProduct struct{ name, binary, fast, node string }
	var mutants []mutantProduct
	for _, m := range []inlineMutation{{"escaped delimiter parity", "inline.ts", "(found.preceding - position) % 2 === 1", "(found.preceding - position) % 2 === 0"}, {"table pipe escaping", "inline.ts", "if(table)", "if(!table)"}, {"minimum absent fence", "inline.ts", "while(runs.includes(count))", "while(runs.includes(count) && count < 1)"}} {
		dir := inlineBuild(t, inlineProgramInputs("markdowninline mutant "+m.name+" lowered"), m.buildLowered)
		bin := inlineBuild(t, inlineNativeInputs("markdowninline mutant "+m.name+" sanitized", true), inlineSource(dir).buildSanitized)
		node := inlineBuild(t, inlineNodeInputs("markdowninline mutant "+m.name+" Node"), inlineSource(dir).buildNode)
		mf := fast
		if runtime.GOOS == "darwin" {
			mf = filepath.Join(inlineBuild(t, inlineNativeInputs("markdowninline mutant "+m.name+" native", false), inlineSource(dir).buildNative), "port")
		}
		mutants = append(mutants, mutantProduct{m.name, filepath.Join(bin, "port"), mf, filepath.Join(node, "main.mjs")})
	}
	library := os.Getenv("ADAMIC_MARKDOWNINLINE_LIBRARY")
	if library == "" {
		t.Log("external library not checked: set ADAMIC_MARKDOWNINLINE_LIBRARY")
	}
	// Each gate-addressable unit contains only its cases. No nested parallel unit
	// hides time from the shard's own census line.
	for ordinal, shard := range shards {
		if ordinal%boxes != box {
			continue
		}
		dir := t.TempDir()
		name := fmt.Sprintf("shard-%04d", ordinal)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			t.Logf("%s: %d cases", shard.content, len(shard.ids))
			if len(shard.ids) == 0 && ordinal != 0 {
				return
			}
			cases := filepath.Join(dir, "cases.txt")
			write(t, cases, inlineInput(texts, shard.ids))
			want := execute(t, nil, goBinary, cases)
			clean(t, "Go", want)
			compare := func(side string, r run) {
				clean(t, side, r)
				if err := inlineDifference(t.Name()+" "+side, shard.ids, r.stdout, want.stdout); err != nil {
					t.Fatal(err)
				}
			}
			compare("native", inlineNativeRun(t, sanitized, fast, "--batch", cases))
			compare("Node", onNode(t, filepath.Join(nodeDir, "main.mjs"), "--batch", cases))
			compare("JavaScript backend", onNode(t, filepath.Join(lowerDir, "program.mjs"), "--batch", cases))
			if library != "" {
				compare("Prettier", execute(t, nil, "node", "testdata/library.mjs", library, cases))
			}
			owns := false
			for _, id := range shard.ids {
				if id == target {
					owns = true
				}
			}
			if owns {
				r := inlineNativeRun(t, filepath.Join(mutantNativeDir, "port"), mutantFast, "--batch", cases)
				clean(t, "planted mutant", r)
				err := inlineDifference(t.Name(), shard.ids, r.stdout, want.stdout)
				if err == nil {
					t.Fatal("planted mutant survived")
				}
				a, b := bytes.Split(r.stdout, []byte("\n")), bytes.Split(want.stdout, []byte("\n"))
				differences := 0
				if len(a) != len(b) {
					t.Fatal("planted mutant changed row count")
				}
				for row := range b {
					if !bytes.Equal(a[row], b[row]) {
						differences++
						if row >= len(shard.ids) || shard.ids[row] != target {
							t.Fatalf("mutant changed foreign case %d", row)
						}
					}
				}
				if differences != 1 {
					t.Fatalf("planted mutant changed %d cases", differences)
				}
				t.Logf("planted mutant caught exactly once: %v", err)
			}
			for _, id := range shard.ids {
				if id%len(inlineModes) != 0 || id/len(inlineModes) >= files {
					continue
				}
				text := texts[id/len(inlineModes)]
				raw := filepath.Join(dir, "raw.md")
				write(t, raw, []byte(text))
				single := filepath.Join(dir, "single.txt")
				write(t, single, inlineInput(texts, []int{id}))
				answer := execute(t, nil, goBinary, single)
				clean(t, "Go raw", answer)
				decoded := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`).Replace(strings.TrimSuffix(string(answer.stdout), "\n"))
				r := inlineNativeRun(t, sanitized, fast, raw, "w")
				clean(t, "raw", r)
				equal(t, "raw", r.stdout, []byte(decoded))
			}
			if ordinal == 0 {
				for _, text := range []string{"", "a", "a\n", "a\r\n\n", "😀*x*", "\\_"} {
					raw := filepath.Join(dir, "raw-extra.md")
					write(t, raw, []byte(text))
					single := filepath.Join(dir, "raw-extra.txt")
					write(t, single, []byte("w"+inlineEncoder.Replace(text)+"\n"))
					answer := execute(t, nil, goBinary, single)
					clean(t, "Go raw witness", answer)
					decoded := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`).Replace(strings.TrimSuffix(string(answer.stdout), "\n"))
					r := inlineNativeRun(t, sanitized, fast, raw, "w")
					clean(t, "raw witness", r)
					equal(t, "raw witness", r.stdout, []byte(decoded))
				}
			}

			// Original mutants use a single small full generated witness set in shard 0.
			if ordinal == 0 {
				var ids []int
				for text := files; text < generatedEnd; text++ {
					if len(texts[text]) <= 128 {
						for mode := range len(inlineModes) {
							ids = append(ids, text*len(inlineModes)+mode)
						}
					}
				}
				probe := filepath.Join(dir, "mutants.txt")
				write(t, probe, inlineInput(texts, ids))
				answer := execute(t, nil, goBinary, probe)
				clean(t, "Go mutant witnesses", answer)
				for _, m := range mutants {
					for _, r := range []run{inlineNativeRun(t, m.binary, m.fast, "--batch", probe), onNode(t, m.node, "--batch", probe)} {
						clean(t, m.name, r)
						if err := inlineDifference(t.Name()+" "+m.name, ids, r.stdout, answer.stdout); err == nil {
							t.Fatalf("%s survived", m.name)
						} else {
							t.Log(err)
						}
					}
				}
			}
		})
	}
}
func TestMarkdownInlineShardUnion(t *testing.T) {
	expected := map[int]bool{0: true, 1: true, 2: true}
	if err := inlineUnion([]inlineShard{{ids: []int{0, 1}}, {ids: []int{2}}}, expected); err != nil {
		t.Fatal(err)
	}
	for _, shards := range [][]inlineShard{{{ids: []int{0, 1}}}, {{ids: []int{0, 1, 2, 2}}}, {{ids: []int{0, 1, 3}}}} {
		if inlineUnion(shards, expected) == nil {
			t.Fatal("invalid union accepted")
		}
	}
}

func TestMarkdownInlineShardSelector(t *testing.T) {
	for _, value := range []string{"", "0/1", "2/3"} {
		t.Setenv("ADAMIC_TEST_SHARD", value)
		if _, _, err := inlineBox(); err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []string{"1", "0/0", "-1/3", "3/3", "x/2"} {
		t.Setenv("ADAMIC_TEST_SHARD", value)
		if _, _, err := inlineBox(); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
}
