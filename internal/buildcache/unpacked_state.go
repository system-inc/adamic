package buildcache

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// ErrSourceNotClosed is TreeState's refusal of an unpacked source not proven closed to this process, wrapped with why:
// cmd/traced tells it from TreeState's other errors, since such a source can't be run traced at all.
var ErrSourceNotClosed = errors.New("an unpacked source records its state only when closed to this process")

// unpackedTreeState is TreeState for an unpacked source (unpackedSource), which no git answers for (#3z7bv9q). Its
// manifest says what commit each repository was unpacked from, never what the source holds now, so the manifest's
// heads stand for the tree only while nothing under root can be written: the state is recorded only when proof, which
// is closedProof outside a test, proves the source closed to this process, and refused otherwise. The record is the
// proof that held, the commit of each repository the manifest names, and dpkg's line. TreeState is only ever held
// against another taken in the same run, so the record need not read as a checkout's does.
func unpackedTreeState(root string, proof func(root string) (string, error)) (string, error) {
	var state strings.Builder
	var err error
	keying(func() {
		var held string
		if held, err = proof(root); err != nil {
			err = fmt.Errorf("%w: %w", ErrSourceNotClosed, err)
			return
		}
		state.WriteString("unpacked closed " + held + "\n")
		var heads []string
		if heads, err = manifestHeads(root); err != nil {
			return
		}
		for _, head := range heads {
			state.WriteString(head + "\n")
		}
		state.WriteString(dpkgState())
	})
	return state.String(), err
}

// closureReads are what closedProofWith reads of this machine, each its own field so a test can answer for it:
// closed_linux.go's closedProof hands it the machine's own.
type closureReads struct {
	// readOnlyMount says the mount root is on is read-only (statfs's ST_RDONLY).
	readOnlyMount func(root string) bool
	// mountInfo is /proc/self/mountinfo's content.
	mountInfo func() ([]byte, error)
	// status is /proc/self/status's content, whose CapEff line holds the process's effective capabilities.
	status func() ([]byte, error)
	// effectiveUid is the process's effective uid, which faccessat2 with AT_EACCESS asks for.
	effectiveUid int
	// owner is the uid owning path, the path itself and never what a link names.
	owner func(path string) (int, error)
	// writable says whether the process's effective uid and groups may write path (faccessat2 with AT_EACCESS).
	writable func(path string) (bool, error)
}

// reopeningCapabilities are the capabilities that would let the process write what access says it can't, or reopen it
// to writing, by capability.h's bit numbers: CAP_CHOWN takes any file for its own, whose mode it may then change as
// the owner; CAP_DAC_OVERRIDE writes past any mode bit; CAP_FOWNER changes any file's mode, as its owner could; and
// CAP_SYS_ADMIN remounts a read-only mount read-write.
var reopeningCapabilities = []struct {
	bit  uint
	name string
}{{0, "CAP_CHOWN"}, {1, "CAP_DAC_OVERRIDE"}, {3, "CAP_FOWNER"}, {sysAdminCapability, "CAP_SYS_ADMIN"}}

// sysAdminCapability is CAP_SYS_ADMIN's bit, which both proofs refuse: it remounts a read-only mount read-write.
const sysAdminCapability = 21

// closedProofWith is how root is proven closed to the process reads describes, as unpackedTreeState's record names it,
// or why it isn't (Loom's ruling on #3z7bv9q). Either root is on a read-only mount and so is every mount at or under
// it (readOnlyMountProof), or nothing under root can be written by the process nor made writable by it (noWriteProof).
func closedProofWith(root string, reads closureReads) (string, error) {
	mountError := readOnlyMountProof(root, reads)
	if mountError == nil {
		return "read-only-mount", nil
	}
	if err := noWriteProof(root, reads); err != nil {
		return "", fmt.Errorf("not on a read-only mount (%v), and %v", mountError, err)
	}
	return "no-write uid " + strconv.Itoa(reads.effectiveUid), nil
}

// readOnlyMountProof is why root isn't proven on a read-only mount, or nil when it is: its own mount is read-only, and
// so is every mount at or under it, by mountinfo, a mount below root being one statfs on root never sees. Root's links
// are resolved first, as mountinfo names every mount point resolved; mountinfo that can't be read proves nothing. A
// process holding CAP_SYS_ADMIN could remount the tree read-write, write and remount it read-only again, so a mount
// proves nothing to it, and capabilities that can't be read prove nothing either.
func readOnlyMountProof(root string, reads closureReads) error {
	if !reads.readOnlyMount(root) {
		return errors.New("its mount is writable")
	}
	status, err := reads.status()
	if err != nil {
		return fmt.Errorf("its capabilities can't be read: %v", err)
	}
	capabilities, ok := effectiveCapabilities(string(status))
	if !ok {
		return errors.New("its capabilities can't be read: /proc/self/status holds no CapEff line it can parse")
	}
	if capabilities&(1<<sysAdminCapability) != 0 {
		return errors.New("the process holds CAP_SYS_ADMIN, which can remount it read-write")
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	mountInfo, err := reads.mountInfo()
	if err != nil {
		return fmt.Errorf("/proc/self/mountinfo: %v", err)
	}
	return mountsReadOnly(filepath.Clean(resolved), mountInfo)
}

// mountsReadOnly is why some mount at or under root, a resolved and clean path, isn't read-only, or nil when every one
// is. Each line of mountInfo is one mount: its fifth field the mount point, its sixth the mount's own options, ro or rw
// among them. "Under" is by whole path components (under), so /src-other is no mount under /src.
func mountsReadOnly(root string, mountInfo []byte) error {
	for _, line := range strings.Split(string(mountInfo), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			return fmt.Errorf("/proc/self/mountinfo holds a line it can't read: %q", line)
		}
		point := filepath.Clean(unescapeMountPoint(fields[4]))
		if under(point, root) && !slices.Contains(strings.Split(fields[5], ","), "ro") {
			return fmt.Errorf("%s is mounted %s", point, fields[5])
		}
	}
	return nil
}

