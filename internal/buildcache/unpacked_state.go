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
	// status is /proc/self/status's content: its CapEff and CapPrm lines hold the process's effective and permitted
	// capabilities, its Uid and Gid lines its real, effective, saved and filesystem ids.
	status func() ([]byte, error)
	// owner is the uid owning path, the path itself and never what a link names.
	owner func(path string) (int, error)
	// writable says whether the process's effective uid and groups may write path (faccessat2 with AT_EACCESS).
	writable func(path string) (bool, error)
}

// reopeningCapabilities are the capabilities that would let the process write what access says it can't, or reopen it
// to writing, by capability.h's bit numbers: CAP_CHOWN takes any file for its own, whose mode it may then change as
// the owner; CAP_DAC_OVERRIDE writes past any mode bit; CAP_FOWNER changes any file's mode, as its owner could;
// CAP_SETGID and CAP_SETUID become any group or user, the owner among them; and CAP_SYS_ADMIN remounts a read-only
// mount read-write.
var reopeningCapabilities = []struct {
	bit  uint
	name string
}{{0, "CAP_CHOWN"}, {1, "CAP_DAC_OVERRIDE"}, {3, "CAP_FOWNER"}, {6, "CAP_SETGID"}, {7, "CAP_SETUID"}, {sysAdminCapability, "CAP_SYS_ADMIN"}}

// sysAdminCapability is CAP_SYS_ADMIN's bit, which both proofs refuse: it remounts a read-only mount read-write.
const sysAdminCapability = 21

// closedProofWith is how root is proven closed to the process reads describes, as unpackedTreeState's record names it,
// or why it isn't (Loom's ruling on #3z7bv9q). Root is resolved first, once, and both proofs take the resolved root: a
// root that is a link would otherwise be walked as the link alone, and mountinfo names every mount point resolved.
// Either root is on a read-only mount and so is every mount at or under it or of its part of the filesystem
// (readOnlyMountProof), or nothing under root can be written by the process nor made writable by it (noWriteProof).
//
// One gap is named rather than closed: root's ancestors are asked nothing, so a process that may write root's parent
// can rename root away and put another tree in its place.
func closedProofWith(root string, reads closureReads) (string, error) {
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolving %s: %v", root, err)
	}
	resolved = filepath.Clean(resolved)
	mountError := readOnlyMountProof(resolved, reads)
	if mountError == nil {
		return "read-only-mount", nil
	}
	uid, err := noWriteProof(resolved, reads)
	if err != nil {
		return "", fmt.Errorf("not on a read-only mount (%v), and %v", mountError, err)
	}
	return "no-write uid " + strconv.Itoa(uid), nil
}

// readOnlyMountProof is why root, resolved and clean, isn't proven on a read-only mount, or nil when it is: its own
// mount is read-only, and so is every mount at or under it and every other mount of its part of the filesystem, by
// mountinfo (mountsReadOnly), each one a mount statfs on root never sees. Mountinfo that can't be read proves nothing.
// A process holding CAP_SYS_ADMIN, effective or permitted, could remount the tree read-write, write and remount it
// read-only again, so a mount proves nothing to it, and capabilities that can't be read prove nothing either.
func readOnlyMountProof(root string, reads closureReads) error {
	if !reads.readOnlyMount(root) {
		return errors.New("its mount is writable")
	}
	status, err := reads.status()
	if err != nil {
		return fmt.Errorf("its capabilities can't be read: %v", err)
	}
	capabilities, err := heldCapabilities(string(status))
	if err != nil {
		return err
	}
	if capabilities&(1<<sysAdminCapability) != 0 {
		return errors.New("the process holds CAP_SYS_ADMIN, which can remount it read-write")
	}
	mountInfo, err := reads.mountInfo()
	if err != nil {
		return fmt.Errorf("/proc/self/mountinfo: %v", err)
	}
	return mountsReadOnly(root, mountInfo)
}

// A mountLine is one line of /proc/self/mountinfo, as mountsReadOnly reads it: its third field, the filesystem's
// device (major:minor), its fourth, the directory of that filesystem mounted, its fifth, the mount point, and its
// sixth, the mount's own options, ro or rw among them. The two paths are unescaped and clean.
type mountLine struct {
	device, filesystemRoot, point, options string
}

