package buildcache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// A file only its group and others may execute (0654) is not executable as git records it: git's index takes the
// owner's bit alone, so every value buildcache gives the file's executable bit says false, as its index entry does
// (#kp7mx3z). Checked against git itself, as cohere's git_tree checks its rule. The hashed side of gitValue is the one
// that read any bit before, so a clean file and the same file edited valued its mode two ways.
// Mutant: executableAsGit reading info.Mode()&0o111, so gitValue says true where git says 100644.
// Not parallel: repositories is the process's.
func TestAGroupExecutableFileIsValuedAsGitRecordsIt(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write(t, root, "group.sh", "echo group\n")
	write(t, root, "owner.sh", "echo owner\n")
	for name, mode := range map[string]os.FileMode{"group.sh": 0o654, "owner.sh": 0o744} {
		if err := os.Chmod(filepath.Join(root, name), mode); err != nil {
			t.Fatal(err)
		}
	}
	gitIn(t, root, "init", "-q")
	gitIn(t, root, "add", ".")
	gitIn(t, root, "commit", "-q", "-m", "modes")
	listed := gitOutput(t, root, "ls-files", "-s")
	// Edited after the commit, keeping their modes, so git calls them dirty and gitValue hashes them, the side that
	// read any bit before, whatever the clock.
	write(t, root, "group.sh", "echo group edited\n")
	write(t, root, "owner.sh", "echo owner edited\n")
	forgetTracked()
	t.Cleanup(forgetTracked)
	for name, want := range map[string]bool{"group.sh": false, "owner.sh": true} {
		mode := "100644"
		if want {
			mode = "100755"
		}
		if !lineHas(listed, mode, name) {
			t.Fatalf("git doesn't record %s as %s, so this test's premise is wrong:\n%s", name, mode, listed)
		}
		path := filepath.Join(root, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, clean := cleanObject(root, path, info); clean {
			t.Fatalf("%s, edited, was taken as clean, so the hashed side isn't what's tested", name)
		}
		got, err := gitValue(root, path, info)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(got, map[bool]string{true: "true ", false: "false "}[want]) {
			t.Errorf("gitValue of %s (git records %s) is %q", name, mode, got)
		}
		if got, want := existence(path), map[bool]string{true: "file true", false: "file false"}[want]; got != want {
			t.Errorf("existence of %s (git records %s) is %q, want %q", name, mode, got, want)
		}
	}
	// A product's digest reads the same rule: 0654 and 0644 digest alike, and 0744 apart.
	digest := func(mode os.FileMode) string {
		t.Helper()
		directory := filepath.Join(t.TempDir(), "product")
		write(t, directory, "run.sh", "echo run\n")
		if err := os.Chmod(filepath.Join(directory, "run.sh"), mode); err != nil {
			t.Fatal(err)
		}
		value, err := productDigest(directory)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	if plain, group, owner := digest(0o644), digest(0o654), digest(0o744); plain != group || plain == owner {
		t.Errorf("product digests: 0644 %s, 0654 %s, 0744 %s; want 0654 equal to 0644 and 0744 apart", plain, group, owner)
	}
}

func gitOutput(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := toolchain.GitCommand(append([]string{"-C", directory}, arguments...)...)
	output, err := command.Output()
	if err != nil {
		t.Fatalf("git %v: %v", arguments, err)
	}
	return string(output)
}

func lineHas(listing, mode, name string) bool {
	for _, line := range strings.Split(listing, "\n") {
		if strings.HasPrefix(line, mode+" ") && strings.HasSuffix(line, "\t"+name) {
			return true
		}
	}
	return false
}
