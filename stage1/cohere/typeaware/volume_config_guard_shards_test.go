package typeaware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/corpusfiles"
)

const testVolumeConfigGuardAndMutantShards = 16

type volumeGuardCase struct{ key, text string }

func volumeGuardShard(key string) int {
	sum := sha256.Sum256([]byte(key))
	return int(sum[0]) % testVolumeConfigGuardAndMutantShards
}
func volumeGuardUnion(t *testing.T, buckets [][]volumeGuardCase, controls []string) {
	t.Helper()
	if len(buckets) != testVolumeConfigGuardAndMutantShards {
		t.Fatal("shard count differs from enumeration")
	}
	seen := map[string]int{}
	for shard, cases := range buckets {
		for _, c := range cases {
			if (c.key == "strict-this-guard" && shard != 0) || (c.key != "strict-this-guard" && volumeGuardShard(c.key) != shard) {
				t.Fatal("wrong shard")
			}
			seen[c.key]++
		}
	}
	if len(seen) != len(controls)+1 {
		t.Fatal("union differs from live enumeration")
	}
	for _, count := range seen {
		if count != 1 {
			t.Fatal("case repeated")
		}
	}

}
func volumeGuardCorpus(h *volumeGuardHarness, shard int, name, oracle, binary, config, manifest string) {
	data, err := os.ReadFile(manifest)
	if err != nil {
		h.t.Fatal(err)
	}
	var paths []string
	seen := map[string]bool{}
	for _, path := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if path == "" {
			continue
		}
		if seen[path] {
			h.t.Fatalf("duplicate corpus case %s", path)
		}
		seen[path] = true
		if volumeGuardShard(filepath.ToSlash(filepath.Clean(path))) == shard {
			paths = append(paths, path)
		}
	}
	h.t.Logf("%s union=%d shard cases=%d", name, len(seen), len(paths))
	if len(paths) != 0 {
		volumeGuardCompare(h, name+"-final-asan", oracle, binary, config, h.write(name+"-shard.manifest", strings.Join(paths, "\n")+"\n"))
	}
}
func volumeGuardArchive(h *volumeGuardHarness, name, overlay string, sanitize bool) string {
	inputs := volumeGuardInputs(h.ctx, "archive-"+name, sanitize)
	if overlay != "" {
		// Overlay paths are temporary; hash the replacement content instead.
		inputs.Flags = append(inputs.Flags, "NoImplicitThis=StrictNullChecks")
	}
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		builder := &volumeGuardHarness{harness: &harness{t: h.t, repository: h.repository, directory: directory}, ctx: h.ctx}
		args := []string{"build", "-buildmode=c-archive", "-o", filepath.Join(directory, "checker.a")}
		if overlay != "" {
			args = append(args, "-overlay", overlay)
		}
		args = append(args, "./bridge/tsgo/archive")
		cmd := exec.Command("go", args...)
		if sanitize {
			cmd.Env = append(os.Environ(), "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
		}
		volumeGuardMust(builder, "checker", cmd)
		return nil
	})
	return filepath.Join(directory, "checker.a")
}
func volumeGuardInputs(ctx context.Context, name string, sanitize bool) buildcache.Inputs {
	return buildcache.Inputs{Name: "volume-guard-" + name, Files: []string{"stage1", "bridge", "internal", "cmd", "cohere", "go.mod"}, Flags: []string{fmt.Sprint(sanitize), os.Getenv("CC"), os.Getenv("CGO_CFLAGS"), os.Getenv("CGO_LDFLAGS"), os.Getenv("GOFLAGS"), os.Getenv("ADAMIC_TOOLS")}, Toolchain: []string{volumeGuardTool(ctx, "clang", "--version"), volumeGuardTool(ctx, "go", "version")}}
}
func volumeGuardNative(h *volumeGuardHarness, stage0, name, entry, archive string, sanitize bool) string {
	inputs := volumeGuardInputs(h.ctx, name, sanitize)
	// Include the exact archive and compiler bytes, including mutant products.
	for _, path := range []string{stage0, archive} {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("%x", sha256.Sum256(data)))
	}
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		builder := &volumeGuardHarness{harness: &harness{t: h.t, repository: h.repository, directory: directory}, ctx: h.ctx}
		args := []string{"build", entry, "-o", filepath.Join(directory, "native"), "--tsgo", archive}
		if sanitize {
			args = append(args, "--sanitize")
		}
		volumeGuardMust(builder, "native", exec.Command(stage0, args...))
		return nil
	})
	return filepath.Join(directory, "native")
}

type volumeGuardProducts struct{ repository, binary, mutant, oracle, asan string }

