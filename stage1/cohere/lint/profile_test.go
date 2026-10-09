package lint

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Opt-in artifacts outlive t.TempDir so callgrind and the timing runner use the same snapshot.
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
	buildProfile(t, directory)
}

func buildProfile(t *testing.T, directory string) {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(lowered)
	if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	if err := native.Build(source, filepath.Join(directory, "scanner"), native.Options{}); err != nil {
		t.Fatal(err)
	}
	if err := native.Build(source, filepath.Join(directory, "counted"), native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
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
	flags := append(native.Flags(native.Options{}), "-g", "-o", filepath.Join(directory, "profiled"))
	flags = append(flags, units...)
	flags = append(flags, "-lm")
	execute(t, "", "clang", flags...)
	t.Logf("release, counted and -O2 -g profiling builds saved in %s", directory)
}

// Exercise the shared profile graph without requiring the external compiler corpus.
func TestProfileCompilation_000(t *testing.T) {
	t.Parallel()
	defer compilationBudget(t)()
	compilationSelected(t)
	compilationCase(t)
	directory := t.TempDir()
	copyPort(t, directory, "", "")
	prepareRegistry(t, directory)
	products := compilationProducts(t)
	path := manifest(t, []string{ownedWitnesses(t, directory, "no-var")[0] + "\tno-var"})
	oracle := goOracle(t)
	want, countedWant := compilationOracleOutputs(t, oracle, path)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"emitted JavaScript", runJavaScript(t, filepath.Join(products, "lint.mjs"), path, false)},
		{"native", execute(t, "", filepath.Join(products, "scanner"), "--manifest", path)},
	} {
		if diff := difference(side.run.output, want); diff != "" {
			t.Fatalf("%s: %s", side.name, diff)
		}
	}
	got := execute(t, "", filepath.Join(products, "profiled"), "--manifest", path).output
	if os.Getenv("ADAMIC_PROFILE_COMPILATION_PLANT") == "1" {
		got = []byte("planted profile disagreement")
	}
	if diff := difference(got, want); diff != "" {
		t.Fatal(diff)
	}
	want = countedWant
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
	command := compilationCommand(t, filepath.Join(products, "counted"), "--manifest", path, "--count")
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
	if diff := difference(got, want); diff != "" {
		t.Fatal(diff)
	}
}

// Not parallel: upstream fixture capture uses process-wide environment state.
func TestProfileSnapshotsAgree(t *testing.T) {
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
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"emitted JavaScript", emittedNode(t, directory, path, false)},
	} {
		if diff := difference(side.run.output, want); diff == "" {
			t.Fatalf("fold mutant survived on %s", side.name)
		} else {
			t.Logf("fold mutant caught on %s: %s", side.name, diff)
		}
	}
}
func TestPositionIndexMutant(t *testing.T) {
	t.Parallel()
	path := manifest(t, generated(t))
	want := execute(t, "", goOracle(t), "--manifest", path).output
	directory := mutant(t, "this.anchors[0] = true;", "this.anchors[0] = false;", "rules/no-warning-comments/rule.ts")
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"emitted JavaScript", emittedNode(t, directory, path, false)},
	} {
		if diff := difference(side.run.output, want); diff == "" {
			t.Fatalf("position mutant survived on %s", side.name)
		} else {
			t.Logf("position mutant caught on %s: %s", side.name, diff)
		}
	}
}
