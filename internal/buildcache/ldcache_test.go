package buildcache

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ldCacheBytes is a loader's cache in glibc's format, each entry flags, soname and path: flags are x86-64's 0x303 when
// the first field is empty, an arm64 guest's 0xa03 for "aarch64", and 0x3 (an ELF libc6 library of no 64-bit kind) otherwise.
func ldCacheBytes(entries [][3]string) []byte {
	const header, entrySize = 48, 24
	strings := []byte{}
	offsets := make([][2]uint32, len(entries))
	base := header + len(entries)*entrySize
	for index, entry := range entries {
		offsets[index][0] = uint32(base + len(strings))
		strings = append(append(strings, entry[1]...), 0)
		offsets[index][1] = uint32(base + len(strings))
		strings = append(append(strings, entry[2]...), 0)
	}
	content := make([]byte, header, base+len(strings))
	copy(content, "glibc-ld.so.cache1.1")
	binary.LittleEndian.PutUint32(content[20:], uint32(len(entries)))
	binary.LittleEndian.PutUint32(content[24:], uint32(len(strings)))
	for index, entry := range entries {
		record := make([]byte, entrySize)
		flags := uint32(0x303)
		switch entry[0] {
		case "":
		case "aarch64":
			flags = 0xa03
		default:
			flags = 0x3
		}
		binary.LittleEndian.PutUint32(record[0:], flags)
		binary.LittleEndian.PutUint32(record[4:], offsets[index][0])
		binary.LittleEndian.PutUint32(record[8:], offsets[index][1])
		content = append(content, record...)
	}
	return append(content, strings...)
}

// The loader's cache is valued by what only it finds: its entries outside the loader's own directories and the Windows
// GPU driver's (WSL's) libraries. So two boxes that installed different libraries in the loader's directories, or run
// different drivers, value it the same (#6eexfyv), while a library ld.so.conf.d puts elsewhere, added, moved or flagged
// otherwise, moves the value; a cache that isn't glibc's format, or doesn't read, is valued by its content. Mutants: the
// loader's directories kept (an install there moves the value); everything outside them left out too (an ld.so.conf.d
// library keeps it); WSL's libraries kept.
func TestTheLoadersCacheIsWhatOnlyItFinds(t *testing.T) {
	t.Parallel()
	machine := [][3]string{{"", "libz.so.1", "/lib/x86_64-linux-gnu/libz.so.1"}, {"", "libc.so.6", "/usr/lib/x86_64-linux-gnu/libc.so.6"},
		{"", "libm.so.6", "/usr/lib/x86_64-linux-gnu/glibc-hwcaps/x86-64-v3/libm.so.6"}, {"", "libfakeroot-0.so", "/usr/lib/x86_64-linux-gnu/libfakeroot/libfakeroot-0.so"}}
	value := func(entries [][3]string) string {
		t.Helper()
		got, ok := ldCacheValueFor(ldCacheBytes(entries), 0x303)
		if !ok {
			t.Fatalf("a cache in glibc's format didn't read: %v", entries)
		}
		return got
	}
	with := func(extra ...[3]string) [][3]string { return append(append([][3]string{}, machine...), extra...) }
	plain := value(machine)
	for name, same := range map[string][][3]string{
		"a library installed in a loader directory": with([3]string{"", "libburn.so.4", "/lib/x86_64-linux-gnu/libburn.so.4"}),
		"one in a loader directory's glibc-hwcaps":  with([3]string{"", "libq.so.1", "/usr/lib/x86_64-linux-gnu/glibc-hwcaps/x86-64-v4/libq.so.1"}),
		"the Windows GPU driver's":                  with([3]string{"", "libnvidia-ngx.so.1", "/usr/lib/wsl/lib/libnvidia-ngx.so.1"}),
	} {
		if value(same) != plain {
			t.Errorf("%s moved the cache's value", name)
		}
	}
	for name, changed := range map[string][][3]string{
		"a library ld.so.conf.d adds outside the loader's directories": with([3]string{"", "libextra.so.1", "/usr/local/lib/libextra.so.1"}),
		"one outside them moved":             {machine[0], machine[1], machine[2], {"", "libfakeroot-0.so", "/usr/lib/x86_64-linux-gnu/libfakeroot2/libfakeroot-0.so"}},
		"one outside them flagged otherwise": {machine[0], machine[1], machine[2], {"x", "libfakeroot-0.so", "/usr/lib/x86_64-linux-gnu/libfakeroot/libfakeroot-0.so"}},
	} {
		if value(changed) == plain {
			t.Errorf("%s kept the cache's value", name)
		}
	}
	for name, content := range map[string][]byte{"another format": []byte("ld.so-1.7.0\x00 an old cache"), "a short one": ldCacheBytes(machine)[:60]} {
		if _, ok := ldCacheValueFor(content, 0x303); ok {
			t.Errorf("%s read as a cache", name)
		}
	}
	if _, known := loaderFlags[runtime.GOARCH]; !known {
		return
	}
	if content, err := os.ReadFile("/etc/ld.so.cache"); err == nil && strings.HasPrefix(string(content), "glibc-ld.so.cache1.1") {
		if _, ok := ldCacheValue(content); !ok {
			t.Fatal("this machine's own loader cache didn't read")
		}
	}
}

