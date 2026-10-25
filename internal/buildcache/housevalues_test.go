package buildcache

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// machineFixture is a machine of the test's own: a file dpkg installed and nobody changed (owned), another no package
// owns (loose), and a dpkg database naming owned, with the machine's places and dpkg's database pointed at them.
func machineFixture(t *testing.T) (owned, loose, dpkg string) {
	t.Helper()
	machine, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	system := filepath.Join(machine, "usr")
	dpkg = filepath.Join(machine, "dpkg")
	owned, loose = filepath.Join(system, "include/stdio.h"), filepath.Join(system, "lib/gcc/crtbegin.o")
	write(t, filepath.Dir(owned), filepath.Base(owned), "int printf(const char *, ...);")
	write(t, filepath.Dir(loose), filepath.Base(loose), "an object no package owns")
	write(t, dpkg, "info/libc6-dev:amd64.list", owned+"\n")
	write(t, dpkg, "info/libc6-dev:amd64.md5sums", fmt.Sprintf("%x  %s\n", md5.Sum([]byte("int printf(const char *, ...);")), strings.TrimPrefix(owned, "/")))
	write(t, dpkg, "status", "Package: libc6-dev\nArchitecture: amd64\nVersion: 2.39-0ubuntu8\n")
	originalPlaces, originalDpkg := systemPlaces, dpkgDirectory
	systemPlaces, dpkgDirectory = []string{system}, dpkg
	packageIndexes.Delete(dpkg)
	t.Cleanup(func() { systemPlaces, dpkgDirectory = originalPlaces, originalDpkg; packageIndexes.Delete(dpkg) })
	return owned, loose, dpkg
}

// handHouse writes a stamp naming paths, at mode, hands it as ADAMIC_HOUSE_VALUES and returns its path. The test's own
// uid is an owner a handed file may have, as root is on a house box.
func handHouse(t *testing.T, stamp string, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "house-system.json")
	if err := os.WriteFile(path, []byte(stamp), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
	originalOwners := houseOwners
	houseOwners = append([]uint32{0}, uint32(os.Getuid()))
	t.Setenv(houseValuesVariable, path)
	t.Cleanup(func() { houseOwners = originalOwners; handedHouses.Delete(path) })
	return path
}

// countPackageOpens counts every open of dpkg's database from here on, and those outside keying.
func countPackageOpens(t *testing.T) (opens, unkeyed *[]string) {
	t.Helper()
	opens, unkeyed = &[]string{}, &[]string{}
	original := openPackageFile
	openPackageFile = func(name string) (*os.File, error) {
		*opens = append(*opens, name)
		if keyingNow.Load() == 0 {
			*unkeyed = append(*unkeyed, name)
		}
		return original(name)
	}
	t.Cleanup(func() { openPackageFile = original })
	return opens, unkeyed
}

// Handed the house's values (#shadow), a machine file the stamp names is valued as the stamp has it, one it doesn't name
// by its content, and dpkg's database is never opened, so no key reads it under a unit. Unhanded, a machine file is
// valued through dpkg exactly as before. Mutant: dpkg consulted while a file is handed.
// Not parallel: it points the machine's places and dpkg's database at the test, and sets the house's owners.
func TestAHandedHouseValuesMachineFilesWithoutDpkg(t *testing.T) {
	owned, loose, dpkg := machineFixture(t)
	if value, err := systemEntryValue(owned); err != nil || value != "package libc6-dev:amd64 version 2.39-0ubuntu8" || value != systemValue(owned) {
		t.Fatalf("unhanded, a file dpkg installed values %q, %v", value, err)
	}
	if value, err := systemEntryValue(loose); err != nil || value != systemValue(loose) {
		t.Fatalf("unhanded, a file no package owns values %q, %v, and systemValue %q", value, err, systemValue(loose))
	}
	packageIndexes.Delete(dpkg)
	opens, _ := countPackageOpens(t)
	handHouse(t, fmt.Sprintf(`{"box": "Adam", "at": "2026-10-25T07:21:52Z", "paths": {%q: "file d4b198c4 false"}}`, loose), 0o444)
	if value, err := systemEntryValue(loose); err != nil || value != "house file d4b198c4 false" {
		t.Fatalf("handed, a stamped file values %q, %v", value, err)
	}
	content := sha256.Sum256([]byte("int printf(const char *, ...);"))
	if value, err := systemEntryValue(owned); err != nil || value != fmt.Sprintf("file false %x", content) {
		t.Fatalf("handed, a file the stamp doesn't name values %q, %v, never by its package", value, err)
	}
	if len(*opens) != 0 {
		t.Fatalf("handed, dpkg's database was opened: %q", *opens)
	}
}

