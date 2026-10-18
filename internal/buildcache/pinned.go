package buildcache

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/toolchain"
)

// commitPattern is a full commit id: a pin names its commit whole, never by a prefix a later object could share.
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

// pinnedKeying runs PinnedGit's git as key-time work: keying, swapped by a test that asks whether it ran there.
var pinnedKeying = keying

// PinnedGit runs git with arguments on root's own objects as key-time work (Keying), for a product built from a pinned
// commit's files (json's pinned corpus, #jwn3n9q). A traced build may not read .git, since nothing there is a file a key
// can name; but a commit id names its tree and every blob under it, so a key that names the commit names all of it, and
// the build reads git's answer, never .git. It refuses a commit that isn't a full 40-hex id, or that no flag of inputs
// names, so the product's key always moves with the pin. git runs with replace objects and grafts off, no system or
// global configuration and no transport (protocol.allow=never, GIT_NO_LAZY_FETCH), so its answer about the commit is
// the commit's alone: a replace ref or a graft could answer with another commit's bytes under this id, a user's config
// could change how it prints, and a fetch would read the network. arguments must ask only about commit and the objects
// under it; stdin, when given, is git's standard input.
func PinnedGit(root string, inputs Inputs, commit string, stdin []byte, arguments ...string) ([]byte, error) {
	if !commitPattern.MatchString(commit) {
		return nil, fmt.Errorf("pinned git for %s: %q isn't a full commit id", inputs.Name, commit)
	}
	if !slices.ContainsFunc(inputs.Flags, func(flag string) bool { return strings.Contains(flag, commit) }) {
		return nil, fmt.Errorf("pinned git for %s: no flag of its key names commit %s, so its key wouldn't move with the pin", inputs.Name, commit)
	}
	var output []byte
	var err error
	pinnedKeying(func() {
		command := toolchain.GitCommandIn(root, append([]string{"-c", "protocol.allow=never", "-c", "core.quotePath=true"}, arguments...)...)
		if command.Err != nil {
			err = command.Err
			return
		}
		command.Env = append(command.Env, "GIT_NO_REPLACE_OBJECTS=1", "GIT_GRAFT_FILE=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_NO_LAZY_FETCH=1")
		// The caller's trace2 event target, when it sets one, goes through: it observes git and can't change its answer,
		// and json's tests read it to prove the corpus never fetches.
		if target := os.Getenv("GIT_TRACE2_EVENT"); target != "" {
			command.Env = append(command.Env, "GIT_TRACE2_EVENT="+target)
		}
		command.Stdin = bytes.NewReader(stdin)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		if output, err = command.Output(); err != nil {
			err = fmt.Errorf("pinned git %s for %s: %w: %s", strings.Join(arguments, " "), inputs.Name, err, strings.TrimSpace(stderr.String()))
		}
	})
	return output, err
}

// A PinnedFile is one blob of a pinned commit's tree: its path in the tree and its bytes.
type PinnedFile struct {
	Path string
	Data []byte
}

