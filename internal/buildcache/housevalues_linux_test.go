package buildcache

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
)

// houseNamespaceChild marks the test binary run again inside a user namespace, as a unit's test command is.
const houseNamespaceChild = "ADAMIC_TEST_HOUSE_NAMESPACE_CHILD"

// The house's own copy, root's in /var/lib/loom/house, is read the same from the host's namespace and from a unit's, which
// maps only the serving user (runner/modulecache_linux.go) so host root shows as the overflow uid (#4wh8ybq: every unit
// of house-values-7's run missed on it). Inside that namespace a copy of the test's own refuses, as does one in a
// directory it can write and one whose bytes aren't its name. Skipped on a box root hasn't given a copy. Mutants: the
// overflow uid never taken, the directory check dropped, the digest compare dropped.
// Not parallel: inside the namespace it sets the house's owners.
func TestTheHousesCopyIsReadInsideAUnitsNamespace(t *testing.T) {
	copies, _ := filepath.Glob("/var/lib/loom/house/*.json")
	if len(copies) == 0 {
		t.Skip("no house copy in /var/lib/loom/house on this box")
	}
	house := copies[0]
	if os.Getenv(houseNamespaceChild) == "" {
		if _, err := readHouseValues(house); err != nil {
			t.Fatalf("in the host's namespace, root's copy refused: %v", err)
		}
		command := exec.Command(os.Args[0], "-test.run=^TestTheHousesCopyIsReadInsideAUnitsNamespace$", "-test.v")
		command.Env = append(os.Environ(), houseNamespaceChild+"=1")
		command.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags:  syscall.CLONE_NEWUSER,
			UidMappings: []syscall.SysProcIDMap{{ContainerID: os.Getuid(), HostID: os.Getuid(), Size: 1}},
			GidMappings: []syscall.SysProcIDMap{{ContainerID: os.Getgid(), HostID: os.Getgid(), Size: 1}},
		}
		output, err := command.CombinedOutput()
		if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) {
			t.Skipf("this box refuses a user namespace: %v", err)
		}
		if err != nil || !strings.Contains(string(output), "--- PASS: TestTheHousesCopyIsReadInsideAUnitsNamespace") {
			t.Fatalf("inside a unit's namespace: %v\n%s", err, output)
		}
		return
	}
	owners, err := ownersHere()
	if err != nil || slices.Contains(owners, 0) || len(owners) != 1 {
		t.Fatalf("inside, root shows as %v, %v, never the overflow uid alone", owners, err)
	}
	if _, err := readHouseValues(house); err != nil {
		t.Fatalf("inside, root's copy refused: %v", err)
	}
	content, err := os.ReadFile(house)
	if err != nil {
		t.Fatal(err)
	}
	own := houseCopy(t, string(content), 0o444)
	if _, err := readHouseValues(own); err == nil || !strings.Contains(err.Error(), "not root") {
		t.Fatalf("inside, a copy of the test's own was read: %v", err)
	}
	originalOwners := houseOwners
	houseOwners = append([]uint32{0}, uint32(os.Getuid()))
	t.Cleanup(func() { houseOwners = originalOwners })
	os.Chmod(filepath.Dir(own), 0o755)
	if _, err := readHouseValues(own); err == nil || !strings.Contains(err.Error(), "can be written by this process") {
		t.Fatalf("inside, a copy in a writable directory was read: %v", err)
	}
	unwritable(t, filepath.Dir(own))
	forged := filepath.Join(t.TempDir(), filepath.Base(house))
	if err := os.WriteFile(forged, append(content, ' '), 0o444); err != nil {
		t.Fatal(err)
	}
	unwritable(t, filepath.Dir(forged))
	if _, err := readHouseValues(forged); err == nil || !strings.Contains(err.Error(), "not the sha256 its name carries") {
		t.Fatalf("inside, a copy whose bytes aren't its name was read: %v", err)
	}
}