// unescapeMountPoint is a mount point as mountinfo writes it, with the kernel's octal escapes (a space as \040, a tab
// as \011, a newline as \012, a backslash as \134) turned back into their bytes.
func unescapeMountPoint(escaped string) string {
	var point strings.Builder
	for index := 0; index < len(escaped); index++ {
		if escaped[index] == '\\' && index+4 <= len(escaped) {
			if value, err := strconv.ParseUint(escaped[index+1:index+4], 8, 8); err == nil {
				point.WriteByte(byte(value))
				index += 3
				continue
			}
		}
		point.WriteByte(escaped[index])
	}
	return point.String()
}

// noWriteProof is why the process isn't proven unable to write anything under root, or nil when it is. Three things
// must hold. Its effective capabilities hold none that reopens the tree (reopeningCapabilities: CAP_CHOWN,
// CAP_DAC_OVERRIDE, CAP_FOWNER, CAP_SYS_ADMIN), read from status's CapEff line, which proves nothing when it can't be
// read. Nothing under root is the effective uid's, since an owner can chmod its
// own file writable again; every entry is asked, links and all. And the kernel answers that the process may write no
// directory and no regular file under root, the manifest included, since a test that could rewrite .tracked could move
// the very heads the record holds. A link's own text changes only through its directory, which is asked, and nothing
// else under root (a socket, a device) holds any of the tree's content.
func noWriteProof(root string, reads closureReads) error {
	status, err := reads.status()
	if err != nil {
		return fmt.Errorf("its capabilities can't be read: %v", err)
	}
	capabilities, ok := effectiveCapabilities(string(status))
	if !ok {
		return errors.New("its capabilities can't be read: /proc/self/status holds no CapEff line it can parse")
	}
	for _, capability := range reopeningCapabilities {
		if capabilities&(1<<capability.bit) != 0 {
			return fmt.Errorf("the process holds %s", capability.name)
		}
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		owner, err := reads.owner(path)
		if err != nil {
			return fmt.Errorf("who owns %s: %v", path, err)
		}
		if owner == reads.effectiveUid {
			return fmt.Errorf("%s is owned by effective uid %d", path, reads.effectiveUid)
		}
		if !entry.IsDir() && !entry.Type().IsRegular() {
			return nil
		}
		writable, err := reads.writable(path)
		if err != nil {
			return fmt.Errorf("whether %s is writable: %v", path, err)
		}
		if writable {
			return fmt.Errorf("%s is writable by effective uid %d", path, reads.effectiveUid)
		}
		return nil
	})
}

// effectiveCapabilities is the CapEff line of status, /proc/self/status's content, as bits, and whether there is one
// that parses.
func effectiveCapabilities(status string) (uint64, bool) {
	for _, line := range strings.Split(status, "\n") {
		if value, ok := strings.CutPrefix(line, "CapEff:"); ok {
			capabilities, err := strconv.ParseUint(strings.TrimSpace(value), 16, 64)
			return capabilities, err == nil
		}
	}
	return 0, false
}

// manifestHeads is a line for each repository the manifest at root names, sorted by the repository's path ("." for the
// tree's own): head <path> <the commit its HEAD file holds>. The manifest keeps one directory per repository at its
// path under trackedManifestDirectory (internal/tracked's package comment), so each HEAD file under it is one
// repository's. A manifest naming no commit for the tree's own is refused: its record would hold no head at all.
func manifestHeads(root string) ([]string, error) {
	manifest := filepath.Join(root, trackedManifestDirectory)
	heads := map[string]string{}
	err := filepath.WalkDir(manifest, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.Type().IsRegular() || entry.Name() != "HEAD" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		repository, err := filepath.Rel(manifest, filepath.Dir(path))
		if err != nil {
			return err
		}
		heads[filepath.ToSlash(repository)] = strings.TrimSpace(string(content))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("the unpacked source's manifest: %v", err)
	}
	if _, ok := heads["."]; !ok {
		return nil, fmt.Errorf("the unpacked source's manifest names no commit for the tree's own (%s/HEAD)", trackedManifestDirectory)
	}
	var lines []string
	for _, repository := range slices.Sorted(maps.Keys(heads)) {
		lines = append(lines, "head "+repository+" "+heads[repository])
	}
	return lines, nil
}
