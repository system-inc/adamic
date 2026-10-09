package tsgo_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"

	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Not parallel: archive builds, sanitizer subprocesses and timing use the same
// machine. The corpus is supplied locally; tests never fetch from the network.
func TestBridge(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	run := func(name string, command *exec.Cmd) ([]byte, error) {
		t.Helper()
		command.Dir = repository
		output, err := command.CombinedOutput()
		if writeError := os.WriteFile(filepath.Join(scratch, name+".log"), output, 0o644); writeError != nil {
			t.Fatal(writeError)
		}
		return output, err
	}
	mustRun := func(name string, command *exec.Cmd) []byte {
		t.Helper()
		output, err := run(name, command)
		if err != nil {
			t.Fatalf("%s: %v\n%s", name, err, output)
		}
		return output
	}
	sanitized := filepath.Join(scratch, "tsgo-asan.a")
	normal := filepath.Join(scratch, "tsgo.a")
	buildArchive := func(name, output, overlay string, sanitize bool) {
		t.Helper()
		arguments := []string{"build", "-buildmode=c-archive", "-o", output}
		if overlay != "" {
			arguments = append(arguments, "-overlay", overlay)
		}
		arguments = append(arguments, "./bridge/tsgo/archive")
		command := exec.Command("go", arguments...)
		if sanitize {
			command.Env = append(os.Environ(), "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
		}
		mustRun(name, command)
	}
	buildArchive("archive", normal, "", false)
	buildArchive("archive-asan", sanitized, "", true)
	driver := filepath.Join(scratch, "api")
	linkDriver := func(name, archive, output string) {
		t.Helper()
		mustRun(name, exec.Command("clang", "-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O1", "-g", "-fsanitize=address,undefined", "-I", filepath.Join(repository, "bridge/tsgo"), filepath.Join(repository, "bridge/tsgo/testdata/api.c"), archive, "-lpthread", "-ldl", "-lm", "-o", output))
	}
	linkDriver("api-link", sanitized, driver)
	config := filepath.Join(repository, "bridge/tsgo/testdata/tsconfig.json")
	sample := filepath.Join(repository, "bridge/tsgo/testdata/sample.ts")
	output := mustRun("api", exec.Command(driver, config, sample, "check"))
	t.Logf("C ABI: 100 queries, outputs survive release, stale and zero handles rejected, second handle distinct; %s", output)
	output, err = run("input-length-mutant", exec.Command(driver, config, sample, "length"))
	if err == nil || !bytes.Contains(output, []byte("AddressSanitizer: heap-buffer-overflow")) {
		t.Fatalf("input length mutant escaped ASan: %v\n%s", err, output)
	}
	t.Log("input length off by one: ASan heap-buffer-overflow")

	stage0 := filepath.Join(scratch, "adamic")
	oracle := filepath.Join(scratch, "oracle")
	mustRun("stage0-build", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	mustRun("oracle-build", exec.Command("go", "build", "-o", oracle, "./bridge/tsgo/oracle"))
	fixture := filepath.Join(repository, "bridge/tsgo/testdata/queries.a")
	for _, arguments := range [][]string{{"build", fixture, "-o", filepath.Join(scratch, "unlinked")}, {"c", fixture}, {"js", fixture}} {
		output, err = run("refusal-"+arguments[0], exec.Command(stage0, arguments...))
		if err == nil || !bytes.Contains(output, []byte("unlinked typescript-go library call")) {
			t.Fatalf("unlinked %s was not refused: %v\n%s", arguments[0], err, output)
		}
	}
	t.Log("build, C and JavaScript commands refuse unlinked checker calls")
	nativeBinary := filepath.Join(scratch, "native-asan")
	mustRun("native-build", exec.Command(stage0, "build", fixture, "-o", nativeBinary, "--tsgo", sanitized, "--sanitize"))
	optimized := filepath.Join(scratch, "native")
	mustRun("native-optimized-build", exec.Command(stage0, "build", fixture, "-o", optimized, "--tsgo", normal))

	roots := []string{sample}
	if corpus := os.Getenv("ADAMIC_TSGO_CORPUS"); corpus != "" {
		// v6.0.3, pinned in README. Four different checker workloads, 400 positions each.
		roots = nil
		for _, file := range []string{"checker.ts", "parser.ts", "types.ts", "utilities.ts"} {
			roots = append(roots, filepath.Join(corpus, "src/compiler", file))
		}
	}
	manifest := filepath.Join(scratch, "queries.tsv")
	var queries strings.Builder
	count := 0
	for _, file := range roots {
		text, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		positions := 400
		if len(text) < positions {
			positions = len(text)
		}
		for index := 0; index < positions; index++ {
			fmt.Fprintf(&queries, "%s\t%d\n", file, index*len(text)/positions)
			count++
		}
	}
	if err := os.WriteFile(manifest, []byte(queries.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	arguments := append([]string{config, manifest}, roots...)
	// Keep stdout separate from timing and sanitizer reports: only answers compare.
	answers := func(name, binary string, timing bool) []byte {
		t.Helper()
		command := exec.Command(binary, arguments...)
		command.Dir = repository
		if timing {
			command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		}
		var stderr bytes.Buffer
		command.Stderr = &stderr
		stdout, err := command.Output()
		if writeError := os.WriteFile(filepath.Join(scratch, name+".log"), stderr.Bytes(), 0o644); writeError != nil {
			t.Fatal(writeError)
		}
		if err != nil {
			t.Fatalf("%s: %v\n%s", name, err, stderr.String())
		}
		if timing {
			t.Logf("%s: %s", name, strings.TrimSpace(stderr.String()))
		} else if stderr.Len() != 0 {
			t.Fatalf("%s sanitizer stderr: %s", name, stderr.String())
		}
		return stdout
	}
	truth := answers("go", oracle, true)
	observed := answers("native-asan", nativeBinary, false)
	if !bytes.Equal(truth, observed) {
		t.Fatalf("native oracle mismatch at byte %d", firstDifference(truth, observed))
	}
	t.Logf("oracle: %d positions across %d files, %d bytes identical under ASan/UBSan/LSan", count, len(roots), len(truth))
	for round := 1; round <= 3; round++ {
		observed := answers("native-round-"+strconv.Itoa(round), optimized, true)
		direct := answers("go-round-"+strconv.Itoa(round), oracle, true)
		if !bytes.Equal(truth, observed) || !bytes.Equal(truth, direct) {
			t.Fatal("timed answers changed")
		}
	}

	// Overlay builds change production code, without mutating the checkout.
	overlay := func(name, path, before, after string) string {
		t.Helper()
		original := filepath.Join(repository, path)
		text, err := os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(text), before) != 1 {
			t.Fatalf("mutant %s has no unique target", name)
		}
		replacement := filepath.Join(scratch, name+filepath.Ext(path))
		if err := os.WriteFile(replacement, []byte(strings.Replace(string(text), before, after, 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
		if err != nil {
			t.Fatal(err)
		}
		path = filepath.Join(scratch, name+"-overlay.json")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	typeLengthOverlay := overlay("output-length", "bridge/tsgo/archive/main.go", "result._type = buffer(answer.Type)", "result._type = buffer(answer.Type)\n\tresult._type.length++")
	typeLengthArchive := filepath.Join(scratch, "length.a")
	buildArchive("output-length-build", typeLengthArchive, typeLengthOverlay, false)
	lengthDriver := filepath.Join(scratch, "length-driver")
	linkDriver("output-length-link", typeLengthArchive, lengthDriver)
	output, err = run("output-length-mutant", exec.Command(lengthDriver, config, sample, "check"))
	if err == nil || !bytes.Contains(output, []byte("AddressSanitizer: heap-buffer-overflow")) {
		t.Fatalf("output length mutant escaped ASan: %v\n%s", err, output)
	}
	t.Log("output string length off by one: ASan heap-buffer-overflow")
	staleOverlay := overlay("stale-handle", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant: retain the released program in the live registry.")
	staleArchive := filepath.Join(scratch, "stale.a")
	buildArchive("stale-build", staleArchive, staleOverlay, false)
	staleDriver := filepath.Join(scratch, "stale-driver")
	linkDriver("stale-link", staleArchive, staleDriver)
	output, err = run("stale-mutant", exec.Command(staleDriver, config, sample, "check"))
	if err == nil || !bytes.Contains(output, []byte("tsgo_query(handle, file, 14, &answer, &error) == TSGO_HANDLE")) {
		t.Fatalf("released handle mutant escaped: %v\n%s", err, output)
	}
	t.Log("released handle kept live: stale-handle assertion catches it")
	wrongOverlay := overlay("wrong-position", "bridge/tsgo/checker/program.go", "Type: typeChecker.TypeToString(typeChecker.GetTypeAtLocation(node))", "Type: typeChecker.TypeToString(typeChecker.GetTypeAtLocation(source.AsNode()))")
	wrongArchive := filepath.Join(scratch, "wrong.a")
	buildArchive("wrong-build", wrongArchive, wrongOverlay, false)
	wrongBinary := filepath.Join(scratch, "wrong-native")
	mustRun("wrong-native-build", exec.Command(stage0, "build", fixture, "-o", wrongBinary, "--tsgo", wrongArchive, "--sanitize"))
	observed = answers("wrong-position", wrongBinary, false)
	if bytes.Equal(truth, observed) {
		t.Fatal("wrong-position type string escaped oracle")
	}
	t.Logf("type from source-file position: oracle mismatch at byte %d", firstDifference(truth, observed))

	linkOverlay := overlay("linkage", "internal/lower/tsgo.go", "if !l.program.TSGoEnabled() {", "if false {")
	output, err = run("linkage-mutant", exec.Command("go", "test", "-overlay", linkOverlay, "-count=1", "-run", "^TestTSGoRequiresLink$", "./bridge/tsgo"))
	if err == nil || !bytes.Contains(output, []byte("lowering accepted an unlinked checker call")) {
		t.Fatalf("linkage mutant escaped refusal test: %v\n%s", err, output)
	}
	t.Log("link opt-in guard removed: refusal test catches it")

	leakOverlay := overlay("leak", "bridge/tsgo/archive/boundary.c", "free(buffer->data);", "/* Mutant: abandon the C output allocation. */")
	leakArchive := filepath.Join(scratch, "leak.a")
	buildArchive("leak-build", leakArchive, leakOverlay, false)
	leakDriver := filepath.Join(scratch, "leak-driver")
	linkDriver("leak-link", leakArchive, leakDriver)
	output, err = run("leak-mutant", exec.Command(leakDriver, config, sample, "check"))
	if err == nil || !bytes.Contains(output, []byte("LeakSanitizer: detected memory leaks")) {
		t.Fatalf("unfreed C output mutant escaped LSan: %v\n%s", err, output)
	}
	t.Log("C output free removed: LeakSanitizer detects the leaked buffers")

	regionFixture := filepath.Join(repository, "bridge/tsgo/testdata/region.a")
	regionArguments := []string{config, sample}
	regionManifest := filepath.Join(scratch, "region.tsv")
	if err := os.WriteFile(regionManifest, []byte(sample+"\t14\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	regionTruth := mustRun("region-oracle", exec.Command(oracle, config, regionManifest, sample))
	// The oracle's timing is on stderr; obtain stdout separately for its frame.
	command := exec.Command(oracle, config, regionManifest, sample)
	command.Dir = repository
	regionTruth, err = command.Output()
	if err != nil {
		t.Fatal(err)
	}
	frame := bytes.SplitN(regionTruth, []byte("\n"), 2)
	var kind, symbolLength, typeLength int
	if len(frame) != 2 {
		t.Fatal("invalid region oracle frame")
	}
	if _, err := fmt.Sscanf(string(frame[0]), "%d %d %d", &kind, &symbolLength, &typeLength); err != nil {
		t.Fatal(err)
	}
	if symbolLength+typeLength > len(frame[1]) {
		t.Fatal("short region oracle frame")
	}
	expected := fmt.Sprintf("%d\n", len(utf16.Encode([]rune(string(frame[1][:symbolLength+typeLength])))))
	healthyRegion := filepath.Join(scratch, "healthy-region")
	mustRun("region-build", exec.Command(stage0, "build", regionFixture, "-o", healthyRegion, "--tsgo", sanitized, "--sanitize", "--count"))
	command = exec.Command(healthyRegion, regionArguments...)
	command.Dir = repository
	var regionReport bytes.Buffer
	command.Stderr = &regionReport
	regionAnswer, err := command.Output()
	if err != nil || string(regionAnswer) != expected || !strings.HasSuffix(regionReport.String(), " regions 1\n") {
		t.Fatalf("region probe did not use one region and agree with Go: %v stdout=%q stderr=%s", err, regionAnswer, regionReport.String())
	}
	t.Logf("region probe: Go and native answer %s; %s", strings.TrimSpace(expected), strings.TrimSpace(regionReport.String()))
	regionOverlay := overlay("region-ownership", "internal/native/runtime/tsgo.c", "adamic_object_new_in(region, &shape)", "adamic_object_new(&shape)")
	regionCompiler := filepath.Join(scratch, "region-adamic")
	mustRun("region-compiler-build", exec.Command("go", "build", "-overlay", regionOverlay, "-o", regionCompiler, "./cmd/adamic"))
	regionBinary := filepath.Join(scratch, "region-native")
	mustRun("region-native-build", exec.Command(regionCompiler, "build", regionFixture, "-o", regionBinary, "--tsgo", normal, "--sanitize"))
	output, err = run("region-ownership-mutant", exec.Command(regionBinary, regionArguments...))
	if err == nil || !bytes.Contains(output, []byte("LeakSanitizer: detected memory leaks")) {
		t.Fatalf("heap result in region variant escaped LSan: %v\n%s", err, output)
	}
	t.Log("region entry allocates on heap: LeakSanitizer catches the unowned result")

}

func firstDifference(left, right []byte) int {
	for index := 0; index < len(left) && index < len(right); index++ {
		if left[index] != right[index] {
			return index
		}
	}
	return min(len(left), len(right))
}

func TestTSGoRequiresLink(t *testing.T) {
	t.Parallel()
	fixture := filepath.Join("testdata", "queries.a")
	loaded, err := load.Load([]string{fixture})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lower.Lower(context.Background(), loaded); err == nil {
		t.Fatal("lowering accepted an unlinked checker call")
	}
	loaded.EnableTSGo()
	lowered, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	if !native.UsesTSGo(lowered) {
		t.Fatal("opted-in bridge calls lost")
	}
}