// PinnedFiles reads every blob of pin's tree whose path keep accepts, sorted by path, from root's own objects through
// PinnedGit: a corpus that moves only when its pin does, never with the working tree (json's, #jwn3n9q; tsprinter's,
// #wy8gcke). inputs must name pin in a flag, so the product built from these bytes keys by it. The corpus never
// fetches: a commit the checkout lacks, or a blob a blobless clone never wrote, is the checkout's to carry, and the
// error names it. corpus names the files in errors ("JSON" says "repository JSON pin ... its 10 JSON files").
func PinnedFiles(root string, inputs Inputs, pin, corpus string, keep func(path string) bool) ([]PinnedFile, error) {
	if !commitPattern.MatchString(pin) {
		return nil, fmt.Errorf("repository %s: invalid commit pin %q", corpus, pin)
	}
	if _, err := PinnedGit(root, inputs, pin, nil, "cat-file", "-e", pin+"^{commit}"); err != nil {
		return nil, fmt.Errorf("repository %s pin %s: environment: the commit is not in the checkout at %s, and the corpus never fetches; a shallow checkout carries it with git fetch --depth=1 origin %s before the run (%w)", corpus, pin, root, pin, err)
	}
	tree, err := PinnedGit(root, inputs, pin, nil, "ls-tree", "-r", "-z", pin)
	if err != nil {
		return nil, fmt.Errorf("repository %s pin %s: %w", corpus, pin, err)
	}
	var names []string
	blobs := map[string]string{}
	for _, entry := range strings.Split(string(tree), "\x00") {
		// <mode> <type> <object>TAB<path>
		header, name, found := strings.Cut(entry, "\t")
		fields := strings.Fields(header)
		if found && len(fields) == 3 && fields[1] == "blob" && keep(name) {
			names = append(names, name)
			blobs[name] = fields[2]
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("repository %s pin %s has no %s files", corpus, pin, corpus)
	}
	sort.Strings(names)
	if err := pinnedBlobsPresent(root, inputs, pin, corpus, names, blobs); err != nil {
		return nil, err
	}
	var requests strings.Builder
	for _, name := range names {
		if strings.ContainsAny(name, "\r\n") {
			return nil, fmt.Errorf("repository %s pin %s: unsupported newline in path %q", corpus, pin, name)
		}
		fmt.Fprintf(&requests, "%s:%s\n", pin, name)
	}
	batch, err := PinnedGit(root, inputs, pin, []byte(requests.String()), "cat-file", "--batch")
	if err != nil {
		return nil, fmt.Errorf("repository %s pin %s: %w", corpus, pin, err)
	}
	reader := bufio.NewReader(bytes.NewReader(batch))
	files := make([]PinnedFile, 0, len(names))
	for _, name := range names {
		header, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("repository %s pin %s path %s: %w", corpus, pin, name, err)
		}
		fields := strings.Fields(header)
		if len(fields) != 3 || fields[1] != "blob" {
			return nil, fmt.Errorf("repository %s pin %s path %s: invalid blob header %q", corpus, pin, name, header)
		}
		size, err := strconv.Atoi(fields[2])
		if err != nil || size < 0 {
			return nil, fmt.Errorf("repository %s pin %s path %s: invalid blob size %q", corpus, pin, name, fields[2])
		}
		data := make([]byte, size)
		if _, err := io.ReadFull(reader, data); err != nil {
			return nil, fmt.Errorf("repository %s pin %s path %s: %w", corpus, pin, name, err)
		}
		newline, err := reader.ReadByte()
		if err != nil || newline != '\n' {
			return nil, fmt.Errorf("repository %s pin %s path %s: missing blob terminator", corpus, pin, name)
		}
		files = append(files, PinnedFile{Path: name, Data: data})
	}
	if _, err := reader.ReadByte(); err != io.EOF {
		return nil, fmt.Errorf("repository %s pin %s: trailing batch output", corpus, pin)
	}
	return files, nil
}

// Loom's pool units are blobless clones (--filter=blob:none): the pin's commit and trees are there, but only the
// blobs a checkout wrote. cat-file would try to fetch each missing one on its own, which PinnedGit refuses; this
// names every file whose blob is missing first, so the failure says which inputs the checkout lacks.
func pinnedBlobsPresent(root string, inputs Inputs, pin, corpus string, names []string, blobs map[string]string) error {
	// --no-walk: the pin's own tree, not every object reachable through its history. --missing=print lists an absent
	// object rather than fetching it.
	objects, err := PinnedGit(root, inputs, pin, nil, "rev-list", "--objects", "--no-walk", "--missing=print", pin)
	if err != nil {
		return fmt.Errorf("repository %s pin %s: %w", corpus, pin, err)
	}
	absent := map[string]bool{}
	for _, line := range strings.Split(string(objects), "\n") {
		if object, found := strings.CutPrefix(line, "?"); found {
			absent[object] = true
		}
	}
	var missing []string
	for _, name := range names {
		if absent[blobs[name]] {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	named := strings.Join(missing, ", ")
	if len(missing) > 5 {
		named = strings.Join(missing[:5], ", ") + fmt.Sprintf(" and %d more", len(missing)-5)
	}
	return fmt.Errorf("repository %s pin %s: environment: %d of its %d %s files have no blob in the checkout at %s, and the corpus never fetches: %s", corpus, pin, len(missing), len(names), corpus, root, named)
}
