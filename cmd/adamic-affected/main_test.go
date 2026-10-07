package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeInput(t *testing.T, path, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		t.Fatal(err)
	}
}
func TestObservedRelativeInputMutant(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "internal", "probe")
	fixture := filepath.Join(root, "oracle", "fixture.a")
	writeInput(t, fixture, "before")
	value := closure{Observed: map[string]string{}}
	observeLine(root, cwd, `openat(AT_FDCWD<`+cwd+`>, "../../oracle/fixture.a", O_RDONLY|O_CLOEXEC) = 3<`+fixture+`>`, &value)
	if len(value.Uncertain) > 0 {
		t.Fatal(value.Uncertain)
	}
	before, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	writeInput(t, fixture, "after")
	if !changed(root, value.Observed) {
		t.Fatal("relative input was skipped")
	}
	// The deliberate mutant drops observed inputs. An independent observation
	// of the fixture's real bytes catches the stale decision.
	mutant := map[string]string{}
	if changed(root, mutant) {
		t.Fatal("mutant did not recreate the wrongly skipped package")
	}
	after, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) == string(after) {
		t.Fatal("proof failed to catch observed-input mutant")
	}
	t.Log("observed mutant wrongly skips; independent fixture observation differs")
}
func TestDirectoryListingMutant(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "oracle", "fixtures")
	writeInput(t, filepath.Join(directory, "first.a"), "one")
	value := closure{Observed: map[string]string{}}
	observeLine(root, root, `getdents64(3<`+directory+`>, [{d_name="first.a"}], 32768) = 32`, &value)
	before, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	writeInput(t, filepath.Join(directory, "new.a"), "two")
	if !changed(root, value.Observed) {
		t.Fatal("new fixture was skipped")
	}
	mutant := map[string]string{}
	if changed(root, mutant) {
		t.Fatal("directory mutant did not wrongly skip")
	}
	after, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) == len(after) {
		t.Fatal("independent listing proof failed to catch mutant")
	}
	t.Log("directory mutant wrongly skips; independent fixture enumeration differs")
}
func TestToolchainMutant(t *testing.T) {
	before := map[string]string{"node --version": "v24.19.0\n", "go version": "same", "clang --version": "same"}
	after := map[string]string{"node --version": "v24.mutant\n", "go version": "same", "clang --version": "same"}
	if reflect.DeepEqual(before, after) {
		t.Fatal("changed node identity does not force all packages")
	}
	delete(before, "node --version")
	delete(after, "node --version")
	if !reflect.DeepEqual(before, after) {
		t.Fatal("mutant did not hide toolchain change")
	}
	t.Log("toolchain mutant hides changed Node; full identity comparison catches it")
}
func TestMissingPathsModesAndSymlinks(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "optional")
	inputs := map[string]string{}
	if err := add(root, path, inputs); err != nil {
		t.Fatal(err)
	}
	writeInput(t, path, "same")
	if !changed(root, inputs) {
		t.Fatal("missing path creation skipped")
	}
	if err := add(root, path, inputs); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0755); err != nil {
		t.Fatal(err)
	}
	if !changed(root, inputs) {
		t.Fatal("mode change skipped")
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	inputs = map[string]string{}
	if err := add(root, link, inputs); err != nil {
		t.Fatal(err)
	}
	writeInput(t, path, "other")
	if !changed(root, inputs) {
		t.Fatal("symlink target bytes skipped")
	}
}
func TestTraceUncertainty(t *testing.T) {
	for _, line := range []string{`openat(AT_FDCWD, "relative", O_RDONLY) = 3`, `openat(AT_FDCWD</repo>, "partial"...`, `<... openat resumed> ) = 3`, `stat(0x1234, 0x5678) = -1 EFAULT`} {
		value := closure{Observed: map[string]string{}}
		observeLine("/repo", "/repo", line, &value)
		if len(value.Uncertain) == 0 {
			t.Fatalf("uncertain trace allowed skipping: %s", line)
		}
	}
}
func TestDirectoryMembershipIgnoresUnlistedBytes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "fixtures")
	file := filepath.Join(path, "same.a")
	writeInput(t, file, "before")
	inputs := map[string]string{}
	if err := add(root, path, inputs); err != nil {
		t.Fatal(err)
	}
	writeInput(t, file, "after")
	if changed(root, inputs) {
		t.Fatal("directory list included unread file contents")
	}
	writeInput(t, filepath.Join(path, "added.a"), "new")
	if !changed(root, inputs) {
		t.Fatal("directory list missed new name")
	}
}
func TestTreeClosure(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "testdata")
	writeInput(t, filepath.Join(path, "nested", "fixture.a"), "before")
	inputs := map[string]string{}
	if err := tree(root, path, inputs); err != nil {
		t.Fatal(err)
	}
	for name := range inputs {
		if strings.HasPrefix(name, root) {
			t.Fatal("repository paths must be relocatable")
		}
	}
	writeInput(t, filepath.Join(path, "nested", "fixture.a"), "after")
	if !changed(root, inputs) {
		t.Fatal("recursive testdata bytes skipped")
	}
}

