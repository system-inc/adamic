// Package adaptedtree is stage3/lane's adapted tree product: the pinned TypeScript checkout with stage3's adaptations
// applied, and built. Its build reads none of lane's code, so it lives here, a LeanProductOf build of what it's handed
// (build), keyed by its own window, its declared Inputs and the recipe of the binary that settles it, this package's
// own (TestProduct_Stage3AdaptedTree), never lane's: settled from lane's test binary, every edit to a lane file rebuilt
// it, 220 s on every lane change's tree. Relative paths resolve from stage3/lane: lane's tests run there, and this
// package's TestMain changes its working directory to that parent directory, as tsprinter's oracles do.
package adaptedtree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// repositoryFromLane is the checkout, as a path from stage3/lane, where every test asking for the product runs.
const repositoryFromLane = "../.."

// pinFlag begins the flag naming the pinned commit, which keys the product and which the build checks out (pinOf).
const pinFlag = "typescript "

// Product is the directory holding the adapted tree, at tree/ inside it.
//
// The adapted tree is the pinned TypeScript checkout with stage3's adaptations applied: 92,000 files, 970 MB, and
// about 158 s of every phase:stage3 run, which built it from nothing each time (#4gp2ytz). It moves only when the pin
// or an adaptation does, so it is a product of its own, built ahead and fetched. Its build never reads the gate input's
// files: buildcache.PinnedGit asks the input's git for the pinned commit's objects as a pack, as key-time work, and
// the key names the commit in a flag, so the key moves with the pin and the read set holds the commit, the adaptation
// scripts, stage3/api's packages and the tools, never 81,500 walked files. The pack is indexed into a repository of
// the product's own, which checks the pin out detached, so the tree's HEAD is the pinned commit itself, as
// stage3/lane/check.py's API check reads it, and stage3/apply.py --adapt-checkout adapts it in place.
//
// The tree is then built: hereby local and hereby tests make TypeScript's own adapted compiler and its test harness
// into built/ (60 MB, about 27 s on Golem), with stage3/typescript's install, the pinned lockfile's devDependencies,
// linked in as node_modules. Neither is Go or reads adamic's compiler, so both move only when the tree does, and one
// product holds both (#4gp2ytz, Tests' ruling). hereby runs from the install's real path, so node resolves every
// module there and the trace values them as the install; the link goes before the product settles.
//
// build is handed, by machine path, what it reads outside the tree's files: the TypeScript source checkout
// (ADAMIC_TYPESCRIPT_SOURCE, read only through buildcache.PinnedGit), stage3/typescript's install (hereby runs from
// it, linked into the tree, then removed) and stage3/api's install (apply.py's pinned_packages reads its
// .package-lock.json and every byte, against node_modules.sha256).
func Product(test testing.TB) string {
	test.Helper()
	repository, absolutePathError := filepath.Abs(repositoryFromLane)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		test.Fatal(
			"ADAMIC_TYPESCRIPT_SOURCE names no checkout of the pinned TypeScript, the gate input the adapted tree is made from",
		)
	}
	pin := typescriptPin(test, repository)
	inputs := buildcache.Inputs{
		Name: "stage3-adapted-tree",
		Files: []string{"stage3/source.json", "stage3/apply.py", "stage3/caches.py", "stage3/adapt",
			"stage3/api/package.json", "stage3/api/package-lock.json", "stage3/api/node_modules.sha256",
			"stage3/typescript/package.json", "stage3/typescript/package-lock.json"},
		Flags: []string{pinFlag + pin},
		Toolchain: []string{
			buildcache.Tool("git", "--version"),
			buildcache.Tool("node", "--version"),
			buildcache.Tool("python3", "--version"),
		},
	}
	paths := []string{
		source,
		filepath.Join(repository, "stage3/typescript/node_modules"),
		filepath.Join(repository, "stage3/api/node_modules"),
	}
	return buildcache.LeanProductOf(test, buildcache.Handed{Inputs: inputs, Paths: paths}, build)
}

