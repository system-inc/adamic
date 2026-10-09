package estree

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// Opt-in measurements run one unit per process so object-cache state is explicit.
func TestRecoveryNativeSplit_000(t *testing.T) { t.Parallel(); measureRecoveryNativeSplit(t, "") }
func TestRecoveryNativeSplit_001(t *testing.T) {
	t.Parallel()
	measureRecoveryNativeSplit(t, "first-accessibility")
}
func TestRecoveryNativeSplit_002(t *testing.T) {
	t.Parallel()
	measureRecoveryNativeSplit(t, "empty-type-list-range")
}
func TestRecoveryNativeSplit_003(t *testing.T) {
	t.Parallel()
	measureRecoveryNativeSplit(t, "module-await")
}

func measureRecoveryNativeSplit(t *testing.T, name string) {
	t.Helper()
	if os.Getenv("ADAMIC_RECOVERY_SPLIT_MEASURE") != "1" {
		t.Skip("opt-in native split measurement")
	}
	started := time.Now()
	setup := beginRecoverySetup(t)
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range recoveryMutations {
		if m.name == name {
			main = mutantPort(t, m.file, m.from, m.to)
		}
	}
	lowered := recoveryLowered(t, setup, main)
	code, err := os.ReadFile(filepath.Join(lowered, "port.c"))
	if err != nil {
		t.Fatal(err)
	}
	lowering := time.Since(started)
	options := native.Options{Sanitize: true, Split: true, Jobs: 4}
	inputs := buildcache.Inputs{
		Name:      "estree-sanitized-native-split",
		Files:     []string{"internal/native", "stage1/cohere/estree/recovery_native_split_measure_test.go"},
		Flags:     append(native.Flags(options), "split=true", "jobs=4", fmt.Sprintf("source-sha256=%x", sha256.Sum256(code))),
		Toolchain: []string{buildcache.Tool("clang", "--version"), buildcache.Tool("getconf", "GNU_LIBC_VERSION")},
	}
	compileStarted := time.Now()
	directory := buildcache.Product(t, inputs, func(directory string) error {
		return native.Build(string(code), filepath.Join(directory, "port"), options)
	})
	t.Logf("SPLIT name=%q lowering=%.6fs native=%.6fs preparation=%.6fs", name, lowering.Seconds(), time.Since(compileStarted).Seconds(), time.Since(started).Seconds())
	list := manifest(t, recoveredGrammar())
	oracle := recoveryAnswer(t, recoveryOracle(t, setup), list, "--manifest")
	got := execute(t, "", filepath.Join(directory, "port"), "--manifest", list)
	if failure := recoveryComparison(oracle, got, name != ""); failure != "" {
		t.Fatal(failure)
	}
	if diff := firstDifference(onNode(t, main, "--manifest", list), got); diff != "" {
		t.Fatal("split/source Node: " + diff)
	}
}
