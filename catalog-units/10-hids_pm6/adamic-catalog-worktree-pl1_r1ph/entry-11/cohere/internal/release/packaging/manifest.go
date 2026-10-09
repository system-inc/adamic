package release

import (
	"encoding/json"
	"fmt"
)

// PackageScope is the npm scope every cohere package publishes under.
//
// A scope rather than a bare name because the bare name is not ours: `cohere` is published on npm
// by someone else, and so was `verify` before it. Two bare names taken in a row is the evidence that
// a scope is the durable answer rather than a fallback. `@system-inc` matches the repository at
// `github.com/system-inc/cohere`, and it is the name Structure already tells a consumer to install.
const PackageScope = "@system-inc"

// UnscopedDispatcherPackageName is the dispatcher's name inside the scope.
//
// It is also the directory the dispatcher is staged in, so a staged release reads the same way
// `node_modules/@system-inc/` does, and the stem every platform package's name begins with.
const UnscopedDispatcherPackageName = "cohere"

// DispatcherPackageName is the package a consumer actually installs.
const DispatcherPackageName = PackageScope + "/" + UnscopedDispatcherPackageName

// ShortCommandName is what a consumer types on the loop they sit in all day.
//
// One letter, because the whole argument of this tool is that verification stops being a thing you
// do and becomes a property the codebase has. A command run after every edit should cost as little
// to type as it costs to run, and six keystrokes charged on that loop is a real toll.
const ShortCommandName = "c"

// FullCommandName is the same command under the name that reads in a script.
//
// Both names are installed, pointing at the same launcher, which costs nothing: npm's `bin` is a
// map from command name to file, so two entries are two symlinks onto one entry point rather than
// one command wrapping another. There is no extra process and no measurable difference between
// them.
//
// Having both retires the one real objection to a single letter. `v` is short enough to collide
// with something already on a given machine, and a consumer who hits that keeps a working command
// rather than a broken install. It is also the name that belongs in continuous integration, in
// documentation, and in a script somebody reads a year from now, where brevity buys nothing and
// saying what it does buys everything.
const FullCommandName = "cohere"

// launcherRelativePath is where the launcher sits inside the dispatcher package.
//
// Declared once because the manifest's `bin` entries and the build that writes the file have to
// agree exactly. A package whose bin points at a path the build never wrote installs successfully
// and produces a command that does not exist, which is a failure the install itself will not
// report.
//
// Named for the full command rather than the short one: the file is the thing, and `v` is a name
// for it.
const launcherRelativePath = "bin/" + FullCommandName

// RepositoryURL is where the source lives, recorded in every published package. In npm's own form,
// git+https with .git, because npm rewrites any other spelling at publish time and warns about it on
// every package, which buries the warnings that matter.
const RepositoryURL = "git+https://github.com/system-inc/cohere.git"

// PlatformManifest is the package.json for one platform's binary package.
//
// The `os` and `cpu` fields are the entire mechanism. npm and pnpm read them to decide which
// optional dependency to install on this machine, so a package whose fields disagree with its
// contents installs on the wrong platform and produces an exec-format error, or installs nowhere
// and looks like an unsupported platform. They are generated from the same target as the binary,
// rather than written by hand, so the two cannot drift.
func PlatformManifest(target Target, version string) ([]byte, error) {
	manifest := map[string]any{
		"name":        target.PackageName(),
		"version":     version,
		"description": fmt.Sprintf("The cohere binary for %s.", target),
		"license":     LicenseExpression,
		"author":      CopyrightHolder,
		"repository":  map[string]string{"type": "git", "url": RepositoryURL},

		"os":  []string{NpmOperatingSystem(target.GoOperatingSystem)},
		"cpu": []string{NpmArchitecture(target.GoArchitecture)},

		// Only the binary and its licenses ship. Without this npm includes whatever is in the directory,
		// and a stray file in a 15 MB package is the kind of thing nobody notices until it is a secret.
		"files": append([]string{"bin/"}, LegalFileNames...),

		// Deliberately no `bin` field. The platform packages are data, not commands: the dispatcher
		// owns the `cohere` name, and a second package claiming it would race for the same link in
		// `node_modules/.bin` and win or lose depending on install order.
	}
	return marshalManifest(manifest)
}

// DispatcherManifest is the package.json for the package a consumer installs.
//
// Every platform package is an optional dependency. That is what makes an install on an unsupported
// platform succeed with a missing binary rather than fail outright — which sounds worse and is
// better, because the failure then happens at `cohere --version` with a message naming the
// platform, instead of inside a package manager's dependency resolution where it says nothing
// useful. The loud failure is the dispatcher's job and it is a better place to fail from.
func DispatcherManifest(version string) ([]byte, error) {
	optionalDependencies := map[string]string{}
	for _, target := range Targets {
		// Pinned exactly, not with a range. The dispatcher and the binary are one artifact split
		// across packages for distribution, so a resolver free to pick a different patch of the
		// binary could pair a dispatcher with a binary built from other rules.
		optionalDependencies[target.PackageName()] = version
	}

	manifest := map[string]any{
		"name":        DispatcherPackageName,
		"version":     version,
		"description": "Type-check, lint, fix, and format a TypeScript codebase in one process.",
		"license":     LicenseExpression,
		"author":      CopyrightHolder,
		"repository":  map[string]string{"type": "git", "url": RepositoryURL},

		"bin": map[string]string{
			ShortCommandName: launcherRelativePath,
			FullCommandName:  launcherRelativePath,
		},
		// schema/ is there so a settings file's "$schema" can name a path inside the install and be
		// validated offline, in any editor, against the schema of the cohere that reads it. SHA256SUMS
		// is what the launcher checks the platform binary against, and a launcher installed without it
		// refuses to run anything. The licenses ship in every package, this one included.
		"files": append([]string{"bin/", SchemaDirectoryName + "/", ChecksumsFileName}, LegalFileNames...),

		"optionalDependencies": optionalDependencies,
	}
	return marshalManifest(manifest)
}

// marshalManifest renders a manifest the way a formatter would leave it.
//
// Two-space indentation and a trailing newline, because these files are committed in staged
// releases and read in diffs. A manifest that reformats on every generation makes every release
// diff look like a rewrite.
func marshalManifest(manifest map[string]any) ([]byte, error) {
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("rendering the package manifest: %w", err)
	}
	return append(encoded, '\n'), nil
}
