package lane

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// The adapted tree is the pinned TypeScript checkout with stage3's adaptations applied: 92,000 files, 970 MB, and
// about 158 s of every phase:stage3 run, which built it from nothing each time (#4gp2ytz). It moves only when the pin
// or an adaptation does, so it is a product of its own, built ahead and fetched. Its build never reads the gate input's
// files: buildcache.PinnedGit asks the input's git for the pinned commit's objects as a pack, as key-time work, and
// the key names the commit in a flag, so the key moves with the pin and the read set holds the commit, the adaptation
// scripts, stage3/api's packages and the tools, never 81,500 walked files. The pack is indexed into a repository of
// the product's own, which checks the pin out detached, so the tree's HEAD is the pinned commit itself, as
// stage3/lane/check.py's API check reads it, and stage3/apply.py --adapt-checkout adapts it in place.
func adaptedTree(t testing.TB) string {
	t.Helper()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("ADAMIC_TYPESCRIPT_SOURCE names no checkout of the pinned TypeScript, the gate input the adapted tree is made from")
	}
	pin := typescriptPin(t, repository)
	inputs := buildcache.Inputs{
		Name: "stage3-adapted-tree",
		Files: []string{"stage3/source.json", "stage3/apply.py", "stage3/caches.py", "stage3/adapt",
			"stage3/api/package.json", "stage3/api/package-lock.json", "stage3/api/node_modules.sha256"},
		Flags:     []string{"typescript " + pin},
		Toolchain: []string{buildcache.Tool("git", "--version"), buildcache.Tool("node", "--version"), buildcache.Tool("python3", "--version")},
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		objects, err := buildcache.PinnedGit(source, inputs, pin, nil, "rev-list", "--objects", "--no-object-names", pin)
		if err != nil {
			return err
		}
		pack, err := buildcache.PinnedGit(source, inputs, pin, objects, "pack-objects", "--stdout")
		if err != nil {
			return err
		}
		tree := filepath.Join(directory, "tree")
		// No template directory, so nothing of git's install enters the product, and no reflogs, which hold the time.
		git := func(stdin []byte, arguments ...string) error {
			command := toolchain.GitCommandIn(tree, append([]string{"-c", "core.logAllRefUpdates=false"}, arguments...)...)
			command.Stdin = bytes.NewReader(stdin)
			return described(command)
		}
		if err := os.MkdirAll(tree, 0o755); err != nil {
			return err
		}
		if err := git(nil, "init", "--quiet", "--template=", "."); err != nil {
			return err
		}
		if err := git(pack, "index-pack", "--stdin"); err != nil {
			return err
		}
		// The pack holds the commit and no parent, as a depth-one clone does.
		if err := os.WriteFile(filepath.Join(tree, ".git/shallow"), []byte(pin+"\n"), 0o644); err != nil {
			return err
		}
		if err := git(nil, "checkout", "--quiet", "--detach", pin); err != nil {
			return err
		}
		adapt := toolchain.CommandIn(repository, "python3", "-I", "-S", "-B", "stage3/apply.py", "--adapt-checkout", tree)
		if err := described(adapt); err != nil {
			return err
		}
		// The index records each file's inode and times, which differ on every build; nothing reads it after adaptation.
		return os.Remove(filepath.Join(tree, ".git/index"))
	})
}

// typescriptPin is the commit stage3/source.json pins.
func typescriptPin(t testing.TB, repository string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repository, "stage3/source.json"))
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Commit string `json:"commit"`
	}
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	return source.Commit
}

// described runs command, and on failure says what it ran and what it printed.
func described(command *exec.Cmd) error {
	if command.Err != nil {
		return command.Err
	}
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w\n%s", command.String(), err, output)
	}
	return nil
}

func TestProduct_Stage3AdaptedTree(t *testing.T) {
	t.Parallel()
	tree := filepath.Join(adaptedTree(t), "tree")
	head, err := toolchain.GitCommandIn(tree, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if pin := typescriptPin(t, repository); string(bytes.TrimSpace(head)) != pin {
		t.Fatalf("the adapted tree is at %s, not the pinned %s", head, pin)
	}
}
