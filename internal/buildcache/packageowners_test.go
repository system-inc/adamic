package buildcache

import (
	"testing"
)

// Where both architectures' twins of one package list a path, the loader architecture's twin owns it, as the house's
// own box says (#qmjv4z4: Marin's guest held libc6-dev and linux-libc-dev for amd64 and arm64, and the arm64 list,
// sorting last, named 29 byte-identical headers apart from Adam's). A path only one list names keeps it, a path two
// different packages list keeps the last as before, and a guest asking as arm64 takes the arm64 twin. Mutant: the last
// sorted list wins; any other architecture's list giving way, of another package too.
func TestATwinListedPathIsOwnedByTheLoaderArchitecturesTwin(t *testing.T) {
	t.Parallel()
	dpkg := t.TempDir()
	write(t, dpkg, "info/libc6-dev:amd64.list", "/usr/include/stdio.h\n/usr/lib/x86_64-linux-gnu/libc.so\n")
	write(t, dpkg, "info/libc6-dev:arm64.list", "/usr/include/stdio.h\n/usr/lib/aarch64-linux-gnu/libc.so\n")
	write(t, dpkg, "info/diverted.list", "/usr/bin/shared\n")
	write(t, dpkg, "info/linux-libc-dev:arm64.list", "/usr/include/linux/types.h\n")
	write(t, dpkg, "info/zz-other.list", "/usr/bin/shared\n")
	write(t, dpkg, "info/aa-tool:amd64.list", "/usr/bin/tool\n")
	write(t, dpkg, "info/bb-tool:arm64.list", "/usr/bin/tool\n")
	for architecture, want := range map[string]map[string]string{
		"amd64": {
			"/usr/include/stdio.h":               "libc6-dev:amd64",
			"/usr/lib/x86_64-linux-gnu/libc.so":  "libc6-dev:amd64",
			"/usr/lib/aarch64-linux-gnu/libc.so": "libc6-dev:arm64",
			"/usr/include/linux/types.h":         "linux-libc-dev:arm64",
			"/usr/bin/shared":                    "zz-other",
			"/usr/bin/tool":                      "bb-tool:arm64",
		},
		"arm64": {
			"/usr/include/stdio.h":       "libc6-dev:arm64",
			"/usr/include/linux/types.h": "linux-libc-dev:arm64",
			"/usr/bin/shared":            "zz-other",
		},
	} {
		owners := packageOwners(dpkg, architecture)
		for path, owner := range want {
			if owners[path] != owner {
				t.Errorf("asked as %s, %s is owned by %q, want %q", architecture, path, owners[path], owner)
			}
		}
	}
}
