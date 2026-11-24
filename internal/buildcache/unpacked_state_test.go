package buildcache

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// The commits unpackedSourceForTest's manifest names for the tree's own repository and its submodule.
const unpackedTreeHead = "1111111111111111111111111111111111111111"
const unpackedModuleHead = "2222222222222222222222222222222222222222"

// unpackedSourceForTest is a source as Loom unpacks one: the tree's files and a submodule's, no .git, and a .tracked
// manifest naming each repository's commit.
func unpackedSourceForTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, "source/a.c", "int a;\n")
	write(t, root, "module/inner.c", "int inner;\n")
	write(t, root, ".tracked/HEAD", unpackedTreeHead+"\n")
	write(t, root, ".tracked/files", "")
	write(t, root, ".tracked/module/HEAD", unpackedModuleHead+"\n")
	write(t, root, ".tracked/module/files", "")
	return root
}

// closedReadsForTest are the reads of a process closed out of every path under a tree: another uid owns all of it,
// none of it is writable, the process holds no capability, and the tree is on a writable mount, so only the no-write
// proof can hold. A test changes one read to open the tree.
func closedReadsForTest() closureReads {
	return closureReads{
		readOnlyMount: func(string) bool { return false },
		mountInfo:     func() ([]byte, error) { return nil, errors.New("this test reads no mountinfo") },
		status:        func() ([]byte, error) { return []byte("Name:\ttraced\nCapEff:\t0000000000000000\n"), nil },
		effectiveUid:  1000,
		owner:         func(string) (int, error) { return 2000, nil },
		writable:      func(string) (bool, error) { return false, nil },
	}
}

// proofForTest is closedProofWith over reads, as unpackedTreeState takes a proof.
func proofForTest(reads closureReads) func(root string) (string, error) {
	return func(root string) (string, error) { return closedProofWith(root, reads) }
}

// An unpacked source closed to the process records the no-write proof for its effective uid and each repository's
// commit from the manifest, the tree's own first, then dpkg's line where there is one.
// Mutant: the submodule's HEAD left out of manifestHeads: red.
func TestAClosedUnpackedSourceRecordsTheManifestsHeads(t *testing.T) {
	t.Parallel()
	state, err := unpackedTreeState(unpackedSourceForTest(t), proofForTest(closedReadsForTest()))
	if err != nil {
		t.Fatal(err)
	}
	want := "unpacked closed no-write uid 1000\nhead . " + unpackedTreeHead + "\nhead module " + unpackedModuleHead + "\n"
	// dpkg's line is read, not compared: a test beside this one may point dpkgDirectory elsewhere while it runs.
	if rest, ok := strings.CutPrefix(state, want); !ok || (rest != "" && !strings.HasPrefix(rest, "dpkg status ")) {
		t.Fatalf("recorded\n%s\nwant\n%s", state, want)
	}
}

// The no-write proof holds only when nothing under the root is the process's to change: a file it owns (whose mode its
// owner can change back), a file it may write, or a capability past both refuses, naming what was found, and so do
// capabilities that can't be read.
// Mutant: the owner not checked: "an owned file" goes red.
// Mutant: regular files not asked whether they are writable, only directories: "a writable file" goes red.
// Mutant: the capabilities not checked: every CAP_ case and both unreadable cases go red.
// Mutant: only CAP_DAC_OVERRIDE and CAP_FOWNER checked: "CAP_CHOWN" and "CAP_SYS_ADMIN" go red.
func TestTheNoWriteProofRefusesWhatTheProcessCouldChange(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		open func(reads *closureReads, file string)
		want func(file string) string
	}{
		"an owned file": {
			open: func(reads *closureReads, file string) {
				reads.owner = func(path string) (int, error) {
					if path == file {
						return 1000, nil
					}
					return 2000, nil
				}
			},
			want: func(file string) string { return file + " is owned by effective uid 1000" },
		},
		"a writable file": {
			open: func(reads *closureReads, file string) {
				reads.writable = func(path string) (bool, error) { return path == file, nil }
			},
			want: func(file string) string { return file + " is writable by effective uid 1000" },
		},
		"CAP_DAC_OVERRIDE": {
			open: func(reads *closureReads, file string) {
				reads.status = func() ([]byte, error) { return []byte("CapEff:\t0000000000000002\n"), nil }
			},
			want: func(file string) string { return "the process holds CAP_DAC_OVERRIDE" },
		},
		"CAP_FOWNER": {
			open: func(reads *closureReads, file string) {
				reads.status = func() ([]byte, error) { return []byte("CapEff:\t0000000000000008\n"), nil }
			},
			want: func(file string) string { return "the process holds CAP_FOWNER" },
		},
		"CAP_CHOWN": {
			open: func(reads *closureReads, file string) {
				reads.status = func() ([]byte, error) { return []byte("CapEff:\t0000000000000001\n"), nil }
			},
			want: func(file string) string { return "the process holds CAP_CHOWN" },
		},
		"CAP_SYS_ADMIN": {
			open: func(reads *closureReads, file string) {
				reads.status = func() ([]byte, error) { return []byte("CapEff:\t0000000000200000\n"), nil }
			},
			want: func(file string) string { return "the process holds CAP_SYS_ADMIN" },
		},
		"a CapEff line that doesn't parse": {
			open: func(reads *closureReads, file string) {
				reads.status = func() ([]byte, error) { return []byte("CapEff:\tall of them\n"), nil }
			},
			want: func(file string) string { return "its capabilities can't be read" },
		},
		"a status that can't be read": {
			open: func(reads *closureReads, file string) {
				reads.status = func() ([]byte, error) { return nil, errors.New("no /proc here") }
			},
			want: func(file string) string { return "its capabilities can't be read" },
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := unpackedSourceForTest(t)
			file := filepath.Join(root, "source/a.c")
			reads := closedReadsForTest()
			test.open(&reads, file)
			state, err := unpackedTreeState(root, proofForTest(reads))
			if !errors.Is(err, ErrSourceNotClosed) || !strings.Contains(err.Error(), test.want(file)) {
				t.Fatalf("recorded %q, err %v, want a refusal saying %q", state, err, test.want(file))
			}
		})
	}
}

