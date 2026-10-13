package buildcache

import (
	"encoding/binary"
	"os"
	"strings"
	"testing"
)

// ldCacheBytes is a loader's cache in glibc's format, each entry flags, soname and path.
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
		if entry[0] != "" {
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
		got, ok := ldCacheValue(ldCacheBytes(entries))
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
		if _, ok := ldCacheValue(content); ok {
			t.Errorf("%s read as a cache", name)
		}
	}
	if content, err := os.ReadFile(ldCache); err == nil && strings.HasPrefix(string(content), "glibc-ld.so.cache1.1") {
		if _, ok := ldCacheValue(content); !ok {
			t.Fatal("this machine's own loader cache didn't read")
		}
	}
}