// Each harness carries only its own unit's deadline; products are published by Setup.
type volumeGuardHarness struct {
	*harness
	ctx context.Context
}

var volumeGuardShared struct {
	products volumeGuardProducts
	ready    bool
}

// The gate may select a shard without selecting its setup. Include the dedicated
// serial setup test in that invocation before parallel leaves are resumed. This
// runs no builds at package initialization and does not select unrelated tests.
func init() {
	for i, arg := range os.Args {
		value := ""
		equal := strings.HasPrefix(arg, "-test.run=")
		if equal {
			value = strings.TrimPrefix(arg, "-test.run=")
		} else if arg == "-test.run" && i+1 < len(os.Args) {
			value = os.Args[i+1]
		} else {
			continue
		}
		if selected, rewritten := volumeGuardSetupFilter(value); selected {
			if equal {
				os.Args[i] = "-test.run=" + rewritten
			} else {
				os.Args[i+1] = rewritten
			}
		}
		return
	}
}

func volumeGuardSetupFilter(value string) (bool, string) {
	parts := strings.SplitN(value, "/", 2)
	filter, err := regexp.Compile(parts[0])
	if err != nil {
		return false, value
	}
	selected := filter.MatchString("TestVolumeConfigGuardAndMutant") || filter.MatchString("TestVolumeConfigGuardAndMutant_Setup")
	for shard := 0; shard < testVolumeConfigGuardAndMutantShards && !selected; shard++ {
		selected = filter.MatchString(fmt.Sprintf("TestVolumeConfigGuardAndMutant_%03d", shard))
	}
	if !selected {
		return false, value
	}
	parts[0] = "(" + parts[0] + ")|^TestVolumeConfigGuardAndMutant_Setup$"
	return true, strings.Join(parts, "/")
}

// Filtered leaves must schedule setup without broadening unrelated selections.
func TestVolumeConfigGuardSetupFilter(t *testing.T) {
	t.Parallel()
	cases := []struct {
		run   string
		setup bool
	}{
		{"^TestVolumeConfigGuardAndMutant_005$", true},
		{"^TestVolumeConfigGuardAndMutant_Setup$", true},
		{"^TestVolumeConfigGuardAndMutant$", true},
		{"^(TestVolumeConfigGuardAndMutant_000|Other)$/specific$", true},
		{"^TestVolumeConfigGuardAndMutantUnion$", false},
		{"^Other$/specific$", false},
		{"[", false},
	}
	for _, c := range cases {
		selected, rewritten := volumeGuardSetupFilter(c.run)
		if selected != c.setup {
			t.Fatalf("filter %q selected setup=%t", c.run, selected)
		}
		if !selected {
			if rewritten != c.run {
				t.Fatalf("unrelated filter %q changed", c.run)
			}
			continue
		}
		before := strings.SplitN(c.run, "/", 2)
		after := strings.SplitN(rewritten, "/", 2)
		if len(before) != len(after) || (len(before) == 2 && before[1] != after[1]) {
			t.Fatalf("child selection changed: %q -> %q", c.run, rewritten)
		}
		old := regexp.MustCompile(before[0])
		next := regexp.MustCompile(after[0])
		if !next.MatchString("TestVolumeConfigGuardAndMutant_Setup") {
			t.Fatal("setup was not selected")
		}
		for _, name := range []string{"Other", "Another", "TestVolumeConfigGuardAndMutant_000", "TestVolumeConfigGuardAndMutant_005", "TestVolumeConfigGuardAndMutantUnion"} {
			if old.MatchString(name) != next.MatchString(name) {
				t.Fatalf("selection of %s changed for %q", name, c.run)
			}
		}
	}
}