func TestRealProcessInputMutants(t *testing.T) {
	// Not parallel: integration runs external tools on one box.
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("observer compiler is unavailable")
	}
	root := t.TempDir()
	packageDirectory := filepath.Join(root, "pkg")
	writeInput(t, filepath.Join(root, "go.mod"), "module affected-proof\n\ngo 1.27\n")
	fixture := filepath.Join(root, "oracle", "value.txt")
	listing := filepath.Join(root, "oracle", "fixtures")
	writeInput(t, fixture, "before")
	writeInput(t, filepath.Join(listing, "first.a"), "first")
	writeInput(t, filepath.Join(packageDirectory, "probe_test.go"), `package probe
import("os";"testing")
func TestRelative(t *testing.T){ b,e:=os.ReadFile("../oracle/value.txt");if e!=nil{t.Fatal(e)};if string(b)!="before"{t.Fatalf("relative fixture changed: %q",b)}}
func TestListing(t *testing.T){ entries,e:=os.ReadDir("../oracle/fixtures");if e!=nil{t.Fatal(e)};if len(entries)!=1{t.Fatalf("fixture enumeration changed: %d",len(entries))}}
`)
	binary := filepath.Join(t.TempDir(), "probe.test")
	if err := logged(root, binary+".build.log", "go", "test", "-c", "-o", binary, "./pkg"); err != nil {
		t.Fatal(err)
	}
	stem := filepath.Join(t.TempDir(), "baseline")
	value := closure{Observed: map[string]string{}, Events: stem + ".jsonl"}
	if err := tracedRun(root, packageInfo{ImportPath: "affected-proof/pkg", Dir: packageDirectory}, binary, stem, &value); err != nil {
		t.Fatal(err)
	}
	// Test the actual traced closure. Uncertainty would mask the deliberate
	// mutants, so this fixture is also required to be fully decoded.
	if len(value.Uncertain) != 0 {
		t.Fatalf("trace uncertainty masks mutants: %v", value.Uncertain)
	}
	if changed(root, value.Observed) {
		t.Fatal("baseline changed immediately")
	}
	writeInput(t, fixture, "after")
	if !changed(root, value.Observed) {
		t.Fatal("real process relative read was skipped")
	}
	if changed(root, map[string]string{}) {
		t.Fatal("observed mutant did not wrongly skip")
	}
	log := filepath.Join(t.TempDir(), "relative-proof.log")
	if err := logged(packageDirectory, log, binary, "-test.run=TestRelative", "-test.v"); err == nil {
		t.Fatal("full package proof did not catch observed mutant")
	}
	output, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), "relative fixture changed") {
		t.Fatal(string(output))
	}
	t.Log("real observed mutant wrongly skipped; uncached test process caught relative reads")
	writeInput(t, fixture, "before")
	withoutDirectories := map[string]string{}
	for path, hash := range value.Observed {
		absolute := path
		if !filepath.IsAbs(path) {
			absolute = filepath.Join(root, path)
		}
		info, err := os.Stat(absolute)
		if err == nil && !info.IsDir() {
			withoutDirectories[path] = hash
		}
	}
	writeInput(t, filepath.Join(listing, "new.a"), "new")
	if !changed(root, value.Observed) {
		t.Fatal("real directory listing addition was skipped")
	}
	if changed(root, withoutDirectories) {
		t.Fatal("directory mutant did not wrongly skip")
	}
	log = filepath.Join(t.TempDir(), "listing-proof.log")
	if err := logged(packageDirectory, log, binary, "-test.run=TestListing", "-test.v"); err == nil {
		t.Fatal("uncached package proof did not catch directory mutant")
	}
	output, err = os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), "fixture enumeration changed: 2") {
		t.Fatal(string(output))
	}
	t.Log("real directory mutant wrongly skipped; uncached fixture enumeration caught the new file")
}

