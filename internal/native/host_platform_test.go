package native_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// process.platform's table (#t90fcv5, Platforms' ruling): the name is the target's, from its
// predefined macros in Node's order. Each case preprocesses host_platform.h alone with the build
// host's own macros undefined (-undef) and one target's defined: Android before Linux, since an
// Android target defines both; wasm32-wasi names nothing, so the runtime panics there; any target the
// table doesn't name is refused when the runtime is built. Mutants: the Android check after Linux's
// answers linux for Android and goes red; a target's row removed reaches the #error and goes red.
func TestHostPlatformNames(t *testing.T) {
	t.Parallel()
	header, err := filepath.Abs(filepath.Join("runtime", "host_platform.h"))
	if err != nil {
		t.Fatal(err)
	}
	probe := filepath.Join(t.TempDir(), "probe.c")
	if err := os.WriteFile(probe, []byte("#include \""+header+"\"\nname ADAMIC_HOST_PLATFORM\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, target := range []struct {
		name   string
		macros []string
		want   string
	}{
		{"android", []string{"__ANDROID__", "__linux__"}, `name "android"`},
		{"linux", []string{"__linux__"}, `name "linux"`},
		{"darwin", []string{"__APPLE__"}, `name "darwin"`},
		{"freebsd", []string{"__FreeBSD__"}, `name "freebsd"`},
		{"openbsd", []string{"__OpenBSD__"}, `name "openbsd"`},
		{"sunos", []string{"__sun"}, `name "sunos"`},
		{"aix", []string{"_AIX"}, `name "aix"`},
		{"wasi", []string{"__wasi__"}, "name ADAMIC_HOST_PLATFORM"},
		{"unnamed", nil, ""},
		{"netbsd", []string{"__NetBSD__"}, ""},
	} {
		t.Run(target.name, func(t *testing.T) {
			t.Parallel()
			arguments := []string{"-E", "-P", "-undef"}
			for _, macro := range target.macros {
				arguments = append(arguments, "-D"+macro)
			}
			output, err := exec.Command("clang", append(arguments, probe)...).CombinedOutput()
			if target.want == "" {
				if err == nil || !strings.Contains(string(output), "a target host_platform.h doesn't name") {
					t.Fatalf("want the #error, got %v: %s", err, output)
				}
				return
			}
			if err != nil {
				t.Fatalf("clang: %v\n%s", err, output)
			}
			if got := strings.TrimSpace(string(output)); got != target.want {
				t.Fatalf("got %q, want %q", got, target.want)
			}
		})
	}
}