// Not parallel: publishes immutable products before any guard shard resumes.
func TestVolumeConfigGuardAndMutant_Setup(t *testing.T) {
	ctx, cancel := volumeGuardUnitContext(t.Name())
	defer cancel()
	started := time.Now()
	defer func() { t.Logf("setup elapsed_s=%.6f cooked=%t", time.Since(started).Seconds(), ctx.Err() != nil) }()
	volumeGuardShared.products = buildVolumeConfigGuardProducts(t, ctx)
	if ctx.Err() != nil {
		t.Fatal("cooked: shared setup exceeded 90s (over budget)")
	}
	volumeGuardShared.ready = true
}
func volumeGuardProductsFor(t *testing.T) volumeGuardProducts {
	t.Helper()
	if !volumeGuardShared.ready {
		t.Fatal("shared setup did not complete; no shard may build products")
	}
	return volumeGuardShared.products
}
func runVolumeConfigGuardShards(t *testing.T) { volumeGuardProductsFor(t) }
func volumeGuardBuckets() ([][]volumeGuardCase, []string) {
	controls := append(volumeControls(), "declare const console: {log():void};", "const detached=console.log;\nexport {};\n")
	buckets := make([][]volumeGuardCase, testVolumeConfigGuardAndMutantShards)
	for i, text := range controls {
		name := fmt.Sprintf("edge-control-%03d.ts", i)
		if i == len(controls)-2 {
			name = "native-globals.d.ts"
		}
		shard := volumeGuardShard(name)
		buckets[shard] = append(buckets[shard], volumeGuardCase{name, text})
	}
	buckets[0] = append(buckets[0], volumeGuardCase{"strict-this-guard", ""})
	return buckets, controls
}
func TestVolumeConfigGuardAndMutantUnion(t *testing.T) {
	t.Parallel()
	buckets, controls := volumeGuardBuckets()
	volumeGuardUnion(t, buckets, controls)
	caught := 0
	for shard, cases := range buckets {
		for _, c := range cases {
			if c.key == "edge-control-000.ts" && !volumeGuardFindingBytesEqual([]byte(c.text+" planted disagreement"), []byte(c.text)) {
				caught++
				t.Logf("planted failure caught by TestVolumeConfigGuardAndMutant_%03d", shard)
			}
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught %d times", caught)
	}
	t.Logf("union=%d, each exactly once", len(controls)+1)
}
func runVolumeConfigGuardShard(t *testing.T, shard int) {
	t.Helper()
	products := volumeGuardProductsFor(t)
	repository, binary, mutant, oracle, asan := products.repository, products.binary, products.mutant, products.oracle, products.asan
	buckets, _ := volumeGuardBuckets()

	ctx, cancel := volumeGuardUnitContext(t.Name())
	defer cancel()
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		t.Logf("shard-%03d elapsed_s=%.6f cooked=%t", shard, elapsed.Seconds(), ctx.Err() != nil)
		if ctx.Err() != nil {
			t.Error("cooked: shard exceeded 90s (over budget)")
		}
	}()
	h := &volumeGuardHarness{harness: &harness{t: t, repository: repository, directory: t.TempDir()}, ctx: ctx}
	if shard == 0 {
		source := h.write("this.ts", "function f(){const x=this.m;const y:number=this;return x;} export {};\n")
		manifest := h.write("this.manifest", source+"\n")
		config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"noImplicitThis":false,"target":"ES2022","lib":["ES2022"]}}`)
		observed := volumeGuardRun(h, "nonstrict-this", exec.Command(binary, config, manifest))
		message := []byte("adamic: panic: volume suite requires noImplicitThis; implicit-this messages are not yet ported\n")
		if code, ok := observed.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Equal(observed.stderr, message) || len(observed.stdout) != 0 {
			t.Fatalf("unsupported this mode escaped: %v %s", observed.err, observed.stderr)
		}
		observed = volumeGuardMust(h, "strict-this-run", exec.Command(mutant, config, manifest))
		if len(observed.stderr) != 0 || !bytes.Contains(observed.stdout, []byte("findings ")) {
			t.Fatal("strict-this mutant did not finish normally")
		}
		t.Log("strict-this compiler-option mutant: refusal expectation catches exit 0 instead of 70")
		truth := volumeGuardMust(h, "nonstrict-go", exec.Command(oracle, config, manifest))
		if bytes.Equal(observed.stdout, truth.stdout) {
			t.Fatal("implicit-this input did not distinguish the production message variants")
		}
		t.Logf("implicit-this mutant also differs from independent Go at byte %d", firstDifference(observed.stdout, truth.stdout))
		config = h.write("strict.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`)
		volumeGuardCompare(h, "strict-this-control", oracle, asan, config, manifest)

	}
	config := h.write("strict.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`)
	var paths []string
	for _, c := range buckets[shard] {
		if c.key == "strict-this-guard" {
			continue
		}
		text := c.text
		if !strings.HasSuffix(c.key, ".d.ts") {
			text += "\nexport {};\n"
		}
		paths = append(paths, h.write(c.key, text))
	}
	// Every shard needs the ambient globals, even when their case belongs elsewhere.
	if len(paths) != 0 {
		paths = append(paths, h.write("ambient.d.ts", "declare const console: {log():void};\n"))
		truth := volumeGuardCompare(h, "edge-controls-asan", oracle, asan, config, h.write("controls.manifest", strings.Join(paths, "\n")+"\n"))
		for _, c := range buckets[shard] {
			if c.key == "edge-control-000.ts" {
				planted := append(append([]byte(nil), truth.stdout...), []byte("planted disagreement\n")...)
				if volumeGuardFindingBytesEqual(planted, truth.stdout) {
					t.Fatal("planted failure survived")
				}
				t.Logf("planted failure caught by shard-%03d", shard)
			}
		}
	}
	if manifest := os.Getenv("ADAMIC_VOLUME_REPOSITORY_MANIFEST"); manifest != "" {
		volumeGuardCorpus(h, shard, "repository", oracle, asan, filepath.Join(repository, "tsconfig.json"), manifest)
	}
	if corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"); corpus != "" {
		paths := corpusfiles.Upstream(t, corpus, compilerCommit, []string{"src/compiler"}, []string{"*.ts"})
		volumeGuardCorpus(h, shard, "compiler", oracle, asan, filepath.Join(corpus, "src/compiler/tsconfig.json"), h.write("compiler.manifest", strings.Join(paths, "\n")+"\n"))
	}
}
func buildVolumeConfigGuardProducts(t *testing.T, ctx context.Context) volumeGuardProducts {
	started := time.Now()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_VOLUME_GUARD_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &volumeGuardHarness{harness: &harness{t: t, repository: repository, directory: directory}, ctx: ctx}
	stage0 := filepath.Join(directory, "adamic")
	stage0Directory := buildcache.Product(t, volumeGuardInputs(ctx, "stage0", false), func(directory string) error {
		cmd, cancel := volumeGuardCommand(ctx, "go", "build", "-o", filepath.Join(directory, "adamic"), "./cmd/adamic")
		defer cancel()
		cmd.Dir = repository
		output, err := volumeGuardOutput(cmd)
		if err != nil {
			return fmt.Errorf("stage0: %w: %s", err, output)
		}
		return nil
	})
	stage0 = filepath.Join(stage0Directory, "adamic")
	archive := volumeGuardArchive(h, "checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/volume_suite.ts")
	binary := volumeGuardNative(h, stage0, "volume", entry, archive, false)
	overlay := h.overlay("strict-this", "bridge/tsgo/checker/facts.go", "option = p.Compiler.Options().NoImplicitThis", "option = p.Compiler.Options().StrictNullChecks")
	mutantArchive := volumeGuardArchive(h, "strict-this-checker", overlay, false)
	mutant := volumeGuardNative(h, stage0, "strict-this-native", entry, mutantArchive, false)
	oracleDirectory := buildcache.Product(t, volumeGuardInputs(ctx, "oracle", false), func(directory string) error {
		builder := &volumeGuardHarness{harness: &harness{t: t, repository: repository, directory: directory}, ctx: ctx}
		virtual := filepath.Join(repository, "cohere/adamic_volume-oracle.go")
		data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/cohere/typeaware/testdata/oracle_volume.go")}})
		if err != nil {
			return err
		}
		cmd := exec.Command("go", "build", "-overlay", builder.write("volume-oracle-overlay.json", string(data)), "-o", filepath.Join(directory, "volume-oracle"), virtual)
		cmd.Dir = filepath.Join(repository, "cohere")
		volumeGuardMust(builder, "volume-oracle-build", cmd)
		return nil
	})
	oracle := filepath.Join(oracleDirectory, "volume-oracle")
	sanitized := volumeGuardArchive(h, "checker-asan", "", true)
	asan := volumeGuardNative(h, stage0, "volume-asan", entry, sanitized, true)
	t.Logf("TestVolumeConfigGuardAndMutant (setup): %.3fs", time.Since(started).Seconds())
	return volumeGuardProducts{repository, binary, mutant, oracle, asan}
}