func TestMissingProbeParentsAndDeviceDescriptors(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "inputs")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	value := closure{Observed: map[string]string{}}
	observeLine(root, root, `openat(AT_FDCWD<`+root+`>, "inputs/missing", O_RDONLY) = -1 ENOENT (No such file or directory)`, &value)
	if len(value.Uncertain) > 0 {
		t.Fatal(value.Uncertain)
	}
	if err := os.Chmod(parent, 0755); err != nil {
		t.Fatal(err)
	}
	if !changed(root, value.Observed) {
		t.Fatal("missing leaf hid changed parent permissions")
	}
	value = closure{Observed: map[string]string{}}
	observeLine(root, root, `openat(AT_FDCWD<`+root+`>, "/dev/null", O_RDWR) = 3</dev/null<char 1:3>>`, &value)
	if len(value.Uncertain) > 0 {
		t.Fatal("device write outside repository cannot be a repository write", value.Uncertain)
	}
}

func TestNodeObservationFailsClosed(t *testing.T) {
	root := t.TempDir()
	fixture := filepath.Join(root, "value.txt")
	writeInput(t, fixture, "before")
	observer, err := buildNotificationObserver(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	stem := filepath.Join(t.TempDir(), "node")
	if err := logged(root, stem+".output", observer, "-o", stem+".trace", "--", "node", "-e", "require('fs').readFileSync('value.txt')"); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(stem + ".trace")
	if err != nil {
		t.Fatal(err)
	}
	value := closure{Observed: map[string]string{}}
	for _, line := range strings.Split(strings.TrimSpace(string(contents)), "\n") {
		observeLine(root, root, line, &value)
	}
	if _, found := value.Observed["value.txt"]; !found {
		t.Fatal("Node child read was not observed")
	}
	// io_uring cannot yet be decoded. Its appearance must prevent skipping.
	if strings.Contains(string(contents), "uncertain multi-path syscall") && len(value.Uncertain) == 0 {
		t.Fatal("unsupported asynchronous input did not fail closed")
	}
}

func TestObservedPathUnionKeyMutant(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"left", "right"} {
		writeInput(t, filepath.Join(root, directory, "value.a"), "before")
	}
	value := closure{Observed: map[string]string{}}
	for _, directory := range []string{"left", "right", "left"} {
		path := filepath.Join(root, directory, "value.a")
		collectLine(root, root, `openat(AT_FDCWD<`+root+`>, "`+directory+`/value.a", O_RDONLY) = entry<`+path+`>`, &value)
	}
	if len(value.Uncertain) > 0 {
		t.Fatal(value.Uncertain)
	}
	mutant := map[string]string{}
	for key := range value.Observed {
		// Drop the directory part of the union key. Hashes are still computed
		// fresh, but they now describe the wrong filesystem inputs.
		if err := add(root, filepath.Join(root, filepath.Base(key)), mutant); err != nil {
			t.Fatal(err)
		}
		if err := add(root, filepath.Join(root, key), value.Observed); err != nil {
			t.Fatal(err)
		}
	}
	writeInput(t, filepath.Join(root, "right", "value.a"), "after")
	if !changed(root, value.Observed) {
		t.Fatal("full path union missed changed bytes")
	}
	if changed(root, mutant) {
		t.Fatal("path-key mutant did not wrongly skip")
	}
	bytes, err := os.ReadFile(filepath.Join(root, "right", "value.a"))
	if err != nil || string(bytes) == "before" {
		t.Fatal("independent file read failed to catch path-key mutant")
	}
	t.Log("path union mutant dropped directories; independent file bytes caught its wrongful skip")
}
