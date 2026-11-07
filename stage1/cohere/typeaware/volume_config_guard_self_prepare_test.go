package typeaware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

const testVolumeConfigGuardAndMutantShards = 16

type volumeGuardCase struct{ key, text string }

func volumeGuardShard(key string) int {
	sum := sha256.Sum256([]byte(key))
	return int(sum[0]) % testVolumeConfigGuardAndMutantShards
}
func volumeGuardUnion(test *testing.T, buckets [][]volumeGuardCase, controls []string) {
	test.Helper()
	if len(buckets) != testVolumeConfigGuardAndMutantShards {
		test.Fatal("shard count differs from enumeration")
	}
	seen := map[string]int{}
	for shard, cases := range buckets {
		for _, guardCase := range cases {
			if (guardCase.key == "strict-this-guard" && shard != 0) ||
				(guardCase.key != "strict-this-guard" && volumeGuardShard(guardCase.key) != shard) {
				test.Fatal("wrong shard")
			}
			seen[guardCase.key]++
		}
	}
	if len(seen) != len(controls)+1 {
		test.Fatal("union differs from live enumeration")
	}
	for _, count := range seen {
		if count != 1 {
			test.Fatal("case repeated")
		}
	}

}

func volumeGuardInputs(callerContext context.Context, name string, sanitize bool) buildcache.Inputs {
	return buildcache.Inputs{
		Name:  "volume-guard-" + name,
		Files: []string{"stage1", "bridge", "internal", "cmd", "cohere", "go.mod"},
		Flags: append(
			[]string{fmt.Sprint(sanitize), "go build " + strings.Join(buildcache.ReproducibleGoFlags(), " ")},
			typeAwareBuildSettings("CC", "CGO_CFLAGS", "CGO_LDFLAGS", "GOFLAGS")...,
		),
		Toolchain: []string{
			buildcache.Tool("clang", "--version"),
			buildcache.Tool("go", "version"),
			typeAwareCompilerReport(typeAwareBuildSetting("CC")),
		},
	}
}
func volumeGuardNative(guardHarness *volumeGuardHarness, stage0, name, entry, archive string, sanitize bool) string {
	inputs := volumeGuardInputs(guardHarness.ctx, name, sanitize)
	// Include the exact archive and compiler bytes, including mutant products.
	for _, path := range []string{stage0, archive} {
		data, readError := os.ReadFile(path)
		if readError != nil {
			guardHarness.t.Fatal(readError)
		}
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("%x", sha256.Sum256(data)))
	}
	directory := buildcache.Product(guardHarness.t, inputs, func(directory string) error {
		builder := &volumeGuardHarness{
			harness: &harness{t: guardHarness.t, repository: guardHarness.repository, directory: directory},
			ctx:     guardHarness.ctx,
		}
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

// Each harness carries its shard's deadline, after shared preparation completes.
type volumeGuardHarness struct {
	*harness
	ctx context.Context
}

var volumeGuardShared struct {
	once     sync.Once
	products volumeGuardProducts
}

func TestVolumeConfigGuardAndMutant_Setup(test *testing.T) {
	test.Parallel()
	volumeGuardProductsFor(test)
}

func volumeGuardProductsFor(test *testing.T) volumeGuardProducts {
	test.Helper()
	volumeGuardShared.once.Do(func() {
		volumeGuardShared.products = buildVolumeConfigGuardProducts(test)
	})
	if volumeGuardShared.products.repository == "" {
		test.Fatal("shared product preparation failed")
	}
	return volumeGuardShared.products
}
func runVolumeConfigGuardShards(test *testing.T) { volumeGuardProductsFor(test) }
func volumeGuardBuckets() ([][]volumeGuardCase, []string) {
	controls := append(
		volumeControls(),
		"declare const console: {log():void};",
		"const detached=console.log;\nexport {};\n",
	)
	buckets := make([][]volumeGuardCase, testVolumeConfigGuardAndMutantShards)
	for index, text := range controls {
		name := fmt.Sprintf("edge-control-%03d.ts", index)
		if index == len(controls)-2 {
			name = "native-globals.d.ts"
		}
		shard := volumeGuardShard(name)
		buckets[shard] = append(buckets[shard], volumeGuardCase{name, text})
	}
	buckets[0] = append(buckets[0], volumeGuardCase{"strict-this-guard", ""})
	return buckets, controls
}
func TestVolumeConfigGuardAndMutantUnion(test *testing.T) {
	test.Parallel()
	buckets, controls := volumeGuardBuckets()
	volumeGuardUnion(test, buckets, controls)
	caught := 0
	for shard, cases := range buckets {
		for _, guardCase := range cases {
			if guardCase.key == "edge-control-000.ts" &&
				!volumeGuardFindingBytesEqual([]byte(guardCase.text+" planted disagreement"), []byte(guardCase.text)) {
				caught++
				test.Logf("planted failure caught by TestVolumeConfigGuardAndMutant_%03d", shard)
			}
		}
	}
	if caught != 1 {
		test.Fatalf("planted disagreement caught %d times", caught)
	}
	test.Logf("union=%d, each exactly once", len(controls)+1)
}
func runVolumeConfigGuardShard(test *testing.T, shard int) {
	test.Helper()
	products := volumeGuardProductsFor(test)
	if shard == 0 {
		products = volumeGuardMutantProductsFor(test, products)
	}
	repository, binary, mutant, oracle, asan := products.repository, products.binary, products.mutant, products.oracle, products.asan
	buckets, _ := volumeGuardBuckets()

	ctx, cancel := volumeGuardUnitContext(test.Name())
	defer cancel()
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		test.Logf("shard-%03d elapsed_s=%.6f cooked=%t", shard, elapsed.Seconds(), ctx.Err() != nil)
		if ctx.Err() != nil {
			test.Error("cooked: shard exceeded 90s (over budget)")
		}
	}()
	harness := &volumeGuardHarness{
		harness: &harness{t: test, repository: repository, directory: test.TempDir()},
		ctx:     ctx,
	}
	if shard == 0 {
		source := harness.write("this.ts", "function f(){const x=this.m;const y:number=this;return x;} export {};\n")
		manifest := harness.write("this.manifest", source+"\n")
		configuration := harness.write(
			"tsconfig.json",
			`{"compilerOptions":{"strict":true,"noImplicitThis":false,"target":"ES2022","lib":["ES2022"]}}`,
		)
		observed := volumeGuardRun(harness, "nonstrict-this", exec.Command(binary, configuration, manifest))
		message := []byte(
			"adamic: panic: volume suite requires noImplicitThis; implicit-this messages are not yet ported\n",
		)
		if code, ok := observed.err.(*exec.ExitError); !ok ||
			code.ExitCode() != 70 ||
			!bytes.Equal(observed.stderr, message) ||
			len(observed.stdout) != 0 {
			test.Fatalf("unsupported this mode escaped: %v %s", observed.err, observed.stderr)
		}
		observed = volumeGuardMust(harness, "strict-this-run", exec.Command(mutant, configuration, manifest))
		if len(observed.stderr) != 0 || !bytes.Contains(observed.stdout, []byte("findings ")) {
			test.Fatal("strict-this mutant did not finish normally")
		}
		test.Log("strict-this compiler-option mutant: refusal expectation catches exit 0 instead of 70")
		truth := volumeGuardMust(harness, "nonstrict-go", exec.Command(oracle, configuration, manifest))
		if bytes.Equal(observed.stdout, truth.stdout) {
			test.Fatal("implicit-this input did not distinguish the production message variants")
		}
		test.Logf(
			"implicit-this mutant also differs from independent Go at byte %d",
			firstDifference(observed.stdout, truth.stdout),
		)
		configuration = harness.write(
			"strict.json",
			`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`,
		)
		volumeGuardCompare(harness, "strict-this-control", oracle, asan, configuration, manifest)

	}
	configuration := harness.write(
		"strict.json",
		`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`,
	)
	var paths []string
	for _, guardCase := range buckets[shard] {
		if guardCase.key == "strict-this-guard" {
			continue
		}
		text := guardCase.text
		if !strings.HasSuffix(guardCase.key, ".d.ts") {
			text += "\nexport {};\n"
		}
		paths = append(paths, harness.write(guardCase.key, text))
	}
	// Every shard needs the ambient globals, even when their case belongs elsewhere.
	if len(paths) != 0 {
		paths = append(paths, harness.write("ambient.d.ts", "declare const console: {log():void};\n"))
		truth := volumeGuardCompare(
			harness,
			"edge-controls-asan",
			oracle,
			asan,
			configuration,
			harness.write("controls.manifest", strings.Join(paths, "\n")+"\n"),
		)
		for _, guardCase := range buckets[shard] {
			if guardCase.key == "edge-control-000.ts" {
				planted := append(append([]byte(nil), truth.stdout...), []byte("planted disagreement\n")...)
				if volumeGuardFindingBytesEqual(planted, truth.stdout) {
					test.Fatal("planted failure survived")
				}
				test.Logf("planted failure caught by shard-%03d", shard)
			}
		}
	}
}

func volumeGuardPreparationHarness(test *testing.T) *volumeGuardHarness {
	test.Helper()
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return &volumeGuardHarness{
		harness: &harness{t: test, repository: repository, directory: test.TempDir()},
		ctx:     context.Background(),
	}
}

// Product units and shards share this dispatcher, including all dependency recipes.
func volumeGuardProduct(harness *volumeGuardHarness, name string) string {
	harness.t.Helper()
	switch name {
	case "stage0":
		directory := buildcache.Product(harness.t, volumeGuardInputs(
			harness.ctx,
			name,
			false,
		), func(directory string) error {
			command, cancel := volumeGuardCommand(harness.ctx, "go", append(
				append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
				"-o",
				filepath.Join(directory, "adamic"),
				"./cmd/adamic",
			)...)
			defer cancel()
			command.Dir = harness.repository
			command.Env = append(command.Env, "PWD="+harness.repository)
			output, volumeGuardOutputError := volumeGuardOutput(command)
			if volumeGuardOutputError != nil {
				return fmt.Errorf("stage0: %w: %s", volumeGuardOutputError, output)
			}
			return nil
		})
		return filepath.Join(directory, "adamic")
	case "oracle":
		// The volume oracle and the checker archives are this package's own requests (volumeOracle, h.archive), so the
		// guard's and the package's checks read one product each.
		return volumeOracle(harness.harness, "volume-oracle", "oracle_volume.go")
	case "archive-checker":
		return harness.archive(false)
	case "archive-checker-asan":
		return harness.archive(true)
	case "archive-strict-this-checker":
		return plantedCheckerArchive(harness.t, "strict-this-checker.a", buildcache.Mutation{
			File:   checkerFacts,
			Before: "option = p.Compiler.Options().NoImplicitThis",
			After:  "option = p.Compiler.Options().StrictNullChecks",
		})
	case "volume", "volume-asan", "strict-this-native":
		stage0 := volumeGuardProduct(harness, "stage0")
		archiveName := "archive-checker"
		if name == "volume-asan" {
			archiveName = "archive-checker-asan"
		}
		if name == "strict-this-native" {
			archiveName = "archive-strict-this-checker"
		}
		archive := volumeGuardProduct(harness, archiveName)
		entry := filepath.Join(harness.repository, "stage1/cohere/typeaware/volume_suite.ts")
		return volumeGuardNative(harness, stage0, name, entry, archive, name == "volume-asan")
	default:
		harness.t.Fatalf("unknown volume guard product %q", name)
		return ""
	}
}

func buildVolumeConfigGuardProducts(test *testing.T) volumeGuardProducts {
	harness := volumeGuardPreparationHarness(test)
	return volumeGuardProducts{
		repository: harness.repository,
		oracle:     volumeGuardProduct(harness, "oracle"),
		asan:       volumeGuardProduct(harness, "volume-asan"),
	}
}

// Only the strict-this shard needs the unsanitized runner and its mutant.
var volumeGuardMutantShared struct {
	once     sync.Once
	products volumeGuardProducts
}

func volumeGuardMutantProductsFor(test *testing.T, products volumeGuardProducts) volumeGuardProducts {
	test.Helper()
	volumeGuardMutantShared.once.Do(func() {
		harness := volumeGuardPreparationHarness(test)
		products.binary = volumeGuardProduct(harness, "volume")
		products.mutant = volumeGuardProduct(harness, "strict-this-native")
		volumeGuardMutantShared.products = products
	})
	if volumeGuardMutantShared.products.mutant == "" {
		test.Fatal("shared mutant preparation failed")
	}
	return volumeGuardMutantShared.products
}

// The planted disagreement exercises the same exact byte comparison as every oracle check.
func volumeGuardFindingBytesEqual(got, want []byte) bool { return bytes.Equal(got, want) }
func volumeGuardCompare(harness *volumeGuardHarness, name, oracle, binary, configuration, manifest string) result {
	harness.t.Helper()
	want := volumeGuardMust(harness, name+"-go", exec.Command(oracle, configuration, manifest))
	got := volumeGuardMust(harness, name+"-native", exec.Command(binary, configuration, manifest))
	if len(got.stderr) != 0 {
		harness.t.Fatalf("sanitizer stderr: %s", got.stderr)
	}
	if !volumeGuardFindingBytesEqual(got.stdout, want.stdout) {
		harness.t.Fatalf("%s mismatch byte %d", name, firstDifference(got.stdout, want.stdout))
	}
	harness.t.Logf("%s: %d identical finding bytes; %s", name, len(want.stdout), summary(want.stdout))
	return want
}

func TestVolumeConfigGuardAndMutant_000(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 0)
}
func TestVolumeConfigGuardAndMutant_001(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 1)
}
func TestVolumeConfigGuardAndMutant_002(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 2)
}
func TestVolumeConfigGuardAndMutant_003(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 3)
}
func TestVolumeConfigGuardAndMutant_004(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 4)
}
func TestVolumeConfigGuardAndMutant_005(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 5)
}
func TestVolumeConfigGuardAndMutant_006(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 6)
}
func TestVolumeConfigGuardAndMutant_007(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 7)
}
func TestVolumeConfigGuardAndMutant_008(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 8)
}
func TestVolumeConfigGuardAndMutant_009(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 9)
}
func TestVolumeConfigGuardAndMutant_010(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 10)
}
func TestVolumeConfigGuardAndMutant_011(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 11)
}
func TestVolumeConfigGuardAndMutant_012(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 12)
}
func TestVolumeConfigGuardAndMutant_013(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 13)
}
func TestVolumeConfigGuardAndMutant_014(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 14)
}
func TestVolumeConfigGuardAndMutant_015(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, 15)
}

