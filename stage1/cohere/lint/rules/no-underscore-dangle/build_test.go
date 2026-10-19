package assertions

import (
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"testing"
)

// The standalone owned profile's sanitized and release builds are keyed products, built ahead by these two tests and
// by nothing else, so no unit compiles C (#tze35sy, Loom's build law). Built inside a test, the sanitized build reached
// 73.1 s and the release build 61.5 s cold on a loaded tree (chg_s6f10de9, Oct 18), past the 60-second law. Timed on a
// drained Pupa, each phase alone: load and lower about 1 s, the sanitized build 18 to 19 s, the release build 8 to 9 s,
// the JavaScript 0.02 s.

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestProduct_NoUnderscoreDangleProfileSanitized(t *testing.T) {
	buildcache.RequireArtifacts(t, profileNative(t, native.Options{Sanitize: true}), buildcache.Artifact{Name: "native"})
}

// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestProduct_NoUnderscoreDangleProfileRelease(t *testing.T) {
	buildcache.RequireArtifacts(t, profileNative(t, native.Options{}), buildcache.Artifact{Name: "native"})
}

func TestNoUnderscoreDangleProfileEmitsJavaScript(t *testing.T) {
	t.Parallel()
	lowered, err := profileLowered()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(t.TempDir(), "emitted.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	t.Log("standalone profile: emitted JavaScript compiled")
}

// profileRule is this rule's directory under stage1/cohere/lint/rules, which names its products.
const profileRule = "no-underscore-dangle"

// profileNative is the profile built natively with options: lowered in this process, then compiled by clang. The
// compiler lowering it is linked into this test binary, where no trace sees it read, so its key names the compiler's
// paths (compilerinputs.Lowering) beside the profile's own sources.
func profileNative(t *testing.T, options native.Options) string {
	t.Helper()
	build := "release"
	if options.Sanitize {
		build = "sanitized"
	}
	flags := append([]string{"load.Load(profile.a)", "lower.Lower", "native.C", fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	// native.Build splits by ADAMIC_NATIVE_SPLIT when its options don't, so the switch is an input.
	flags = append(flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	inputs := buildcache.Inputs{
		Name:      "lint-rule-profile-" + profileRule + "-" + build,
		Files:     append(compilerinputs.Lowering(), profileSources(t)...),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")},
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		lowered, err := profileLowered()
		if err != nil {
			return err
		}
		return native.Build(native.C(lowered), filepath.Join(directory, "native"), options)
	})
}

// profileImport is a relative module specifier a source imports or re-exports from. A bare one ('adamic') is the
// standard library, which is the compiler's own (internal/).
var profileImport = regexp.MustCompile(`(?:from|import)\s+['"](\.[^'"]*)['"]`)

// profileSources is every file the profile's lowering reads beyond the compiler, relative to the repository root: this
// file, profile.a and everything it imports, followed import by import.
func profileSources(t *testing.T) []string {
	t.Helper()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs("profile.a")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{filepath.Join(root, "stage1/cohere/lint/rules", profileRule, "build_test.go"): true}
	pending := []string{entry}
	for len(pending) > 0 {
		path := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if seen[path] {
			continue
		}
		seen[path] = true
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range profileImport.FindAllStringSubmatch(string(source), -1) {
			pending = append(pending, filepath.Join(filepath.Dir(path), match[1]))
		}
	}
	var sources []string
	for path := range seen {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		sources = append(sources, filepath.ToSlash(relative))
	}
	sort.Strings(sources)
	return sources
}

// profileLowered loads and lowers profile.a, about a second.
func profileLowered() (*ir.Program, error) {
	entry, err := filepath.Abs("profile.a")
	if err != nil {
		return nil, err
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		return nil, err
	}
	return lower.Lower(context.Background(), program)
}