// mountsReadOnly is why root, a resolved and clean path, can be written through some mount, or nil when it can't. Every
// mount at or under root must be read-only, and so must every mount of root's own filesystem that reaches root's part
// of it: a read-only bind of /var/lib/job/tree proves nothing while its filesystem is mounted read-write at
// /var/lib/job, through which every file is rewritten. Root's own mount is the one at the longest mount point root is at
// or under (of two at one point, the later, mounted over the earlier, is the one root's path reaches), and with none,
// nothing is proven. A mount of the same device reaches root's part when the directory it mounts is root's directory
// in the filesystem, above it or inside it. "Under" is by whole path components (under), so /src-other is no mount
// under /src.
func mountsReadOnly(root string, mountInfo []byte) error {
	var mounts []mountLine
	for _, line := range strings.Split(string(mountInfo), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			return fmt.Errorf("/proc/self/mountinfo holds a line it can't read: %q", line)
		}
		mounts = append(mounts, mountLine{device: fields[2], filesystemRoot: filepath.Clean(unescapeMountPoint(fields[3])),
			point: filepath.Clean(unescapeMountPoint(fields[4])), options: fields[5]})
	}
	own := -1
	for index, mount := range mounts {
		if under(root, mount.point) && (own < 0 || len(mount.point) >= len(mounts[own].point)) {
			own = index
		}
	}
	if own < 0 {
		return fmt.Errorf("/proc/self/mountinfo names no mount %s is on", root)
	}
	relative, err := filepath.Rel(mounts[own].point, root)
	if err != nil {
		return err
	}
	rootInFilesystem := filepath.Join(mounts[own].filesystemRoot, relative)
	for _, mount := range mounts {
		if slices.Contains(strings.Split(mount.options, ","), "ro") {
			continue
		}
		if under(mount.point, root) {
			return fmt.Errorf("%s is mounted %s", mount.point, mount.options)
		}
		if mount.device == mounts[own].device && (under(mount.filesystemRoot, rootInFilesystem) || under(rootInFilesystem, mount.filesystemRoot)) {
			return fmt.Errorf("%s mounts the filesystem %s is on %s", mount.point, root, mount.options)
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

// noWriteProof is the uid the process is proven unable to write anything under root as, or why it isn't proven. Four
// things must hold, each one what the process can't do, never only what it doesn't do now. Its capabilities, effective
// or permitted (which it may raise to effective), hold none that reopens the tree (reopeningCapabilities: CAP_CHOWN,
// CAP_DAC_OVERRIDE, CAP_FOWNER, CAP_SETGID, CAP_SETUID, CAP_SYS_ADMIN), read from status's CapEff and CapPrm lines.
// Its real, effective, saved and filesystem uids are one uid, and its gids one gid, read from status's Uid and Gid
// lines, since a process holding another id among them can switch to it. Either read proves nothing when it can't be
// read. Nothing under root is that uid's, since an owner can chmod its
// own file writable again; every entry is asked, links and all. And the kernel answers that the process may write no
// directory and no regular file under root, the manifest included, since a test that could rewrite .tracked could move
// the very heads the record holds. A link's own text changes only through its directory, which is asked, and nothing
// else under root (a socket, a device) holds any of the tree's content. One blind spot is named rather than closed: a
// link under root whose target is outside it lets what is read through the link change at the target, which no entry
// under root shows. The git path has the same spot, since status sees a link's text and never its target's change.
func noWriteProof(root string, reads closureReads) (int, error) {
	status, err := reads.status()
	if err != nil {
		return 0, fmt.Errorf("its capabilities and ids can't be read: %v", err)
	}
	capabilities, err := heldCapabilities(string(status))
	if err != nil {
		return 0, err
	}
	for _, capability := range reopeningCapabilities {
		if capabilities&(1<<capability.bit) != 0 {
			return 0, fmt.Errorf("the process holds %s", capability.name)
		}
	}
	uid, err := statusIdentity(string(status), "Uid")
	if err != nil {
		return 0, err
	}
	if _, err := statusIdentity(string(status), "Gid"); err != nil {
		return 0, err
	}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		owner, err := reads.owner(path)
		if err != nil {
			return fmt.Errorf("who owns %s: %v", path, err)
		}
		if owner == uid {
			return fmt.Errorf("%s is owned by uid %d, the process's", path, uid)
		}
		if !entry.IsDir() && !entry.Type().IsRegular() {
			return nil
		}
		writable, err := reads.writable(path)
		if err != nil {
			return fmt.Errorf("whether %s is writable: %v", path, err)
		}
		if writable {
			return fmt.Errorf("%s is writable by uid %d", path, uid)
		}
		return nil
	})
	return uid, err
}

