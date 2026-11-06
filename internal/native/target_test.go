package native

import (
	"os"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Not parallel: t.Setenv changes the process environment variable WASI_SYSROOT.
func TestWASITargetFlags(t *testing.T) {
	t.Setenv("WASI_SYSROOT", t.TempDir())
	flags := strings.Join(Flags(Options{Target: "wasm32-wasi"}), " ")
	for _, flag := range []string{"--target=wasm32-wasi", "--sysroot=" + os.Getenv("WASI_SYSROOT"), "-DADAMIC_TARGET_WASI=1", "-Oz", "-mno-atomics", "-ffp-contract=off"} {
		if !strings.Contains(flags, flag) {
			t.Errorf("missing %s in %s", flag, flags)
		}
	}
	if strings.Contains(strings.Join(Flags(Options{}), " "), "WASI") {
		t.Fatal("native flags select WASI")
	}
}

// Not parallel: t.Setenv changes the process environment variable WASI_SYSROOT.
func TestWASIRefusesUnsupportedOptions(t *testing.T) {
	t.Setenv("WASI_SYSROOT", t.TempDir())
	// Each refusal by its own words: past them a WASI build fails anyway on the empty sysroot's missing SDK clang, so
	// an error alone can't tell a dropped refusal from a kept one.
	for _, refusal := range []struct {
		options Options
		want    string
	}{
		{Options{Request: true}, "request exports require wasm32-wasi"},
		{Options{Target: "unknown"}, `unsupported target "unknown"`},
		{Options{Target: "wasm32-wasi", Sanitize: true}, "sanitizers are not supported for wasm32-wasi"},
		{Options{Target: "wasm32-wasi", Debug: true}, "debug builds are not supported for wasm32-wasi"},
		{Options{Target: "wasm32-wasi", cpu: "native"}, "native CPU selection is not supported for wasm32-wasi"},
	} {
		if err := ValidateOptions(refusal.options); err == nil || !strings.Contains(err.Error(), refusal.want) {
			t.Fatalf("%+v: got %v, want %q", refusal.options, err, refusal.want)
		}
	}
	t.Setenv("WASI_SYSROOT", "")
	if err := ValidateOptions(Options{Target: "wasm32-wasi"}); err == nil || !strings.Contains(err.Error(), "requires WASI_SYSROOT") {
		t.Fatalf("missing sysroot: %v", err)
	}
	t.Setenv("WASI_SYSROOT", filepath.Join(t.TempDir(), "absent"))
	if err := ValidateOptions(Options{Target: "wasm32-wasi"}); err == nil {
		t.Fatal("accepted nonexistent sysroot")
	}
}

// The triple table holds four rows and nothing else (#5xsd2ha): each platform's triple, Debian's spelling on linux and
// Apple's on macOS; any other pair refused by its name; and ValidateOptions taking each linux triple, the host's own
// darwin one alone, and refusing a darwin triple for another platform as a cross build no SDK is pinned for. Mutants:
// a row's triple misspelled; a fifth row; the darwin refusal dropped.
func TestTheTripleTable(t *testing.T) {
	t.Parallel()
	want := map[[2]string]string{
		{"linux", "amd64"}:  "x86_64-linux-gnu",
		{"linux", "arm64"}:  "aarch64-linux-gnu",
		{"darwin", "arm64"}: "arm64-apple-macosx",
		{"darwin", "amd64"}: "x86_64-apple-macosx",
	}
	for platform, triple := range want {
		if got, err := Triple(platform[0], platform[1]); err != nil || got != triple {
			t.Errorf("Triple(%s, %s) is %q, %v; want %s", platform[0], platform[1], got, err, triple)
		}
	}
	if len(triples) != len(want) {
		t.Errorf("the table holds %d rows, want the %d above", len(triples), len(want))
	}
	for _, platform := range [][2]string{{"windows", "amd64"}, {"linux", "riscv64"}, {"darwin", "386"}, {"", ""}} {
		_, err := Triple(platform[0], platform[1])
		if err == nil || !strings.Contains(err.Error(), "no native target for "+platform[0]+"/"+platform[1]+":") {
			t.Errorf("Triple(%q, %q): %v, want a refusal naming the pair", platform[0], platform[1], err)
		}
	}
	for platform, triple := range want {
		err := ValidateOptions(Options{Target: triple})
		host := platform[0] == goruntime.GOOS && platform[1] == goruntime.GOARCH
		switch {
		case platform[0] == "linux" || host:
			if err != nil {
				t.Errorf("ValidateOptions refuses %s on %s/%s: %v", triple, goruntime.GOOS, goruntime.GOARCH, err)
			}
		case err == nil || !strings.Contains(err.Error(), triple+": a darwin cross build has no pinned SDK"):
			t.Errorf("ValidateOptions on %s/%s took %s: %v, want it refused as a darwin cross build", goruntime.GOOS, goruntime.GOARCH, triple, err)
		}
	}
	if err := ValidateOptions(Options{Target: "riscv64-linux-gnu"}); err == nil || !strings.Contains(err.Error(), `unsupported target "riscv64-linux-gnu"`) {
		t.Errorf("ValidateOptions took a triple the table doesn't hold: %v", err)
	}
}

// An empty target is the build's own triple, read from `GOOS` and `GOARCH`, and this process's platform only where
// they're unset; a build for a darwin platform not the host's is refused by name, and one for a platform the table has
// no row for names it. Mutants: buildPlatform reading runtime.GOOS and runtime.GOARCH alone; the darwin refusal
// dropped.
// Not parallel: t.Setenv changes the process environment variables GOOS and GOARCH.
func TestAnEmptyTargetIsTheBuildsTriple(t *testing.T) {
	t.Setenv("GOOS", "")
	t.Setenv("GOARCH", "")
	host, err := Triple(goruntime.GOOS, goruntime.GOARCH)
	if err != nil {
		t.Skipf("this host has no native target: %v", err)
	}
	if resolved, err := ResolveTarget(Options{}); err != nil || resolved.Target != host {
		t.Errorf("unset, the target resolves to %q, %v; want the host's own %s", resolved.Target, err, host)
	}
	for _, arch := range []string{"amd64", "arm64"} {
		t.Setenv("GOOS", "linux")
		t.Setenv("GOARCH", arch)
		triple, _ := Triple("linux", arch)
		if resolved, err := ResolveTarget(Options{}); err != nil || resolved.Target != triple {
			t.Errorf("for linux/%s the target resolves to %q, %v; want %s", arch, resolved.Target, err, triple)
		}
		if resolved, err := ResolveTarget(Options{Target: "wasm32-wasi"}); err != nil || resolved.Target != "wasm32-wasi" {
			t.Errorf("for linux/%s wasm32-wasi resolves to %q, %v; want it kept", arch, resolved.Target, err)
		}
		t.Setenv("GOOS", "darwin")
		if arch == goruntime.GOARCH && goruntime.GOOS == "darwin" {
			continue
		}
		if err := ValidateOptions(Options{}); err == nil || !strings.Contains(err.Error(), "a darwin cross build has no pinned SDK") {
			t.Errorf("a darwin/%s build on %s/%s: %v, want it refused as a darwin cross build", arch, goruntime.GOOS, goruntime.GOARCH, err)
		}
	}
	t.Setenv("GOOS", "windows")
	t.Setenv("GOARCH", "amd64")
	if err := ValidateOptions(Options{}); err == nil || !strings.Contains(err.Error(), "no native target for windows/amd64") {
		t.Errorf("a windows/amd64 build: %v, want it refused by its platform's name", err)
	}
}

// One C product keyed for linux/amd64 and for linux/arm64 in one process keys apart: keyed as the products that build a
// native program key one, by native.Flags and the compiler's report, with no platform part of their own, so the triple
// Flags passes as --target is the only thing that tells the two arches apart. Mutant A: Flags without the --target, and
// the two keys collide here.
// Not parallel: t.Setenv changes the process environment variables GOOS and GOARCH.
func TestAProductKeysEachArchApart(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]string{}
	for _, arch := range []string{"amd64", "arm64"} {
		t.Setenv("GOOS", "linux")
		t.Setenv("GOARCH", arch)
		inputs := buildcache.Inputs{Name: "native triple probe", Flags: Flags(Options{Sanitize: true}), Toolchain: []string{"clang version 20.1.8"}}
		keys[arch] = buildcache.NameKey(repository, inputs)
	}
	if keys["amd64"] == keys["arm64"] {
		t.Fatalf("one product keyed for linux/amd64 and for linux/arm64 holds one key, %s: its flags name no target", keys["amd64"])
	}
}