// The read-only mount proof holds only when every mount at or under the root is read-only, as mountinfo has them, and
// so is every mount of the root's filesystem that reaches the root's part of it: a mount point is under the root by
// whole path components, and its octal escapes are decoded. A root on no mount mountinfo names proves nothing.
// Mutant: mountinfo not asked, only the root's own mount: "a writable mount below" goes red.
// Mutant: under the root by string prefix: "a writable sibling" goes red.
// Mutant: escapes not decoded: "an escaped writable mount below" goes red.
// Mutant: matched by path only, the device pass skipped: "a read-only bind whose filesystem is mounted writable above"
// goes red.
// Mutant: a root on no mount named taken as proven: "the root's own mount missing" goes red.
func TestAReadOnlyMountHoldsOnlyWhenEveryMountUnderTheRootIsReadOnly(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		root, mountInfo string
		holds           bool
	}{
		"a writable mount below": {"/src", "22 1 0:21 / /src ro,relatime - ext4 /dev/sda1 ro\n" +
			"23 22 0:22 / /src/cache rw,relatime - tmpfs tmpfs rw\n", false},
		"a read-only mount below, a writable sibling and a writable root above": {"/src", "21 0 0:20 / / rw,relatime - ext4 /dev/sda2 rw\n" +
			"22 21 0:21 / /src ro,relatime - ext4 /dev/sda1 ro\n" +
			"23 22 0:22 / /src/cache ro,relatime - tmpfs tmpfs ro\n" +
			"24 21 0:23 / /src-other rw,relatime - ext4 /dev/sdb1 rw\n", true},
		"an escaped writable mount below": {"/src tree", "22 1 0:21 / /src\\040tree ro,relatime - ext4 /dev/sda1 ro\n" +
			"23 22 0:22 / /src\\040tree/cache rw,relatime - tmpfs tmpfs rw\n", false},
		"a read-only bind whose filesystem is mounted writable above": {"/var/lib/job/tree", "21 0 8:2 / / ro,relatime - ext4 /dev/sda2 ro\n" +
			"29 21 8:1 /job /var/lib/job rw,relatime - ext4 /dev/sda1 rw\n" +
			"30 21 8:1 /job/tree /var/lib/job/tree ro,relatime - ext4 /dev/sda1 rw\n", false},
		"a read-only bind whose filesystem is mounted writable for another subtree": {"/var/lib/job/tree", "21 0 8:2 / / ro,relatime - ext4 /dev/sda2 ro\n" +
			"29 21 8:1 /other /mnt/other rw,relatime - ext4 /dev/sda1 rw\n" +
			"30 21 8:1 /job/tree /var/lib/job/tree ro,relatime - ext4 /dev/sda1 rw\n", true},
		"a read-only bind and another device writable above": {"/var/lib/job/tree", "21 0 8:2 / / ro,relatime - ext4 /dev/sda2 ro\n" +
			"29 21 8:3 /job /var/lib/job rw,relatime - ext4 /dev/sdb1 rw\n" +
			"30 29 8:1 /job/tree /var/lib/job/tree ro,relatime - ext4 /dev/sda1 rw\n", true},
		"the root's own mount missing": {"/var/lib/job/tree", "29 1 8:1 /other /mnt/other ro,relatime - ext4 /dev/sda1 ro\n", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := mountsReadOnly(test.root, []byte(test.mountInfo)); (err == nil) != test.holds {
				t.Fatalf("held %t (%v), want %t", err == nil, err, test.holds)
			}
		})
	}
}