// The cache is valued by the entries the build's own loader takes, flagged exactly as glibc's _dl_cache_check_flags
// wants (0x303 on x86_64), so an arm64 Mac guest's cache, every entry its own aarch64 one (0xa03), values as one with
// none of them (#q4z4cmj, Mentat's M2 on #bdxnd3h: 91% of a guest's units missed on it). An x86-64 entry changed still
// moves the value. Mutants: every entry kept whatever its flags; the check dropped for the arm64 flag only.
func TestTheLoadersCacheIsOnlyWhatItsOwnArchitectureTakes(t *testing.T) {
	t.Parallel()
	house := [][3]string{{"", "libc.so.6", "/usr/lib/x86_64-linux-gnu/libc.so.6"},
		{"", "libfakeroot-0.so", "/usr/lib/x86_64-linux-gnu/libfakeroot/libfakeroot-0.so"}}
	guest := append(append([][3]string{}, house...), [3]string{"aarch64", "libc.so.6", "/lib/aarch64-linux-gnu/libc.so.6"},
		[3]string{"aarch64", "libLLVM.so.18.1", "/lib/aarch64-linux-gnu/libLLVM.so.18.1"},
		[3]string{"aarch64", "libfakeroot-0.so", "/usr/lib/aarch64-linux-gnu/libfakeroot/libfakeroot-0.so"})
	value := func(entries [][3]string) string {
		t.Helper()
		got, ok := ldCacheValueFor(ldCacheBytes(entries), 0x303)
		if !ok {
			t.Fatalf("a cache in glibc's format didn't read: %v", entries)
		}
		return got
	}
	if value(guest) != value(house) {
		t.Error("an arm64 guest's own aarch64 entries moved the value an x86-64 loader reads")
	}
	moved := [][3]string{house[0], {"", "libfakeroot-0.so", "/usr/lib/x86_64-linux-gnu/libfakeroot2/libfakeroot-0.so"}}
	if value(moved) == value(house) {
		t.Error("an x86-64 entry the loader takes moved and the value didn't")
	}
	if value([][3]string{house[0]}) == value(house) {
		t.Error("an x86-64 entry the loader takes went away and the value didn't")
	}
	// The real guests' caches, measured Oct 17: every entry 0xa03. One is as good as none to an amd64 loader.
	if value([][3]string{{"aarch64", "libc.so.6", "/lib/aarch64-linux-gnu/libc.so.6"}}) != value(nil) {
		t.Error("a cache holding only another architecture's entries doesn't value as an empty one")
	}
}

// A loader cache that doesn't read as the build's loader reads it is unreadable, a miss, never valued by its bytes nor
// as an empty cache: either would let two boxes whose loaders find different libraries value alike. Mutants: the
// content's hash kept as the fallback; an unreadable cache valued as no entries.
func TestALoaderCacheThatDoesntReadIsUnreadable(t *testing.T) {
	if _, known := loaderFlags[runtime.GOARCH]; !known {
		t.Skipf("no loader rule for %s: the value is unreadable here by design", runtime.GOARCH)
	}
	directory := t.TempDir()
	original := ldCache
	t.Cleanup(func() { ldCache = original })
	empty, ok := ldCacheValueFor(ldCacheBytes(nil), 0x303)
	if !ok {
		t.Fatal("an empty cache in glibc's format didn't read")
	}
	for name, content := range map[string][]byte{
		"another format": []byte("ld.so-1.7.0\x00 an old cache"),
		"a cut one":      ldCacheBytes([][3]string{{"", "libextra.so.1", "/usr/local/lib/libextra.so.1"}})[:60],
	} {
		ldCache = filepath.Join(directory, strings.ReplaceAll(name, " ", "-"))
		if err := os.WriteFile(ldCache, content, 0o644); err != nil {
			t.Fatal(err)
		}
		got := systemValue(ldCache)
		if !strings.HasPrefix(got, "unreadable: ") || got == empty {
			t.Errorf("%s values as %q, want unreadable", name, got)
		}
	}
	ldCache = filepath.Join(directory, "good")
	if err := os.WriteFile(ldCache, ldCacheBytes([][3]string{{"aarch64", "libc.so.6", "/lib/aarch64-linux-gnu/libc.so.6"}}), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := systemValue(ldCache); got != empty {
		t.Errorf("a guest's cache values as %q, want %q, the empty cache's", got, empty)
	}
}