// Bound both the driver and compilers it launches without an external deadline tool.
func volumeGuardCommand(parent context.Context, name string, args ...string) (*exec.Cmd, context.CancelFunc) {
	cancelContext, cancel := context.WithCancel(parent)
	command := exec.CommandContext(cancelContext, name, args...)
	// A tool named bare (go) is toolchain's program, under toolchain's environment; a binary named by its path runs as
	// itself.
	if !strings.ContainsRune(name, filepath.Separator) {
		tool := toolchain.Command(name, args...)
		command.Path, command.Err, command.Env = tool.Path, tool.Err, tool.Env
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if killError == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return killError
	}
	// No WaitDelay: exec reads volumeGuardOutput's buffer to its end, where one closed it a second after go exited and
	// failed a build that had finished but was slow to drain under load (#n3gk22h's shape). Not childguard.Capture,
	// which has no Started: volumeGuardUnitContext's cook kills the groups volumeGuardExecute registers before it exits
	// the process, and a group it couldn't register would outlive the exit. go waits for the compilers it starts before
	// it exits, so nothing holds a pipe after it, and the cook or the unit's ceiling ends a wait on one that did.
	return command, cancel
}
func volumeGuardRun(harness *volumeGuardHarness, name string, original *exec.Cmd) result {
	harness.t.Helper()
	command, cancel := volumeGuardCommand(harness.ctx, original.Path, original.Args[1:]...)
	defer cancel()
	command.Dir, command.Env, command.Stdin = original.Dir, original.Env, original.Stdin
	harness.next++
	if command.Dir == "" {
		command.Dir = harness.repository
	}
	if command.Env != nil {
		// exec.Cmd sets PWD only for a command with no Env of its own; toolchain's have one.
		command.Env = append(command.Env, "PWD="+command.Dir)
	}
	stem := filepath.Join(harness.directory, fmt.Sprintf("%03d-%s", harness.next, strings.ReplaceAll(name, "/", "-")))
	out, createError := os.Create(stem + ".stdout")
	if createError != nil {
		harness.t.Fatal(createError)
	}
	defer out.Close()
	report, createError := os.Create(stem + ".stderr")
	if createError != nil {
		harness.t.Fatal(createError)
	}
	defer report.Close()
	command.Stdout, command.Stderr = out, report
	started := time.Now()
	runError := volumeGuardExecute(command)
	elapsed := time.Since(started)
	stdout, readError := os.ReadFile(out.Name())
	if readError != nil {
		harness.t.Fatal(readError)
	}
	stderr, readError := os.ReadFile(report.Name())
	if readError != nil {
		harness.t.Fatal(readError)
	}
	observed := result{stdout: stdout, stderr: stderr, elapsed: elapsed, err: runError}
	if harness.ctx.Err() != nil {
		harness.t.Fatalf("%s cooked: child deadline exceeded (over budget)", name)
	}
	return observed
}
func volumeGuardMust(harness *volumeGuardHarness, name string, command *exec.Cmd) result {
	harness.t.Helper()
	observed := volumeGuardRun(harness, name, command)
	if observed.err != nil {
		harness.t.Fatalf("%s: %v\n%s\n%s", name, observed.err, observed.stdout, observed.stderr)
	}
	return observed
}

