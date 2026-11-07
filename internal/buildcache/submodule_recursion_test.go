package buildcache

import (
	"os"
	"path/filepath"
	"testing"
)

// superproject is a repository holding one submodule, module, whose tracked inner.c is edited, so git would call the
// submodule dirty if asked.
func superproject(test *testing.T) (parent, module string) {
	test.Helper()
	source := test.TempDir()
	write(test, source, "inner.c", "int inner;\n")
	gitIn(test, source, "init", "-q")
	gitIn(test, source, "add", ".")
	gitIn(test, source, "commit", "-q", "-m", "module")
	parent, err := filepath.EvalSymlinks(test.TempDir())
	if err != nil {
		test.Fatal(err)
	}
	write(test, parent, "outer.c", "int outer;\n")
	gitIn(test, parent, "init", "-q")
	gitIn(test, parent, "add", ".")
	gitIn(test, parent, "commit", "-q", "-m", "parent")
	gitIn(test, parent, "-c", "protocol.file.allow=always", "submodule", "add", "-q", source, "module")
	gitIn(test, parent, "commit", "-q", "-m", "module")
	module = filepath.Join(parent, "module")
	write(test, module, "inner.c", "int edited;\n")
	return parent, module
}

// A repository's reading asks none of its submodules whether it is dirty, since git's answer recurses into each and
// sweeps its index again (#npd568j); the submodule's own reading still names its edited file.
// Mutant: repositoryAt's diff-files without --ignore-submodules, so the parent's dirty files name module.
// Not parallel: repositories is the process's.
func TestAParentsReadingNeverAsksItsSubmodules(test *testing.T) {
	parent, module := superproject(test)
	forgetTracked()
	test.Cleanup(forgetTracked)
	if files := repositoryAt(parent); files == nil || files.all || files.dirty["module"] {
		test.Fatalf("the parent's reading asked its submodule whether it is dirty: %+v", files)
	}
	if files := repositoryAt(module); files == nil || !files.dirty["inner.c"] {
		test.Fatalf("the submodule's own reading doesn't mark its edited file dirty: %+v", files)
	}
}

// A file inside a submodule is valued by the submodule's own listing, never by the parent's, which holds the
// submodule only as a gitlink and now never asks it whether it is dirty: an edited file is hashed, and its value moves
// with its bytes.
// Mutant: cleanObject taking a gitlink's mode (160000) as a file's. It stays green, and is recorded as such: the nearest
// repository above module/inner.c is the submodule's, where inner.c is 100644, and the parent's gitlink entry is
// named module, a directory, which no key values as a file. Dropping the mode check altogether stays green too.
// Not parallel: repositories is the process's.
func TestAFileInASubmoduleIsValuedByTheSubmodulesOwnListing(test *testing.T) {
	parent, module := superproject(test)
	forgetTracked()
	test.Cleanup(forgetTracked)
	name := filepath.Join(module, "inner.c")
	value := func(content string) string {
		test.Helper()
		write(test, module, "inner.c", content)
		forgetTracked()
		info, err := os.Lstat(name)
		if err != nil {
			test.Fatal(err)
		}
		if _, _, clean := cleanObject(parent, name, info); clean {
			test.Fatalf("an edited file inside the submodule was taken as clean")
		}
		got, err := gitValue(parent, name, info)
		if err != nil {
			test.Fatal(err)
		}
		if want := "false " + blobID([]byte(content), 40); got != want {
			test.Fatalf("inner.c holding %q valued %q, want %q", content, got, want)
		}
		return got
	}
	if value("int edited;\n") == value("int edited again;\n") {
		test.Fatal("inner.c's value didn't move with its bytes")
	}
}

// TreeState sees a change in a submodule however deep, though no repository's status asks its submodules: each one
// with a .git gets its own status and ls-files. An edit is seen by ls-files -m, a staged change (the working copy
// matching the index) only by the submodule's own status.
// Mutant: status run only for the root: "staged in a submodule" and "staged two submodules deep" go red.
// Mutant: status run for every repository but module/deep: "staged two submodules deep" goes red.
// Not parallel: it changes the working directory (inRepository) and repositories.
func TestTreeStateSeesASubmodulesChangeHoweverDeep(test *testing.T) {
	for name, change := range map[string]func(parent string){
		"edited in a submodule": func(parent string) { write(test, parent, "module/inner.c", "int edited;\n") },
		"edited two submodules deep": func(
			parent string,
		) {
			write(test, parent, "module/deep/deeper.c", "int edited;\n")
		},
		"staged in a submodule": func(parent string) {
			write(test, parent, "module/inner.c", "int staged;\n")
			gitIn(test, filepath.Join(parent, "module"), "add", "inner.c")
		},
		"staged two submodules deep": func(parent string) {
			write(test, parent, "module/deep/deeper.c", "int staged;\n")
			gitIn(test, filepath.Join(parent, "module/deep"), "add", "deeper.c")
		},
	} {
		test.Run(name, func(test *testing.T) {
			deepSource := test.TempDir()
			write(test, deepSource, "deeper.c", "int deeper;\n")
			gitIn(test, deepSource, "init", "-q")
			gitIn(test, deepSource, "add", ".")
			gitIn(test, deepSource, "commit", "-q", "-m", "deep")
			source := test.TempDir()
			write(test, source, "inner.c", "int inner;\n")
			gitIn(test, source, "init", "-q")
			gitIn(test, source, "add", ".")
			gitIn(test, source, "-c", "protocol.file.allow=always", "submodule", "add", "-q", deepSource, "deep")
			gitIn(test, source, "commit", "-q", "-m", "module")
			parent, err := filepath.EvalSymlinks(test.TempDir())
			if err != nil {
				test.Fatal(err)
			}
			write(test, parent, "go.mod", "module github.com/system-inc/adamic\n\ngo 1.27\n")
			gitIn(test, parent, "init", "-q")
			gitIn(test, parent, "-c", "protocol.file.allow=always", "submodule", "add", "-q", source, "module")
			gitIn(
				test,
				parent,
				"-c",
				"protocol.file.allow=always",
				"submodule",
				"update",
				"-q",
				"--init",
				"--recursive",
			)
			gitIn(test, parent, "add", ".")
			gitIn(test, parent, "commit", "-q", "-m", "parent")
			inRepository(test, parent)
			forgetTracked()
			test.Cleanup(forgetTracked)
			before, err := TreeState()
			if err != nil {
				test.Fatal(err)
			}
			change(parent)
			after, err := TreeState()
			if err != nil {
				test.Fatal(err)
			}
			if after == before {
				test.Fatalf("the change went unseen: the state stayed\n%s", before)
			}
		})
	}
}
