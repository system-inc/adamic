package assertions

import (
	"bytes"
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testCompileProfilesShards = 3

// ADAMIC_TEST_SHARD=i/n selects units locally; unset runs all. The gate selects
// shard-NNN directly. Compiler products are prepared once before the units until
// internal/buildcache is available; units verify every original backend contract.
func TestCompileProfiles(t *testing.T) {
	entry, err := filepath.Abs("profile.a")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	// Inputs: Name=lowered-profile; Files=profile.a and transitive imports,
	// Go compiler sources; Flags=default lowering; Toolchain=Go.
	lowerProduct := func(dir string) error {
		program, err := load.Load([]string{entry})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "profile.c"), []byte(native.C(lowered)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "emitted.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
	}
	start := time.Now()
	if err := lowerProduct(directory); err != nil {
		t.Fatal(err)
	}
	t.Logf("build lowered profile and emitted JS wall %.6fs", time.Since(start).Seconds())
	source, err := os.ReadFile(filepath.Join(directory, "profile.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, build := range []struct {
		name    string
		options native.Options
	}{
		{"native", native.Options{Sanitize: true}}, {"release", native.Options{}},
	} {
		// Inputs: Name=build.name; Files=profile.c and native runtime;
		// Flags=native.Flags(build.options); Toolchain=clang. Each product is shared.
		product := func(dir string) error {
			return native.Build(string(source), filepath.Join(dir, build.name), build.options)
		}
		start := time.Now()
		if err := product(directory); err != nil {
			t.Fatal(err)
		}
		t.Logf("build %s profile wall %.6fs", build.name, time.Since(start).Seconds())
	}
	products := []struct{ name, file string }{
		{"sanitized-native", "native"}, {"release-native", "release"}, {"emitted-javascript", "emitted.mjs"},
	}
	// Compare the actual unit inventory against the unsplit backend enumeration.
	expected := map[string]bool{"sanitized-native": true, "release-native": true, "emitted-javascript": true}
	seen := map[string]bool{}
	for _, product := range products {
		if seen[product.name] || !expected[product.name] {
			t.Fatalf("repeated or unexpected backend case %q", product.name)
		}
		seen[product.name] = true
	}
	if len(products) != testCompileProfilesShards || len(seen) != len(expected) {
		t.Fatalf("enumerated %d shards, declared %d", len(products), testCompileProfilesShards)
	}
	for id := range expected {
		if !seen[id] {
			t.Fatalf("missing backend case %q", id)
		}
	}
	t.Logf("union: one profile x %d backends = %d unique cases across %d shards", len(expected), len(seen), len(products))
	selected, count := -1, 1
	if value := os.Getenv("ADAMIC_TEST_SHARD"); value != "" {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
		selected, err = strconv.Atoi(parts[0])
		if err != nil {
			t.Fatal(err)
		}
		count, err = strconv.Atoi(parts[1])
		if err != nil || count < 1 || selected < 0 || selected >= count {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
	}
	for i, product := range products {
		if selected >= 0 && i%count != selected {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			defer func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %s exceeds 30s", elapsed)
				}
			}()
			t.Logf("profile.a: %s", product.name)
			if err := checkCompiledProfile(filepath.Join(directory, product.file), product.name); err != nil {
				t.Fatal(err)
			}
			// Passing the real release binary as the sanitized product plants wrong
			// compiler flags. Only shard-000 must catch the missing ASan instrumentation.
			t.Run("planted-wrong-sanitizer-flags", func(t *testing.T) {
				file := product.file
				if i == 0 {
					file = "release"
				}
				err := checkCompiledProfile(filepath.Join(directory, file), product.name)
				if (err != nil) != (i == 0) {
					t.Fatalf("wrong-flags mutant shard-%03d: %v", i, err)
				}
				if i == 0 && !strings.Contains(err.Error(), "ASan instrumentation missing") {
					t.Fatalf("wrong failure: %v", err)
				}
				t.Logf("wrong-flags mutant caught by shard-%03d=%v", i, err != nil)
			})
		})
	}
}

func checkCompiledProfile(path, backend string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || len(data) == 0 {
		return fmt.Errorf("%s: missing compiled artifact", backend)
	}
	if backend == "emitted-javascript" {
		output, err := exec.Command("node", "--check", path).CombinedOutput()
		if err != nil || len(output) != 0 {
			return fmt.Errorf("emitted JS syntax: %v: %s", err, output)
		}
		return nil
	}
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("%s: native artifact is not executable", backend)
	}
	sanitized := bytes.Contains(data, []byte("__asan_init"))
	if backend == "sanitized-native" && !sanitized {
		return fmt.Errorf("sanitized-native: ASan instrumentation missing")
	}
	if backend == "release-native" && sanitized {
		return fmt.Errorf("release-native: unexpected ASan instrumentation")
	}
	return nil
}