// build makes the adapted tree in directory from what it's handed: the TypeScript source checkout, stage3/typescript's
// install and stage3/api's install, in that order (Product). It's a top-level function of what it's handed alone
// (buildcache.LeanProductOf): no test, no package state (TestTheAdaptedTreeBuildReachesOnlyItsDeclaredFile). It checks
// out the commit handed.Inputs' flag names, so the tree is always the one its key names, and finds the repository as
// every caller's working directory names it (repositoryFromLane), where apply.py and the adaptations are read as files
// of the tree.
func build(handed buildcache.Handed, directory string) error {
	if len(handed.Paths) != 3 {
		return fmt.Errorf(
			"adaptedtree: handed %q, want the TypeScript source, stage3/typescript's install and stage3/api's",
			handed.Paths,
		)
	}
	source, install := handed.Paths[0], handed.Paths[1]
	repository, absolutePathError := filepath.Abs(repositoryFromLane)
	if absolutePathError != nil {
		return absolutePathError
	}
	pin, pinOfError := pinOf(handed.Inputs)
	if pinOfError != nil {
		return pinOfError
	}
	objects, pinnedGitError := buildcache.PinnedGit(
		source,
		handed.Inputs,
		pin,
		nil,
		"rev-list",
		"--objects",
		"--no-object-names",
		pin,
	)
	if pinnedGitError != nil {
		return pinnedGitError
	}
	// One thread, so the pack's bytes are a function of its objects: pack-objects' threaded delta search lays the same
	// objects out differently from build to build, which an audit rebuild reads as Nondeterministic (Node's ruling on
	// #qwfw9pv). Its objects are the pin's alone, since the source is the gate input's depth-one checkout.
	pack, pinnedGitError := buildcache.PinnedGit(
		source,
		handed.Inputs,
		pin,
		objects,
		"-c",
		"pack.threads=1",
		"pack-objects",
		"--stdout",
	)
	if pinnedGitError != nil {
		return pinnedGitError
	}
	tree := filepath.Join(directory, "tree")
	// No template directory, so nothing of git's install enters the product, and no reflogs, which hold the time.
	git := func(stdin []byte, arguments ...string) error {
		command := toolchain.GitCommandIn(tree, append([]string{"-c", "core.logAllRefUpdates=false"}, arguments...)...)
		command.Stdin = bytes.NewReader(stdin)
		return described(command)
	}
	// The build's home is an empty directory of its own, removed before the product settles (isolated says why).
	home := filepath.Join(directory, "home")
	for _, made := range []string{tree, home} {
		if createDirectoryError := os.MkdirAll(made, 0o755); createDirectoryError != nil {
			return createDirectoryError
		}
	}
	if gitError := git(nil, "init", "--quiet", "--template=", "."); gitError != nil {
		return gitError
	}
	if gitError := git(pack, "index-pack", "--stdin"); gitError != nil {
		return gitError
	}
	// The pack holds the commit and no parent, as a depth-one clone does.
	if writeError := os.WriteFile(filepath.Join(tree, ".git/shallow"), []byte(pin+"\n"), 0o644); writeError != nil {
		return writeError
	}
	if gitError := git(nil, "checkout", "--quiet", "--detach", pin); gitError != nil {
		return gitError
	}
	// -E and -S keep python off the box's PYTHON* settings and site directories, whose listing would enter the read
	// set; not -I, which would also drop the script's directory, where apply.py finds caches.py.
	adapt := isolated(
		home,
		toolchain.CommandIn(repository, "python3", "-E", "-S", "-B", "stage3/apply.py", "--adapt-checkout", tree),
	)
	// apply.py runs git itself (add, write-tree, update-ref), which would read the home's ~/.gitconfig: Workshop
	// has one and Golem none, so tree 276cc52a refused the product. GitEnv's GIT_ settings keep every git
	// configuration but the repository's out of reach, as toolchain's own git runs.
	for _, setting := range toolchain.GitEnv() {
		if strings.HasPrefix(setting, "GIT_") {
			adapt.Env = append(adapt.Env, setting)
		}
	}
	if describedError := described(adapt); describedError != nil {
		return describedError
	}
	// The index records each file's inode and times, which differ on every build; nothing reads it after adaptation.
	if removeError := os.Remove(filepath.Join(tree, ".git/index")); removeError != nil {
		return removeError
	}
	if _, statError := os.Stat(install); statError != nil {
		return fmt.Errorf(
			"stage3/typescript's install, the adapted tree's devDependencies, isn't in the checkout: %w",
			statError,
		)
	}
	link := filepath.Join(tree, "node_modules")
	if symlinkError := os.Symlink(install, link); symlinkError != nil {
		return symlinkError
	}
	for _, task := range []string{"local", "tests"} {
		if describedError := described(
			isolated(home, toolchain.CommandIn(tree, "node", filepath.Join(install, "hereby/bin/hereby.js"), task)),
		); describedError != nil {
			return describedError
		}
	}
	if removeError := os.Remove(link); removeError != nil {
		return removeError
	}
	// The typecheck's tsconfig.tsbuildinfo files record each library's path relative to the build, up through the
	// checkout to the install's real path, so two checkouts would build two products under one key (Golem, Oct 16:
	// 13 files). They serve only an incremental rebuild, and nothing runs from them.
	walkError := filepath.WalkDir(filepath.Join(
		tree,
		"built",
	), func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil || entry.IsDir() || filepath.Ext(path) != ".tsbuildinfo" {
			return pathError
		}
		return os.Remove(path)
	})
	if walkError != nil {
		return walkError
	}
	if keepEmptyDirectoriesError := keepEmptyDirectories(
		filepath.Join(tree, ".git"),
	); keepEmptyDirectoriesError != nil {
		return keepEmptyDirectoriesError
	}
	return os.RemoveAll(home)
}

