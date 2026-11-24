package buildcache

import (
	"errors"
	"os"
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

// noCapabilities and oneUser are status's capability bits and ids for a process that holds no capability and is uid
// and gid 1000 every way (real, effective, saved, filesystem).
const noCapabilities = "0000000000000000"
const oneUser = "1000\t1000\t1000\t1000"

// statusForTest is /proc/self/status as the kernel writes it, cut to the lines the proofs read.
func statusForTest(effective, permitted, uids, gids string) func() ([]byte, error) {
	return func() ([]byte, error) {
		return []byte("Name:\ttraced\nUid:\t" + uids + "\nGid:\t" + gids + "\nCapPrm:\t" + permitted + "\nCapEff:\t" + effective + "\n"), nil
	}
}

// closedReadsForTest are the reads of a process closed out of every path under a tree: another uid owns all of it,
// none of it is writable, the process holds no capability and is uid 1000 every way, and the tree is on a writable
// mount, so only the no-write proof can hold. A test changes one read to open the tree.
func closedReadsForTest() closureReads {
	return closureReads{
		readOnlyMount: func(string) bool { return false },
		mountInfo:     func() ([]byte, error) { return nil, errors.New("this test reads no mountinfo") },
		status:        statusForTest(noCapabilities, noCapabilities, oneUser, oneUser),
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

// The no-write proof holds only when nothing under the root is the process's to change, now or once it uses what it
// holds: a file it owns (whose mode its owner can change back), a file it may write, a capability past both, effective
// or permitted, or ids it can switch among refuses, naming what was found, and so does a status that can't be read.
// Mutant: the owner not checked: "an owned file" goes red.
// Mutant: regular files not asked whether they are writable, only directories: "a writable file" goes red.
// Mutant: the capabilities not checked: every CAP_ case and the unreadable capability cases go red.
// Mutant: only CAP_DAC_OVERRIDE and CAP_FOWNER checked: "CAP_CHOWN", "CAP_SETGID", "CAP_SETUID" and "CAP_SYS_ADMIN" go
// red.
// Mutant: CapPrm not read, heldCapabilities' `return effective | permitted, nil` to `return effective, nil`:
// "CAP_DAC_OVERRIDE permitted" goes red.
// Mutant: the Uid and Gid lines' ids not compared, statusIdentity's `if id != ids[0] {` to `if false && id != ids[0] {`:
// "a saved uid apart" and "gids apart" go red.
func TestTheNoWriteProofRefusesWhatTheProcessCouldChange(t *testing.T) {
	t.Parallel()
	holding := func(effective, permitted string) func(reads *closureReads, file string) {
		return func(reads *closureReads, file string) {
			reads.status = statusForTest(effective, permitted, oneUser, oneUser)
		}
	}
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
			want: func(file string) string { return file + " is owned by uid 1000, the process's" },
		},
		"a writable file": {
			open: func(reads *closureReads, file string) {
				reads.writable = func(path string) (bool, error) { return path == file, nil }
			},
			want: func(file string) string { return file + " is writable by uid 1000" },
		},
		"CAP_DAC_OVERRIDE":           {open: holding("0000000000000002", noCapabilities), want: func(string) string { return "the process holds CAP_DAC_OVERRIDE" }},
		"CAP_DAC_OVERRIDE permitted": {open: holding(noCapabilities, "0000000000000002"), want: func(string) string { return "the process holds CAP_DAC_OVERRIDE" }},
		"CAP_FOWNER":                 {open: holding("0000000000000008", noCapabilities), want: func(string) string { return "the process holds CAP_FOWNER" }},
		"CAP_CHOWN":                  {open: holding("0000000000000001", noCapabilities), want: func(string) string { return "the process holds CAP_CHOWN" }},
		"CAP_SETGID":                 {open: holding("0000000000000040", noCapabilities), want: func(string) string { return "the process holds CAP_SETGID" }},
		"CAP_SETUID":                 {open: holding("0000000000000080", noCapabilities), want: func(string) string { return "the process holds CAP_SETUID" }},
		"CAP_SYS_ADMIN":              {open: holding("0000000000200000", noCapabilities), want: func(string) string { return "the process holds CAP_SYS_ADMIN" }},
		"a saved uid apart": {
			open: func(reads *closureReads, file string) {
				reads.status = statusForTest(noCapabilities, noCapabilities, "1000\t1000\t0\t1000", oneUser)
			},
			want: func(string) string { return "its Uid ids differ" },
		},
		"gids apart": {
			open: func(reads *closureReads, file string) {
				reads.status = statusForTest(noCapabilities, noCapabilities, oneUser, "1000\t1000\t1000\t27")
			},
			want: func(string) string { return "its Gid ids differ" },
		},
		"a Uid line that doesn't parse": {
			open: func(reads *closureReads, file string) {
				reads.status = statusForTest(noCapabilities, noCapabilities, "1000", oneUser)
			},
			want: func(string) string { return "its ids can't be read" },
		},
		"a CapEff line that doesn't parse": {
			open: func(reads *closureReads, file string) {
				reads.status = statusForTest("all of them", noCapabilities, oneUser, oneUser)
			},
			want: func(string) string { return "its capabilities can't be read" },
		},
		"no CapPrm line": {
			open: func(reads *closureReads, file string) {
				reads.status = func() ([]byte, error) {
					return []byte("Uid:\t" + oneUser + "\nGid:\t" + oneUser + "\nCapEff:\t" + noCapabilities + "\n"), nil
				}
			},
			want: func(string) string { return "its capabilities can't be read" },
		},
		"a status that can't be read": {
			open: func(reads *closureReads, file string) {
				reads.status = func() ([]byte, error) { return nil, errors.New("no /proc here") }
			},
			want: func(string) string { return "its capabilities and ids can't be read" },
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
// Mutant: the options read from the superblock's field, mountsReadOnly's `options: fields[5]})` to
// `options: fields[len(fields)-1]})`: "a writable mount of a read-only superblock" goes red, since the mount's own
// option, sixth, is the one that governs writes through it.
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
		"the root's own mount missing":               {"/var/lib/job/tree", "29 1 8:1 /other /mnt/other ro,relatime - ext4 /dev/sda1 ro\n", false},
		"a writable mount of a read-only superblock": {"/src", "22 1 0:21 / /src rw,relatime - ext4 /dev/sda1 ro\n", false},
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

// A read-only mount proves nothing to a process that could remount it read-write: one holding CAP_SYS_ADMIN, effective
// or permitted, or one whose capabilities can't be read, is refused by the mount proof itself, asked alone over a
// mountinfo whose every mount is read-only, so no other proof's words can answer for it.
// Mutant: readOnlyMountProof's `if capabilities&(1<<sysAdminCapability) != 0 {` to
// `if false && capabilities&(1<<sysAdminCapability) != 0 {`: both CAP_SYS_ADMIN cases go red.
// Mutant: CapPrm not read (heldCapabilities): "CAP_SYS_ADMIN permitted" goes red.
func TestTheReadOnlyMountProofRefusesAProcessThatCanRemount(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		status func() ([]byte, error)
		want   string
	}{
		"CAP_SYS_ADMIN held":      {statusForTest("0000000000200000", noCapabilities, oneUser, oneUser), "which can remount it read-write"},
		"CAP_SYS_ADMIN permitted": {statusForTest(noCapabilities, "0000000000200000", oneUser, oneUser), "which can remount it read-write"},
		"capabilities unreadable": {func() ([]byte, error) { return nil, errors.New("no /proc here") }, "its capabilities can't be read: no /proc here"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			reads := closedReadsForTest()
			reads.readOnlyMount = func(string) bool { return true }
			reads.mountInfo = func() ([]byte, error) { return []byte("21 0 8:1 / / ro,relatime - ext4 /dev/sda1 ro\n"), nil }
			reads.status = test.status
			if err := readOnlyMountProof("/src", reads); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("the mount proof answered %v, want a refusal saying %q", err, test.want)
			}
		})
	}
}

// A root that is a link is proven over the tree it names, not as the link alone, which the walk would never descend:
// a link to a tree the process may write, or owns, refuses.
// Mutant: the no-write proof given the unresolved root, closedProofWith's `uid, err := noWriteProof(resolved, reads)`
// to `uid, err := noWriteProof(root, reads)`: both cases go red.
func TestARootThatIsALinkIsProvenOverItsTree(t *testing.T) {
	t.Parallel()
	for name, open := range map[string]func(reads *closureReads, link string){
		"a writable tree": func(reads *closureReads, link string) {
			reads.writable = func(string) (bool, error) { return true, nil }
		},
		"an owned tree": func(reads *closureReads, link string) {
			reads.owner = func(path string) (int, error) {
				if path == link {
					return 2000, nil
				}
				return 1000, nil
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			link := filepath.Join(t.TempDir(), "source")
			if err := os.Symlink(unpackedSourceForTest(t), link); err != nil {
				t.Fatal(err)
			}
			reads := closedReadsForTest()
			open(&reads, link)
			if proof, err := closedProofWith(link, reads); err == nil {
				t.Fatalf("proved %q over a link to a tree the process could change", proof)
			}
		})
	}
}

// A manifest whose submodule's directory holds its files but no HEAD is refused, naming the submodule, rather than
// recorded without its commit.
// Mutant: manifestHeads' check of every repository's HEAD skipped: red.
func TestAManifestMissingASubmodulesHeadIsRefused(t *testing.T) {
	t.Parallel()
	root := unpackedSourceForTest(t)
	if err := os.Remove(filepath.Join(root, ".tracked/module/HEAD")); err != nil {
		t.Fatal(err)
	}
	if state, err := unpackedTreeState(root, proofForTest(closedReadsForTest())); err == nil || !strings.Contains(err.Error(), "no commit for module") {
		t.Fatalf("recorded %q, err %v", state, err)
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