// heldCapabilities is every capability status, /proc/self/status's content, says the process holds or may raise: its
// CapEff line's bits and its CapPrm line's, or why they can't be read.
func heldCapabilities(status string) (uint64, error) {
	var effective, permitted uint64
	for _, read := range []struct {
		name string
		bits *uint64
	}{{"CapEff", &effective}, {"CapPrm", &permitted}} {
		value, ok := statusLine(status, read.name)
		if !ok {
			return 0, fmt.Errorf("its capabilities can't be read: /proc/self/status holds no %s line", read.name)
		}
		bits, err := strconv.ParseUint(value, 16, 64)
		if err != nil {
			return 0, fmt.Errorf("its capabilities can't be read: /proc/self/status's %s line: %v", read.name, err)
		}
		*read.bits = bits
	}
	return effective | permitted, nil
}

// statusIdentity is the one id status's Uid or Gid line (name) holds as the process's real, effective, saved and
// filesystem id, or why there isn't one: the line is missing or doesn't parse, or its four ids differ, so the process
// can switch among them.
func statusIdentity(status, name string) (int, error) {
	value, ok := statusLine(status, name)
	if !ok {
		return 0, fmt.Errorf("its ids can't be read: /proc/self/status holds no %s line", name)
	}
	ids := strings.Fields(value)
	if len(ids) != 4 {
		return 0, fmt.Errorf("its ids can't be read: /proc/self/status's %s line holds %q, not four ids", name, value)
	}
	for _, id := range ids {
		if id != ids[0] {
			return 0, fmt.Errorf("its %s ids differ (real, effective, saved, filesystem: %s), so it can switch among them", name, value)
		}
	}
	id, err := strconv.ParseUint(ids[0], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("its ids can't be read: /proc/self/status's %s line: %v", name, err)
	}
	return int(id), nil
}

// statusLine is the value of status's line for name, its whitespace trimmed, and whether there is one.
func statusLine(status, name string) (string, bool) {
	for _, line := range strings.Split(status, "\n") {
		if value, ok := strings.CutPrefix(line, name+":"); ok {
			return strings.TrimSpace(value), true
		}
	}
	return "", false
}

// manifestHeads is a line for each repository the manifest at root names, sorted by the repository's path ("." for the
// tree's own): head <path> <the commit its HEAD file holds>. The manifest keeps one directory per repository at its
// path under trackedManifestDirectory (internal/tracked's package comment), so each HEAD file under it is one
// repository's. A manifest naming no commit for the tree's own is refused: its record would hold no head at all. So is
// a repository's directory, one holding its commit or files, with no HEAD beside them: its commit would go unrecorded,
// and a change to it unseen. A directory holding none of the three is only a step on a deeper repository's path.
func manifestHeads(root string) ([]string, error) {
	manifest := filepath.Join(root, trackedManifestDirectory)
	heads := map[string]string{}
	repositories := map[string]bool{}
	err := filepath.WalkDir(manifest, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.Type().IsRegular() || (entry.Name() != "HEAD" && entry.Name() != "commit" && entry.Name() != "files") {
			return nil
		}
		repository, err := filepath.Rel(manifest, filepath.Dir(path))
		if err != nil {
			return err
		}
		repositories[filepath.ToSlash(repository)] = true
		if entry.Name() != "HEAD" {
			return nil
		}
		content, err := os.ReadFile(path)
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
	for _, repository := range slices.Sorted(maps.Keys(repositories)) {
		if _, ok := heads[repository]; !ok {
			return nil, fmt.Errorf("the unpacked source's manifest names no commit for %s (%s/HEAD)", repository, filepath.Join(trackedManifestDirectory, repository))
		}
	}
	var lines []string
	for _, repository := range slices.Sorted(maps.Keys(heads)) {
		lines = append(lines, "head "+repository+" "+heads[repository])
	}
	return lines, nil
}