var volumeGuardCommandGroups sync.Map

func volumeGuardExecute(command *exec.Cmd) error {
	if startError := command.Start(); startError != nil {
		return startError
	}
	volumeGuardCommandGroups.Store(command.Process.Pid, struct{}{})
	defer volumeGuardCommandGroups.Delete(command.Process.Pid)
	return command.Wait()
}
func volumeGuardOutput(command *exec.Cmd) ([]byte, error) {
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	volumeGuardExecuteError := volumeGuardExecute(command)
	return output.Bytes(), volumeGuardExecuteError
}

// Cover time spent hashing cache inputs or waiting on a product lock as well as
// child commands. A cooked unit kills all this split's active process groups.
func volumeGuardUnitContext(name string) (context.Context, context.CancelFunc) {
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	stop := context.AfterFunc(timeoutContext, func() {
		if timeoutContext.Err() != context.DeadlineExceeded {
			return
		}
		volumeGuardCommandGroups.Range(func(pid, value any) bool {
			_ = syscall.Kill(-pid.(int), syscall.SIGKILL)
			return true
		})
		fmt.Fprintf(os.Stderr, "cooked: %s exceeded 90s (over budget)\n", name)
		os.Exit(124)
	})
	return timeoutContext, func() { stop(); cancel() }
}

// Exercise the holding shard's planted disagreement against real oracle bytes.
func TestVolumeConfigGuardAndMutantPlantedFailure(test *testing.T) {
	test.Parallel()
	runVolumeConfigGuardShard(test, volumeGuardShard("edge-control-000.ts"))
}
