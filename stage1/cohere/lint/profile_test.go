package lint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"github.com/system-inc/adamic/internal/native"
)

// Opt-in artifacts outlive t.TempDir so callgrind and the timing runner use the same snapshot.
// Not parallel: TestProfileSnapshotsAgree reads what it saves, and a serial test finishes first.
func TestProfileArtifacts(t *testing.T) {
	directory := os.Getenv("ADAMIC_LINT_PROFILE_DIR")
	if directory == "" {
		t.Skip("set ADAMIC_LINT_PROFILE_DIR to a scratch directory")
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	sourceRoot := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if sourceRoot == "" {
		t.Fatal("missing pinned compiler checkout")
	}
	pin := execute(t, sourceRoot, "git", "rev-parse", "HEAD")
	if strings.TrimSpace(string(pin.output)) != compilerCommit {
		t.Fatal("compiler checkout has the wrong pin")
	}
	copyPort(t, directory, "", "")
	prepareRegistry(t, directory)
	var files []string
	err := filepath.WalkDir(filepath.Join(sourceRoot, "src/compiler"), func(path string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && strings.HasSuffix(path, ".ts") {
			files = append(files, path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	var manifest strings.Builder
	for _, path := range files {
		manifest.WriteString(path + "\n")
	}
	if err := os.WriteFile(filepath.Join(directory, "compiler.txt"), []byte(manifest.String()), 0644); err != nil {
		t.Fatal(err)
	}
	oracle, err := os.ReadFile(goOracle(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "oracle"), oracle, 0755); err != nil {
		t.Fatal(err)
	}
	pilot := filepath.Join(directory, "pilot.ts")
	witness, err := os.ReadFile(filepath.Join(directory, "rules/no-unnecessary-boolean-literal-compare/testdata/witness.ts.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pilot, witness, 0644); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(directory, "pilot.tsconfig.json")
	options := fmt.Sprintf(`{"compilerOptions":{"strict":true},"files":[%q]}`, pilot)
	if err := os.WriteFile(config, []byte(options), 0644); err != nil {
		t.Fatal(err)
	}
	pilotManifest := "program " + config + "\n" + pilot + "\t@typescript-eslint/no-unnecessary-boolean-literal-compare\n"
	if err := os.WriteFile(filepath.Join(directory, "pilot.txt"), []byte(pilotManifest), 0644); err != nil {
		t.Fatal(err)
	}
	buildProfile(t, directory)
}

func buildProfile(t *testing.T, directory string) {
	t.Helper()
	started := time.Now()
	built := checkerCompile(t, directory)
	t.Logf("profile lowered program build: %s", time.Since(started))
	source := built.c
	started = time.Now()
	archive := checkerArchive(t, false)
	t.Logf("profile release checker archive build: %s", time.Since(started))
	if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	for _, build := range []struct {
		name    string
		options native.Options
	}{{"scanner", native.Options{}}, {"counted", native.Options{Count: true}}} {
		started = time.Now()
		if err := nativeBuild(func() error {
			return buildCheckerWithRuntime(source, filepath.Join(directory, build.name), archive, build.options)
		}); err != nil {
			t.Fatal(err)
		}
		t.Logf("profile %s build: %s", build.name, time.Since(started))
	}
	started = time.Now()
	runtime := filepath.Join(repository, "internal/native/runtime")
	entries, err := os.ReadDir(runtime)
	if err != nil {
		t.Fatal(err)
	}
	units := []string{filepath.Join(directory, "main.c")}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(runtime, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(directory, entry.Name())
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(entry.Name(), ".c") {
			units = append(units, path)
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "tsgo.h"), bridge.Header, 0644); err != nil {
		t.Fatal(err)
	}
	flags := append(native.Flags(native.Options{}), "-DADAMIC_TSGO", "-g", "-o", filepath.Join(directory, "profiled"))
	flags = append(flags, units...)
	flags = append(flags, archive, "-lm", "-lpthread", "-ldl")
	if err := nativeBuild(func() error {
		_, err := run("", nil, "clang", flags...)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	t.Logf("profile profiled build: %s", time.Since(started))
	t.Logf("release, counted and -O2 -g profiling builds saved in %s", directory)
}

const testProfileCompilationShards = 1

// Exercise the shared profile graph without requiring the external compiler corpus.
// ADAMIC_TEST_SHARD=i/n selects a shard; unset runs all. The single witness
// fits one shard with its build products prepared once before parallel work.
func TestProfileCompilation(t *testing.T) {
	t.Parallel()
	setup := time.Now()
	directory := t.TempDir()
	copyPort(t, directory, "", "")
	prepareRegistry(t, directory)
	buildStart := time.Now()
	buildProfile(t, directory)
	started := time.Now()
	oracle := goOracle(t)
	t.Logf("profile Go oracle build: %s", time.Since(started))
	module := emittedJavaScript(t, directory)
	buildTime := time.Since(buildStart)
	rows := []string{ownedWitnesses(t, directory, "no-var")[0] + "\tno-var"}
	assignments := profileAssignments(t, rows, testProfileCompilationShards)
	t.Logf("setup with builds=%s without builds=%s", time.Since(setup), time.Since(setup)-buildTime)
	for i, cases := range assignments {
		if !profileSelected(t, i, len(assignments)) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			path := manifest(t, cases)
			compareWithJavaScript(t, oracle, filepath.Join(directory, "scanner"), directory, path, module)
			want := execute(t, "", oracle, "--manifest", path).output
			got := execute(t, "", filepath.Join(directory, "profiled"), "--manifest", path).output
			profileCheck(t, got, want)
			want = execute(t, "", oracle, "--manifest", path, "--count").output
			output, err := os.CreateTemp(t.TempDir(), "counted-output-")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			stats, err := os.CreateTemp(t.TempDir(), "counted-stats-")
			if err != nil {
				t.Fatal(err)
			}
			defer stats.Close()
			command := exec.Command(filepath.Join(directory, "counted"), "--manifest", path, "--count")
			command.Stdout, command.Stderr = output, stats
			if err := command.Run(); err != nil {
				t.Fatal(err)
			}
			got, err = os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			counters, err := os.ReadFile(stats.Name())
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(counters), "adamic: counts: allocations ") {
				t.Fatalf("unexpected counted stderr: %s", counters)
			}
			t.Logf("counted instrumentation: %s", counters)
			profileCheck(t, got, want)
		})
	}
}

// It reads the snapshots TestProfileArtifacts saves, when both are pointed at one directory. That test is
// serial, so it has finished before this one is released.
func TestProfileSnapshotsAgree(t *testing.T) {
	t.Parallel()
	skipWhenRuleScoped(t)
	asked := os.Getenv("ADAMIC_LINT_PROFILE_SNAPSHOTS")
	if asked == "" {
		t.Skip("set ADAMIC_LINT_PROFILE_SNAPSHOTS")
	}
	rows := generated(t)
	var recovery []string
	for _, row := range upstream(t) {
		if strings.HasSuffix(row, "\tunsupported-recovery") {
			recovery = append(recovery, row)
		} else {
			rows = append(rows, row)
		}
	}
	for _, root := range []string{filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler"), filepath.Join(repository, "stage1")} {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
				absolute, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				rows = append(rows, absolute)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	oracle := goOracle(t)
	// Classify recovery rows as TestRulesAgree does, so a captured case Go parses with a diagnostic (an
	// octal escape in strict mode, say) is compared in recovery mode rather than panicking the oracle.
	path := manifest(t, recoveryRows(t, oracle, rows))
	want := execute(t, "", oracle, "--manifest", path).output
	for _, directory := range filepath.SplitList(asked) {
		// Match the ordinary suite's explicit recovery boundary in both snapshots.
		for _, row := range recovery {
			checkRecoveryRefusal(t, oracle, filepath.Join(directory, "scanner"), directory, row)
			checkRecoveryRefusal(t, oracle, filepath.Join(directory, "profiled"), directory, row)
		}
		for _, side := range []struct {
			name string
			run  execution
		}{
			{"release", execute(t, "", filepath.Join(directory, "scanner"), "--manifest", path)},
			{"profiled", execute(t, "", filepath.Join(directory, "profiled"), "--manifest", path)},
			{"Node", node(t, directory, path, false)},
			{"emitted JavaScript", emittedNode(t, directory, path, false)},
		} {
			if diff := difference(side.run.output, want); diff != "" {
				t.Fatalf("%s %s: %s", directory, side.name, diff)
			}
		}
		t.Logf("%s: release, profiled and Node identical to Go: %d bytes", directory, len(want))
	}
}
func TestCommentFoldMutant(t *testing.T) {
	t.Parallel()
	path := manifest(t, generated(t))
	want := execute(t, "", goOracle(t), "--manifest", path).output
	directory := mutant(t, "point - 32 : point", "point - 31 : point", "unicode.ts")
	binary := buildPort(t, directory, true)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"emitted JavaScript", emittedNode(t, directory, path, false)},
		{"native", execute(t, "", binary, "--manifest", path)},
	} {
		if diff := difference(side.run.output, want); diff == "" {
			t.Fatalf("fold mutant survived on %s", side.name)
		} else {
			t.Logf("fold mutant caught on %s: %s", side.name, diff)
		}
	}
}

const testPositionIndexMutantShards = 1

// ADAMIC_TEST_SHARD=i/n selects a shard; unset runs all. All generated cases
// form one corpus range, preserving the original existential mutant check on
// every runtime. Build products are prepared once before parallel work.
func TestPositionIndexMutant(t *testing.T) {
	t.Parallel()
	setup := time.Now()
	rows := generated(t)
	directory := mutant(t, "this.anchors[0] = true;", "this.anchors[0] = false;", "rules/no-warning-comments/rule.ts")
	buildStart := time.Now()
	started := time.Now()
	oracle := goOracle(t)
	t.Logf("position Go oracle build: %s", time.Since(started))
	started = time.Now()
	built := checkerCompile(t, directory)
	t.Logf("position lowered program build: %s", time.Since(started))
	started = time.Now()
	archive := ""
	if built.bridge {
		archive = checkerArchive(t, true)
	}
	t.Logf("position sanitized checker archive build: %s", time.Since(started))
	started = time.Now()
	binary := checkerBinary(t, built.c, archive, true)
	t.Logf("position sanitized native build: %s", time.Since(started))
	started = time.Now()
	module := emittedJavaScript(t, directory)
	t.Logf("position emitted JavaScript product: %s", time.Since(started))
	buildTime := time.Since(buildStart)
	assignments := profileAssignments(t, rows, testPositionIndexMutantShards)
	t.Logf("setup with builds=%s without builds=%s", time.Since(setup), time.Since(setup)-buildTime)
	for i, cases := range assignments {
		if !profileSelected(t, i, len(assignments)) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			path := manifest(t, cases)
			want := execute(t, "", oracle, "--manifest", path).output
			for _, side := range []struct {
				name string
				run  execution
			}{
				{"Node", node(t, directory, path, false)},
				{"emitted JavaScript", runJavaScript(t, module, path, false)},
				{"native", execute(t, "", binary, "--manifest", path)},
			} {
				positionMutantCheck(t, side.name, side.run.output, want)
			}
		})
	}
}
