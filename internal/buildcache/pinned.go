package buildcache

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
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
		command.Stdin = bytes.NewReader(stdin)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		if output, err = command.Output(); err != nil {
			err = fmt.Errorf("pinned git %s for %s: %w: %s", strings.Join(arguments, " "), inputs.Name, err, strings.TrimSpace(stderr.String()))
		}
	})
	return output, err
}