// The planted disagreement exercises the same exact byte comparison as every oracle check.
func volumeGuardFindingBytesEqual(got, want []byte) bool { return bytes.Equal(got, want) }
func volumeGuardCompare(h *volumeGuardHarness, name, oracle, binary, config, manifest string) result {
	h.t.Helper()
	want := volumeGuardMust(h, name+"-go", exec.Command(oracle, config, manifest))
	got := volumeGuardMust(h, name+"-native", exec.Command(binary, config, manifest))
	if len(got.stderr) != 0 {
		h.t.Fatalf("sanitizer stderr: %s", got.stderr)
	}
	if !volumeGuardFindingBytesEqual(got.stdout, want.stdout) {
		h.t.Fatalf("%s mismatch byte %d", name, firstDifference(got.stdout, want.stdout))
	}
	h.t.Logf("%s: %d identical finding bytes; %s", name, len(want.stdout), summary(want.stdout))
	return want
}

func TestVolumeConfigGuardAndMutant_000(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 0)
}
func TestVolumeConfigGuardAndMutant_001(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 1)
}
func TestVolumeConfigGuardAndMutant_002(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 2)
}
func TestVolumeConfigGuardAndMutant_003(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 3)
}
func TestVolumeConfigGuardAndMutant_004(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 4)
}
func TestVolumeConfigGuardAndMutant_005(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 5)
}
func TestVolumeConfigGuardAndMutant_006(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 6)
}
func TestVolumeConfigGuardAndMutant_007(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 7)
}
func TestVolumeConfigGuardAndMutant_008(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 8)
}
func TestVolumeConfigGuardAndMutant_009(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 9)
}
func TestVolumeConfigGuardAndMutant_010(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 10)
}
func TestVolumeConfigGuardAndMutant_011(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 11)
}
func TestVolumeConfigGuardAndMutant_012(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 12)
}
func TestVolumeConfigGuardAndMutant_013(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 13)
}
func TestVolumeConfigGuardAndMutant_014(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 14)
}
func TestVolumeConfigGuardAndMutant_015(t *testing.T) {
	t.Parallel()
	runVolumeConfigGuardShard(t, 15)
}

