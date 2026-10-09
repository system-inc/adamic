// Command build_dispatcher_package writes one npm package to a directory, and compiles nothing.
//
// A release is seven npm packages: six holding a cross-compiled binary each, and the one a consumer
// actually installs, which holds no binary at all. That last one is the dispatcher package — a Node
// entry point that reads the platform, resolves the matching binary package, and execs it. The word
// `build` here is npm's sense and not Go's: this writes a package.json and a script, and there is no
// compilation anywhere in it.
//
// It also does not publish. `command/cohere-release` stages a full release into a directory for a
// person to inspect before anything reaches the registry, and this writes its dispatcher half the
// same way.
//
// The full release takes about two minutes and correctly refuses to stage a partial one when any
// target fails to build. So exercising the INSTALL path — whether `pnpm add` links the right entry
// point, whether `require.resolve` finds the binary package, whether the `os` and `cpu` fields match
// what npm expects — costs two minutes of cross-compilation to test a part with no Go in it.
//
// This writes that one package in a second. It is also what lets the install path be tested while
// `command/cohere` is mid-edit by another author and would not build at all.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/system-inc/cohere/internal/release/packaging"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: npm/build_dispatcher_package <output-directory>")
		os.Exit(1)
	}

	directory := filepath.Join(os.Args[1], release.UnscopedDispatcherPackageName)
	if err := os.MkdirAll(filepath.Join(directory, "bin"), 0o755); err != nil {
		panic(err)
	}

	manifest, err := release.DispatcherManifest("1.0.0")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "package.json"), manifest, 0o644); err != nil {
		panic(err)
	}

	launcherPath := filepath.Join(directory, "bin", "cohere")
	if err := os.WriteFile(launcherPath, []byte(release.DispatcherLauncher()), 0o755); err != nil {
		panic(err)
	}
	if err := os.Chmod(launcherPath, 0o755); err != nil {
		panic(err)
	}

	// The launcher runs only a binary its checksums list, so they are taken from the platform packages
	// already staged beside this one. With none there it would install and run nothing, so it stops here.
	var platformDirectories []string
	for _, target := range release.Targets {
		if _, err := os.Stat(filepath.Join(os.Args[1], target.DirectoryName(), "bin")); err == nil {
			platformDirectories = append(platformDirectories, target.DirectoryName())
		}
	}
	if len(platformDirectories) == 0 {
		fmt.Fprintln(os.Stderr, "no platform package is staged in", os.Args[1], "so the launcher would have no binary it could check; stage one there first")
		os.Exit(1)
	}
	checksums, err := release.Checksums(os.Args[1], platformDirectories)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(directory, release.ChecksumsFileName), checksums, 0o644); err != nil {
		panic(err)
	}

	fmt.Println("staged", directory)
}