// A root on a read-only mount is proven by it only when mountinfo can be read, which is what proves no mount below it
// writable.
// Mutant: mountinfo that can't be read taken as no mount below: "unreadable" goes red.
func TestTheReadOnlyMountProofNeedsMountinfo(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		mountInfo func() ([]byte, error)
		want      string
	}{
		"unreadable":            {func() ([]byte, error) { return nil, errors.New("no /proc here") }, "/proc/self/mountinfo: no /proc here"},
		"read-only, none below": {func() ([]byte, error) { return []byte("21 0 8:1 / / ro,relatime - ext4 /dev/sda1 ro\n"), nil }, ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			reads := closedReadsForTest()
			reads.readOnlyMount = func(string) bool { return true }
			reads.mountInfo = test.mountInfo
			// Every path is writable, so a refused mount proof leaves no proof standing.
			reads.writable = func(string) (bool, error) { return true, nil }
			proof, err := closedProofWith(t.TempDir(), reads)
			if test.want == "" && (err != nil || proof != "read-only-mount") {
				t.Fatalf("proved %q, err %v, want read-only-mount", proof, err)
			}
			if test.want != "" && (err == nil || !strings.Contains(err.Error(), test.want)) {
				t.Fatalf("proved %q, err %v, want a refusal saying %q", proof, err, test.want)
			}
		})
	}
}

// A read-only mount proves nothing to a process that could remount it read-write: one holding CAP_SYS_ADMIN, or one
// whose capabilities can't be read, is refused by the mount proof, and with every path writable no proof stands.
// Mutant: the mount proof asks no capability: "CAP_SYS_ADMIN held" goes red.
func TestTheReadOnlyMountProofRefusesAProcessThatCanRemount(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		status func() ([]byte, error)
		want   string
	}{
		"CAP_SYS_ADMIN held":      {func() ([]byte, error) { return []byte("CapEff:\t0000000000200000\n"), nil }, "CAP_SYS_ADMIN"},
		"capabilities unreadable": {func() ([]byte, error) { return nil, errors.New("no /proc here") }, "its capabilities can't be read"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			reads := closedReadsForTest()
			reads.readOnlyMount = func(string) bool { return true }
			reads.mountInfo = func() ([]byte, error) { return []byte{}, nil }
			reads.status = test.status
			// Every path is writable, so a refused mount proof leaves no proof standing.
			reads.writable = func(string) (bool, error) { return true, nil }
			proof, err := closedProofWith(t.TempDir(), reads)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("proved %q, err %v, want a refusal saying %q", proof, err, test.want)
			}
		})
	}
}

// A real unpacked source this process may write, as t.TempDir leaves it, is refused on every platform, not recorded as
// its manifest: the process owns it, can write it, and off Linux nothing is proven at all. No live test proves a
// closed source, which needs a read-only mount or a tree another uid owns, neither of which a test can make unprivileged.
// Mutant: the proof skipped, so a closed source is claimed unasked: red.
func TestAWritableUnpackedSourceIsRefused(t *testing.T) {
	t.Parallel()
	state, err := treeStateAt(unpackedSourceForTest(t))
	if !errors.Is(err, ErrSourceNotClosed) {
		t.Fatalf("recorded %q, err %v", state, err)
	}
}

// A checkout carrying a manifest beside its .git is still asked through git, as every checkout is: its state starts
// with git's HEAD, and no proof of closure is asked of it.
// Mutant: treeStateAt takes the manifest's path whenever .tracked is there: red.
func TestACheckoutWithAManifestTakesGitsPath(t *testing.T) {
	t.Parallel()
	root := unpackedSourceForTest(t)
	gitIn(t, root, "init", "-q")
	gitIn(t, root, "add", "source")
	gitIn(t, root, "commit", "-q", "-m", "tree")
	head, err := toolchain.GitCommand("-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	state, err := treeStateAt(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(state, string(head)) || strings.Contains(state, "unpacked closed") {
		t.Fatalf("recorded\n%s\nwant git's HEAD %s first and no manifest", state, head)
	}
}
