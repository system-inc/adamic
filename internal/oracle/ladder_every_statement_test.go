package oracle

import (
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// A statement writes its throw checks as one ladder only where that's shorter than the checks in
// place (internal/native/exceptions.go), so most fixtures' checks are written in place and their
// own runs never take the ladder's releases. Under native.Options.LadderEveryStatement every
// statement with checks writes the ladder, and these tests run the exception and environment
// fixtures that way: each built sanitized, held to Node, and leak-checked where Node exits 0.
//
// The fixtures are every registered fixture whose name says throw, try, catch, finally or
// exception (fixtures, those that lower and aren't checked), and every environment_*.a in testdata,
// the count fixtures the Environment tests read. They're split in three so each test keeps within
// its budget: the exception fixtures, the environment_coverage_ ones, and the other environment
// ones. A group that writes no ladder at all fails, since then the knob did nothing.
func TestLadderEveryStatementExceptionFixturesAgreeWithNode(test *testing.T) {
	test.Parallel()
	laddersEverywhereAgree(test, "exceptions")
}

func TestLadderEveryStatementEnvironmentFixturesAgreeWithNode(test *testing.T) {
	test.Parallel()
	laddersEverywhereAgree(test, "environment")
}

func TestLadderEveryStatementEnvironmentCoverageFixturesAgreeWithNode(test *testing.T) {
	test.Parallel()
	laddersEverywhereAgree(test, "environment_coverage")
}

// ladderEveryStatementFixtures is a group's fixtures, as paths from the repository's root.
func ladderEveryStatementFixtures(test *testing.T, group string) []string {
	test.Helper()
	var paths []string
	if group == "exceptions" {
		named := regexp.MustCompile(`(?i)throw|try|catch|finally|exception`)
		for _, fixture := range fixtures {
			if fixture.lowers && !fixture.checked && named.MatchString(filepath.Base(fixture.path)) {
				paths = append(paths, fixture.path)
			}
		}
	} else {
		matches, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/environment_*.a"))
		if err != nil {
			test.Fatal(err)
		}
		for _, match := range matches {
			coverage := strings.HasPrefix(filepath.Base(match), "environment_coverage_")
			if coverage == (group == "environment_coverage") {
				paths = append(paths, filepath.Join("internal/oracle/testdata", filepath.Base(match)))
			}
		}
	}
	if len(paths) < 10 {
		test.Fatalf("%d %s fixtures, want the group's whole set", len(paths), group)
	}
	return paths
}

// laddersEverywhereAgree runs a group's fixtures with a ladder on every statement that has checks.
func laddersEverywhereAgree(test *testing.T, group string) {
	test.Helper()
	var laddered atomic.Int64
	test.Cleanup(func() {
		if !test.Failed() && laddered.Load() == 0 {
			test.Errorf("no %s fixture's C has a ladder under LadderEveryStatement", group)
		}
	})
	for _, fixture := range ladderEveryStatementFixtures(test, group) {
		test.Run(fixture, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture))
			if err != nil {
				test.Fatal(err)
			}
			want := onNode(test, path)
			program, err := lowered(test, path)
			if err != nil {
				test.Fatalf("Lower: %v", err)
			}
			code := native.CWithOptions(program, native.Options{LadderEveryStatement: true})
			if strings.Contains(code, "_done:;") {
				laddered.Add(1)
			}
			binary := filepath.Join(test.TempDir(), "program")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				test.Fatal(err)
			}
			var environment []string
			if runtime.GOOS == "linux" {
				environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
			}
			got := executeWith(test, environment, binary)
			if difference := disagreement(want, got); difference != "" {
				// Whether the program agrees with its checks in place says whether the ladder is at fault.
				inPlace, _ := nativelyUncached(test, program)
				test.Fatalf(
					"%s with a ladder on every statement (in place: %q)\nnode:   exit %d, stdout %q, stderr %q\nnative: exit %d, stdout %q, stderr %q",
					difference,
					disagreement(want, inPlace),
					want.exitCode,
					want.stdout,
					want.stderr,
					got.exitCode,
					got.stdout,
					got.stderr,
				)
			}
			if want.exitCode == 0 {
				if report := leakChecked(test, code, binary); report != "" {
					test.Errorf("leaks with a ladder on every statement:\n%s", report)
				}
			}
		})
	}
}