// A native build cross-compiled for the arch this host isn't, `GOARCH` set to it, names that arch's triple in its flags
// and keys it in the runtime archive's key, as this machine and as every machine spell it, with that arch as the key's
// platform: an arm64 build on an amd64 box holds aarch64-linux-gnu and arm64, never x86_64-linux-gnu or amd64. The
// archive's key is read from its parts, with no clang run, since this host has no headers for the other arch. Mutant B:
// buildPlatform returning runtime.GOARCH always, so the cross build keys as this host's arch, and this goes red.
// Not parallel: t.Setenv changes the process environment variables GOOS and GOARCH.
func TestACrossBuildKeysTheBuildsArch(t *testing.T) {
	other := map[string]string{"amd64": "arm64", "arm64": "amd64"}[goruntime.GOARCH]
	if other == "" {
		t.Skipf("no other arch is a native target beside %s", goruntime.GOARCH)
	}
	t.Setenv("GOOS", "linux")
	t.Setenv("GOARCH", other)
	triple, err := Triple("linux", other)
	if err != nil {
		t.Fatal(err)
	}
	hostTriple, err := Triple("linux", goruntime.GOARCH)
	if err != nil {
		t.Fatal(err)
	}
	flags := Flags(Options{})
	if !slices.Contains(flags, "--target="+triple) || slices.Contains(flags, "--target="+hostTriple) {
		t.Errorf("a linux/%s build on a %s host compiles with %q, want --target=%s alone", other, goruntime.GOARCH, flags, triple)
	}
	files := []runtimeFile{{"probe.c", []byte("int probe;\n")}}
	for name, parts := range map[string][]string{
		"this machine's":  runtimeKeyParts(files, flags, "clang", "clang version 20.1.8", ""),
		"every machine's": portableKeyParts(files, flags, "clang", ""),
	} {
		for _, part := range []string{"arch " + other, "flag --target=" + triple} {
			if !slices.Contains(parts, part) {
				t.Errorf("the archive's key as %s spelling holds no %q: %q", name, part, parts)
			}
		}
		for _, part := range []string{"arch " + goruntime.GOARCH, "flag --target=" + hostTriple} {
			if slices.Contains(parts, part) {
				t.Errorf("the archive's key as %s spelling holds the host's %q: a linux/%s build keyed as %s", name, part, other, goruntime.GOARCH)
			}
		}
	}
}