// keepEmptyDirectories writes an empty .keep into each directory under root that holds nothing (#rhj78rk, Node's
// ruling). The house's lay keeps no empty directory, and git finds a repository only with refs/ beside objects/ and
// HEAD: a laid copy of this tree's .git, whose refs/heads and refs/tags are empty since the pin is checked out
// detached, would read "not a git repository" to stage3/lane/check.py's API check. git's loose-ref walk skips every
// name beginning with a dot, as no ref name may begin with one, so a .keep is read as no ref and the tree reads as
// before.
func keepEmptyDirectories(root string) error {
	empty := []string{}
	walkError := filepath.WalkDir(root, func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil || !entry.IsDir() {
			return pathError
		}
		entries, readDirectoryError := os.ReadDir(path)
		if readDirectoryError == nil && len(entries) == 0 {
			empty = append(empty, path)
		}
		return readDirectoryError
	})
	if walkError != nil {
		return walkError
	}
	for _, directory := range empty {
		if writeError := os.WriteFile(filepath.Join(directory, ".keep"), nil, 0o644); writeError != nil {
			return writeError
		}
	}
	return nil
}

// pinOf is the commit inputs' pin flag names, the one the key holds.
func pinOf(inputs buildcache.Inputs) (string, error) {
	for _, flag := range inputs.Flags {
		if pin, ok := strings.CutPrefix(flag, pinFlag); ok {
			return pin, nil
		}
	}
	return "", fmt.Errorf("adaptedtree: %s's flags %q name no TypeScript commit", inputs.Name, inputs.Flags)
}

// isolated keeps the adapted tree's build off the box: its home is home, an empty directory inside the product's own
// build, and node's compile cache is off. hereby's cli.js and TypeScript's tsc.js call module.enableCompileCache(),
// which writes under the temporary directory, and Workshop's is shared (/tmp/adamic-gate), so tree 77eff2ce refused the
// product for /tmp/adamic-gate/node-compile-cache. With no `HOME` at all, node's os.homedir() asks the user database,
// and /etc/passwd differs between boxes (chg_cjg1c5e4, red on Cloud), so the home is set, never dropped: reads inside
// the product's own directory are its build's, never its inputs.
func isolated(home string, command *exec.Cmd) *exec.Cmd {
	command.Env = slices.DeleteFunc(command.Env, func(setting string) bool {
		return strings.HasPrefix(setting, "HOME=") || strings.HasPrefix(setting, "XDG_CONFIG_HOME=")
	})
	command.Env = append(command.Env, "HOME="+home, "XDG_CONFIG_HOME="+home, "NODE_DISABLE_COMPILE_CACHE=1")
	return command
}

// typescriptPin is the commit stage3/source.json pins.
func typescriptPin(test testing.TB, repository string) string {
	test.Helper()
	data, readError := os.ReadFile(filepath.Join(repository, "stage3/source.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var source struct {
		Commit string `json:"commit"`
	}
	if decodeError := json.Unmarshal(data, &source); decodeError != nil {
		test.Fatal(decodeError)
	}
	return source.Commit
}

// described runs command, and on failure says what it ran and what it printed.
func described(command *exec.Cmd) error {
	if command.Err != nil {
		return command.Err
	}
	if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
		return fmt.Errorf("%s: %w\n%s", command.String(), combinedOutputError, output)
	}
	return nil
}
