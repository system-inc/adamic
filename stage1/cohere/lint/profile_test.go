package lint

import (
	"context"
	"github.com/system-inc/adamic/stage1/cohere/executable"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Opt-in artifacts outlive test.TempDir so callgrind and the timing runner use the same snapshot.
// Not parallel: artifacts are written to the fixed ADAMIC_LINT_PROFILE_DIR directory.
func TestProfileArtifacts(test *testing.T) {
	directory := os.Getenv("ADAMIC_LINT_PROFILE_DIR")
	if directory == "" {
		test.Skip("set ADAMIC_LINT_PROFILE_DIR to a scratch directory")
	}
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	sourceRoot := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if sourceRoot == "" {
		test.Fatal("missing pinned compiler checkout")
	}
	pin := execute(test, sourceRoot, "git", "rev-parse", "HEAD")
	if strings.TrimSpace(string(pin.output)) != compilerCommit {
		test.Fatal("compiler checkout has the wrong pin")
	}
	copyPort(test, directory, "", "")
	prepareRegistry(test, directory)
	var files []string
	walkError := filepath.WalkDir(filepath.Join(
		sourceRoot,
		"src/compiler",
	), func(path string, entry os.DirEntry, pathError error) error {
		if pathError == nil && !entry.IsDir() && strings.HasSuffix(path, ".ts") {
			files = append(files, path)
		}
		return pathError
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	var manifest strings.Builder
	for _, path := range files {
		manifest.WriteString(path + "\n")
	}
	if writeError := os.WriteFile(
		filepath.Join(directory, "compiler.txt"),
		[]byte(manifest.String()),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	oracle, readError := os.ReadFile(lintGoOracleProduct(test))
	if readError != nil {
		test.Fatal(readError)
	}
	if writeError := executable.Write(filepath.Join(directory, "oracle"), oracle); writeError != nil {
		test.Fatal(writeError)
	}
	buildProfile(test, directory)
}

func buildProfile(test *testing.T, directory string) {
	test.Helper()
	program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := native.C(lowered)
	if writeError := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if buildError := native.Build(source, filepath.Join(directory, "scanner"), native.Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	if buildError := native.Build(
		source,
		filepath.Join(directory, "counted"),
		native.Options{Count: true},
	); buildError != nil {
		test.Fatal(buildError)
	}
	// Reads: internal/native/runtime/*.[ch]
	runtime := filepath.Join(repository, "internal/native/runtime")
	entries, readDirectoryError := os.ReadDir(runtime)
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	units := []string{filepath.Join(directory, "main.c")}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h") {
			continue
		}
		data, readError := os.ReadFile(filepath.Join(runtime, entry.Name()))
		if readError != nil {
			test.Fatal(readError)
		}
		path := filepath.Join(directory, entry.Name())
		if writeError := os.WriteFile(path, data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
		if strings.HasSuffix(entry.Name(), ".c") {
			units = append(units, path)
		}
	}
	// The runtime's units are compiled with the program's features as flags, as main.c's #defines say.
	flags := append(native.SourceFlags(source, native.Options{}), "-g", "-o", filepath.Join(directory, "profiled"))
	flags = append(flags, units...)
	flags = append(flags, "-lm")
	execute(test, "", "clang", flags...)
	test.Logf("release, counted and -O2 -g profiling builds saved in %s", directory)
}

// Exercise the shared profile graph without requiring the external compiler corpus.
func TestProfileCompilation_000(test *testing.T) {
	test.Parallel()
	compilationSelected(test)
	directory, products, path, want, countedWant := compilationPrepare(test)
	// Shared product fetches and the overlay oracle finish before the case budget.
	defer compilationBudget(test)()

	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(test, directory, path, false)},
		{"emitted JavaScript", runJavaScript(test, filepath.Join(products, "lint.mjs"), path, false)},
		{"native", execute(test, "", filepath.Join(products, "scanner"), "--manifest", path)},
	} {
		if diff := difference(side.run.output, want); diff != "" {
			test.Fatalf("%s: %s", side.name, diff)
		}
	}
	got := execute(test, "", filepath.Join(products, "profiled"), "--manifest", path).output
	if os.Getenv("ADAMIC_PROFILE_COMPILATION_PLANT") == "1" {
		got = []byte("planted profile disagreement")
	}
	if diff := difference(got, want); diff != "" {
		test.Fatal(diff)
	}
	want = countedWant
	output, temporaryFileError := os.CreateTemp(test.TempDir(), "counted-output-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer output.Close()
	stats, temporaryFileError := os.CreateTemp(test.TempDir(), "counted-stats-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer stats.Close()
	command := compilationCommand(test, filepath.Join(products, "counted"), "--manifest", path, "--count")
	command.Stdout, command.Stderr = output, stats
	if runError := command.Run(); runError != nil {
		test.Fatal(runError)
	}
	got, readError := os.ReadFile(output.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	counters, readError := os.ReadFile(stats.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	if !strings.HasPrefix(string(counters), "adamic: counts: allocations ") {
		test.Fatalf("unexpected counted stderr: %s", counters)
	}
	test.Logf("counted instrumentation: %s", counters)
	if diff := difference(got, want); diff != "" {
		test.Fatal(diff)
	}
}

// Not parallel: upstream fixture capture uses process-wide environment state.
func TestProfileSnapshotsAgree(test *testing.T) {
	asked := os.Getenv("ADAMIC_LINT_PROFILE_SNAPSHOTS")
	if asked == "" {
		test.Skip("set ADAMIC_LINT_PROFILE_SNAPSHOTS")
	}
	rows := generated(test)
	var recovery []string
	for _, row := range upstream(test) {
		if strings.HasSuffix(row, "\tunsupported-recovery") {
			recovery = append(recovery, row)
		} else {
			rows = append(rows, row)
		}
	}
	for _, root := range []string{
		filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler"),
		filepath.Join(repository, "stage1"),
	} {
		walkError := filepath.WalkDir(root, func(path string, entry os.DirEntry, pathError error) error {
			if pathError != nil {
				return pathError
			}
			if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
				absolute, absolutePathError := filepath.Abs(path)
				if absolutePathError != nil {
					return absolutePathError
				}
				rows = append(rows, absolute)
			}
			return nil
		})
		if walkError != nil {
			test.Fatal(walkError)
		}
	}
	oracle := lintGoOracleProduct(test)
	// Classify recovery rows as TestRulesAgree does, so a captured case Go parses with a diagnostic (an
	// octal escape in strict mode, say) is compared in recovery mode rather than panicking the oracle.
	path := manifest(test, recoveryRows(test, oracle, rows))
	want := execute(test, "", oracle, "--manifest", path).output
	for _, directory := range filepath.SplitList(asked) {
		// Match the ordinary suite's explicit recovery boundary in both snapshots.
		for _, row := range recovery {
			checkRecoveryRefusal(test, oracle, filepath.Join(directory, "scanner"), directory, row)
			checkRecoveryRefusal(test, oracle, filepath.Join(directory, "profiled"), directory, row)
		}
		for _, side := range []struct {
			name string
			run  execution
		}{
			{"release", execute(test, "", filepath.Join(directory, "scanner"), "--manifest", path)},
			{"profiled", execute(test, "", filepath.Join(directory, "profiled"), "--manifest", path)},
			{"Node", node(test, directory, path, false)},
			{"emitted JavaScript", emittedNode(test, directory, path, false)},
		} {
			if diff := difference(side.run.output, want); diff != "" {
				test.Fatalf("%s %s: %s", directory, side.name, diff)
			}
		}
		test.Logf("%s: release, profiled and Node identical to Go: %d bytes", directory, len(want))
	}
}
func TestCommentFoldMutant(test *testing.T) {
	test.Parallel()
	path := manifest(test, generated(test))
	want := execute(test, "", lintGoOracleProduct(test), "--manifest", path).output
	directory := mutant(test, "point - 32 : point", "point - 31 : point", "unicode.ts")
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(test, directory, path, false)},
		{"emitted JavaScript", emittedNode(test, directory, path, false)},
	} {
		if diff := difference(side.run.output, want); diff == "" {
			test.Fatalf("fold mutant survived on %s", side.name)
		} else {
			test.Logf("fold mutant caught on %s: %s", side.name, diff)
		}
	}
}
func TestPositionIndexMutant(test *testing.T) {
	test.Parallel()
	path := manifest(test, generated(test))
	want := execute(test, "", lintGoOracleProduct(test), "--manifest", path).output
	directory := mutant(
		test,
		"this.anchors[0] = true;",
		"this.anchors[0] = false;",
		"rules/no-warning-comments/rule.ts",
	)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(test, directory, path, false)},
		{"emitted JavaScript", emittedNode(test, directory, path, false)},
	} {
		if diff := difference(side.run.output, want); diff == "" {
			test.Fatalf("position mutant survived on %s", side.name)
		} else {
			test.Logf("position mutant caught on %s: %s", side.name, diff)
		}
	}
}