// A handed file the code under test could have written, or one that isn't the house's stamp, never values a key: each is
// refused, and a lookup through it misses saying why, never valuing through dpkg or by content instead. Mutants: the
// owner check dropped, the mode check dropped, the link refusal dropped.
// Not parallel: it sets the house's owners and ADAMIC_HOUSE_VALUES.
func TestAHandedHouseFileTheCodeUnderTestCouldWriteIsRefused(t *testing.T) {
	_, loose, _ := machineFixture(t)
	good := fmt.Sprintf(`{"paths": {%q: "file d4b198c4 false"}}`, loose)
	for _, test := range []struct {
		name, stamp, want string
		mode              os.FileMode
		hand              func(t *testing.T, path string) string
	}{
		{name: "group-writable", stamp: good, mode: 0o464, want: "writable by group or other"},
		{name: "other-writable", stamp: good, mode: 0o446, want: "writable by group or other"},
		{name: "not root's", stamp: good, mode: 0o444, want: "not root", hand: func(t *testing.T, path string) string {
			if os.Getuid() == 0 {
				t.Skip("run as root, the test's own file is root's")
			}
			houseOwners = []uint32{0}
			return path
		}},
		{name: "a link", stamp: good, mode: 0o444, want: "house values", hand: func(t *testing.T, path string) string {
			link := filepath.Join(t.TempDir(), "link.json")
			if err := os.Symlink(path, link); err != nil {
				t.Fatal(err)
			}
			t.Setenv(houseValuesVariable, link)
			t.Cleanup(func() { handedHouses.Delete(link) })
			return link
		}},
		{name: "a directory", stamp: good, mode: 0o444, want: "not a regular file", hand: func(t *testing.T, path string) string {
			directory := t.TempDir()
			os.Chmod(directory, 0o555)
			t.Setenv(houseValuesVariable, directory)
			t.Cleanup(func() { os.Chmod(directory, 0o755); handedHouses.Delete(directory) })
			return directory
		}},
		{name: "naming no path", stamp: `{"box": "Adam", "paths": {}}`, mode: 0o444, want: "names no path"},
		{name: "not json", stamp: `paths`, mode: 0o444, want: "house values"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := handHouse(t, test.stamp, test.mode)
			if test.hand != nil {
				path = test.hand(t, path)
			}
			value, err := systemEntryValue(loose)
			if err == nil || !strings.Contains(err.Error(), test.want) || !strings.Contains(err.Error(), path) {
				t.Fatalf("handed %s, valued %q, %v", test.name, value, err)
			}
			if _, err := treeEntryValue("", "", readEntry{Kind: "system", Path: loose}); err == nil {
				t.Fatalf("handed %s, a lookup valued its system entry", test.name)
			}
		})
	}
}

// The stamp's box and moment are never in a value: two boxes whose stamps differ only by them key alike, as the house's
// one check says they should. Mutant: the box or the moment let into the value.
// Not parallel: it sets the house's owners and ADAMIC_HOUSE_VALUES.
func TestTheHousesBoxAndMomentNeverValueAKey(t *testing.T) {
	_, loose, _ := machineFixture(t)
	var values []string
	for _, box := range []string{`"box": "Adam", "at": "2026-10-25T07:21:52Z"`, `"box": "Eve", "at": "2026-10-24T01:00:00Z"`} {
		handHouse(t, fmt.Sprintf(`{%s, "paths": {%q: "file d4b198c4 false"}}`, box, loose), 0o444)
		value, err := systemEntryValue(loose)
		if err != nil {
			t.Fatal(err)
		}
		values = append(values, value)
	}
	if values[0] != values[1] {
		t.Fatalf("two stamps differing by box and moment value %q", values)
	}
}

// A traced lookup values a product's machine files as keying, so the dpkg database it reads to value them is the key's
// own work, never the build's or the test's (#shadow, Builder's read: the window holds; a unit's trace can't see it).
// Mutant: productFor run outside keying in get.
// Not parallel: newRig, and it points the machine's places and dpkg's database at the test.
func TestATracedLookupValuesMachineFilesAsKeying(t *testing.T) {
	r := newRig(t)
	owned, _, dpkg := machineFixture(t)
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() {
		r.open(r.pid, "source/a.c")
		r.open(r.pid, owned)
	})
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	forgetHeldForLife()
	systemContents.Range(func(key, _ any) bool { systemContents.Delete(key); return true })
	packageIndexes.Delete(dpkg)
	opens, unkeyed := countPackageOpens(t)
	if _, err := Get(port, func(string) error { t.Fatal("the lookup built"); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(*opens) == 0 || len(*unkeyed) != 0 {
		t.Fatalf("the lookup opened dpkg's database at %q; outside keying at %q", *opens, *unkeyed)
	}
}
