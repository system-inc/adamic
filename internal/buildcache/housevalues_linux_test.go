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

// The house's own copy, root's in /var/lib/loom/house, is read the same from the host's namespace and from a unit's,
// which maps only the serving user (runner/modulecache_linux.go) so host root shows as the overflow uid (#4wh8ybq:
// every unit of house-values-7's run missed on it). Inside that namespace a copy of the test's own refuses, as does one
// in a directory it can write and one whose bytes aren't its name. Skipped on a box root hasn't given a copy. Mutants:
// the overflow uid never taken, the directory check dropped, the digest compare dropped.
// Not parallel: inside the namespace it sets the house's owners.
func TestTheHousesCopyIsReadInsideAUnitsNamespace(test *testing.T) {
	copies, _ := filepath.Glob("/var/lib/loom/house/*.json")
	if len(copies) == 0 {
		test.Skip("no house copy in /var/lib/loom/house on this box")
	}
	house := copies[0]
	if os.Getenv(houseNamespaceChild) == "" {
		if _, readHouseValuesError := readHouseValues(house); readHouseValuesError != nil {
			test.Fatalf("in the host's namespace, root's copy refused: %v", readHouseValuesError)
		}
		command := exec.Command(os.Args[0], "-test.run=^TestTheHousesCopyIsReadInsideAUnitsNamespace$", "-test.v")
		command.Env = append(os.Environ(), houseNamespaceChild+"=1")
		command.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags:  syscall.CLONE_NEWUSER,
			UidMappings: []syscall.SysProcIDMap{{ContainerID: os.Getuid(), HostID: os.Getuid(), Size: 1}},
			GidMappings: []syscall.SysProcIDMap{{ContainerID: os.Getgid(), HostID: os.Getgid(), Size: 1}},
		}
		output, combinedOutputError := command.CombinedOutput()
		if errors.Is(combinedOutputError, syscall.EPERM) || errors.Is(combinedOutputError, syscall.EACCES) {
			test.Skipf("this box refuses a user namespace: %v", combinedOutputError)
		}
		if combinedOutputError != nil ||
			!strings.Contains(string(output), "--- PASS: TestTheHousesCopyIsReadInsideAUnitsNamespace") {
			test.Fatalf("inside a unit's namespace: %v\n%s", combinedOutputError, output)
		}
		return
	}
	owners, ownersHereError := ownersHere()
	if ownersHereError != nil || slices.Contains(owners, 0) || len(owners) != 1 {
		test.Fatalf("inside, root shows as %v, %v, never the overflow uid alone", owners, ownersHereError)
	}
	if _, readHouseValuesError := readHouseValues(house); readHouseValuesError != nil {
		test.Fatalf("inside, root's copy refused: %v", readHouseValuesError)
	}
	content, readError := os.ReadFile(house)
	if readError != nil {
		test.Fatal(readError)
	}
	own := houseCopy(test, string(content), 0o444)
	if _, readHouseValuesError := readHouseValues(
		own,
	); readHouseValuesError == nil || !strings.Contains(readHouseValuesError.Error(), "not root") {
		test.Fatalf("inside, a copy of the test's own was read: %v", readHouseValuesError)
	}
	originalOwners := houseOwners
	houseOwners = append([]uint32{0}, uint32(os.Getuid()))
	test.Cleanup(func() { houseOwners = originalOwners })
	os.Chmod(filepath.Dir(own), 0o755)
	if _, readHouseValuesError := readHouseValues(own); readHouseValuesError == nil ||
		!strings.Contains(readHouseValuesError.Error(), "can be written by this process") {
		test.Fatalf("inside, a copy in a writable directory was read: %v", readHouseValuesError)
	}
	unwritable(test, filepath.Dir(own))
	forged := filepath.Join(test.TempDir(), filepath.Base(house))
	if writeError := os.WriteFile(forged, append(content, ' '), 0o444); writeError != nil {
		test.Fatal(writeError)
	}
	unwritable(test, filepath.Dir(forged))
	if _, readHouseValuesError := readHouseValues(
		forged,
	); readHouseValuesError == nil ||
		!strings.Contains(readHouseValuesError.Error(), "not the sha256 its name carries") {
		test.Fatalf("inside, a copy whose bytes aren't its name was read: %v", readHouseValuesError)
	}
}