// Bound both the driver and compilers it launches without an external deadline tool.
func volumeGuardCommand(parent context.Context, name string, args ...string) (*exec.Cmd, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = time.Second
	return cmd, cancel
}
func volumeGuardRun(h *volumeGuardHarness, name string, original *exec.Cmd) result {
	h.t.Helper()
	cmd, cancel := volumeGuardCommand(h.ctx, original.Path, original.Args[1:]...)
	defer cancel()
	cmd.Dir, cmd.Env, cmd.Stdin = original.Dir, original.Env, original.Stdin
	h.next++
	if cmd.Dir == "" {
		cmd.Dir = h.repository
	}
	stem := filepath.Join(h.directory, fmt.Sprintf("%03d-%s", h.next, strings.ReplaceAll(name, "/", "-")))
	out, err := os.Create(stem + ".stdout")
	if err != nil {
		h.t.Fatal(err)
	}
	defer out.Close()
	report, err := os.Create(stem + ".stderr")
	if err != nil {
		h.t.Fatal(err)
	}
	defer report.Close()
	cmd.Stdout, cmd.Stderr = out, report
	started := time.Now()
	runErr := volumeGuardExecute(cmd)
	elapsed := time.Since(started)
	stdout, err := os.ReadFile(out.Name())
	if err != nil {
		h.t.Fatal(err)
	}
	stderr, err := os.ReadFile(report.Name())
	if err != nil {
		h.t.Fatal(err)
	}
	observed := result{stdout: stdout, stderr: stderr, elapsed: elapsed, err: runErr}
	if h.ctx.Err() != nil || observed.elapsed >= 90*time.Second {
		h.t.Fatalf("%s cooked: child deadline exceeded (over budget)", name)
	}
	return observed
}
func volumeGuardMust(h *volumeGuardHarness, name string, cmd *exec.Cmd) result {
	h.t.Helper()
	observed := volumeGuardRun(h, name, cmd)
	if observed.err != nil {
		h.t.Fatalf("%s: %v\n%s\n%s", name, observed.err, observed.stdout, observed.stderr)
	}
	return observed
}
func volumeGuardTool(parent context.Context, name string, args ...string) string {
	cmd, cancel := volumeGuardCommand(parent, name, args...)
	defer cancel()
	output, err := volumeGuardOutput(cmd)
	report := strings.Join(append([]string{name}, args...), " ") + ": " + strings.TrimSpace(string(output))
	if err != nil {
		report += " (" + err.Error() + ")"
	}
	return report
}

var volumeGuardCommandGroups sync.Map

func volumeGuardExecute(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	volumeGuardCommandGroups.Store(cmd.Process.Pid, struct{}{})
	defer volumeGuardCommandGroups.Delete(cmd.Process.Pid)
	return cmd.Wait()
}
func volumeGuardOutput(cmd *exec.Cmd) ([]byte, error) {
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := volumeGuardExecute(cmd)
	return output.Bytes(), err
}

// Cover time spent hashing cache inputs or waiting on a product lock as well as
// child commands. A cooked unit kills all this split's active process groups.
func volumeGuardUnitContext(name string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	stop := context.AfterFunc(ctx, func() {
		if ctx.Err() != context.DeadlineExceeded {
			return
		}
		volumeGuardCommandGroups.Range(func(pid, value any) bool {
			_ = syscall.Kill(-pid.(int), syscall.SIGKILL)
			return true
		})
		fmt.Fprintf(os.Stderr, "cooked: %s exceeded 90s (over budget)\n", name)
		os.Exit(124)
	})
	return ctx, func() { stop(); cancel() }
}
