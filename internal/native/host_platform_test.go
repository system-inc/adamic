package native_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// process.platform's table (#t90fcv5, Platforms' ruling): the uname() sysnames whose lowercase is
// Node's own name map to it, and every other sysname has no answer, so the runtime stops, named,
// rather than pass a lowercased guess through (Node says android or cygwin where uname says Linux or
// CYGWIN_NT-10.0). The header compiles alone, so the table is tested here without a runtime.
// Mutants: a lowercasing fallback answers "cygwin_nt-10.0" and "netbsd" and goes red; a dropped row
// answers "(none)" for its sysname and goes red.
func TestHostPlatformNames(t *testing.T) {
	t.Parallel()
	header, err := filepath.Abs(filepath.Join("runtime", "host_platform.h"))
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	probe := filepath.Join(directory, "probe.c")
	source := "#include \"" + header + "\"\n#include <stdio.h>\nint main(int count, char **names) {\n\tfor (int at = 1; at < count; at++) {\n\t\tconst char *name = adamic_host_platform_name(names[at]);\n\t\tprintf(\"%s=%s\\n\", names[at], name == NULL ? \"(none)\" : name);\n\t}\n\treturn 0;\n}\n"
	if err := os.WriteFile(probe, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "probe")
	if output, err := exec.Command("clang", "-std=c11", "-Wall", "-Werror", "-o", binary, probe).CombinedOutput(); err != nil {
		t.Fatalf("clang: %v\n%s", err, output)
	}
	want := map[string]string{
		"Linux": "linux", "Darwin": "darwin", "FreeBSD": "freebsd", "OpenBSD": "openbsd", "SunOS": "sunos", "AIX": "aix",
		"CYGWIN_NT-10.0": "(none)", "Windows_NT": "(none)", "NetBSD": "(none)", "linux": "(none)", "": "(none)",
	}
	sysnames := []string{}
	for sysname := range want {
		sysnames = append(sysnames, sysname)
	}
	output, err := exec.Command(binary, sysnames...).Output()
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")
	if len(lines) != len(sysnames) {
		t.Fatalf("want %d answers, got %q", len(sysnames), output)
	}
	for index, line := range lines {
		if expected := sysnames[index] + "=" + want[sysnames[index]]; line != expected {
			t.Errorf("got %q, want %q", line, expected)
		}
	}
}
