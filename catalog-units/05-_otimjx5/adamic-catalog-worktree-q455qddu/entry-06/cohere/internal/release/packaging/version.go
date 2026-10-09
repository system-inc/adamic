// Package release holds what a shipped cohere knows about itself: which build it is, and where its
// binary comes from on a machine that has no source and no Go toolchain.
//
// The two halves answer the two questions a bug report has to answer. `Provenance` says what was
// running, so a report names a version, a toolchain, and a pinned compiler commit rather than
// "latest". `Resolve` says which file ran, and refuses to guess — a resolver that falls through to
// something plausible is how the gate this tool replaces printed green over zero files for days.
//
// `Build` here stages the published packages, which is the other end of the same path: it is what
// puts a binary where `Resolve` will find it, stamped with what `Provenance` will report.
package release

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// These are stamped at link time by the release pipeline, with `-ldflags=-X`. Their defaults are
// what a local `go build` produces, and they say so plainly rather than reading like a real
// release: a binary that cannot name its own provenance must not be able to look like one that can.
var (
	// version is the published npm version, like "0.3.1".
	version = "dev"

	// compilerCommit is the commit of the vendored compiler this binary statically links. It is
	// most of the code in the binary and it moves independently of this repository, so a bug that
	// reproduces on one pin and not another is only diagnosable if the binary carries it.
	compilerCommit = "unknown"

	// compilerUpstream is the repository that commit belongs to.
	//
	// It is stamped rather than written down because the upstream has already moved once: the
	// vendored compiler was `microsoft/typescript-go` until that repository was archived, and is
	// now `microsoft/TypeScript`. A hardcoded label survives a migration like that while silently
	// becoming false, which is the worst outcome available here — a bug report would name a commit
	// against a repository it does not exist in, and the reader would have no way to tell.
	compilerUpstream = "unknown"

	// goToolchain is the toolchain that compiled this binary, like "go1.27.0". The compiler is an
	// input to the behavior, not just to the build: it is recorded for the same reason the rebuild
	// cache hashes it.
	goToolchain = "unknown"

	// selfCommit is the commit of this repository the binary was built from, stamped by the launcher
	// when it builds from a snapshot of that commit.
	//
	// A snapshot has no `.git`, so Go writes no `vcs.revision` for it, and building it with VCS
	// stamping on would make Go walk up to the enclosing checkout and stamp that tree's state instead,
	// dirty files and all. So the launcher turns stamping off and states the commit itself. It can,
	// because it extracted the snapshot from that commit and nothing else.
	selfCommit = ""

	// formatterIdentity names everything that can change what the formatter prints, stamped by the
	// launcher when it builds (see dispatch.FormatterIdentity). Empty for a build nobody stamped.
	formatterIdentity = ""
)

// Provenance is everything a shipped binary knows about where it came from.
type Provenance struct {
	// Version is the published version, or "dev" for a local build.
	Version string

	// SelfCommit is the commit of THIS repository the binary was built from, or "" when the build
	// carried no version-control stamp.
	//
	// Separate from CompilerCommit, which names the vendored compiler and moves independently. A
	// local build reports its version as "dev", which says nothing about which rules it holds, and
	// every measurement taken with it is only reproducible if the binary can name its own vintage.
	// Two numbers taken from binaries at different commits look exactly like a contradiction, and
	// the cheapest way to tell those apart is for each number to arrive with its commit attached.
	SelfCommit string

	// CompilerCommit is the pinned commit of the vendored compiler.
	CompilerCommit string

	// CompilerUpstream is the repository CompilerCommit belongs to.
	//
	// A commit hash means nothing without the repository it lives in, and this one has moved: a
	// pin recorded before the `microsoft/TypeScript` migration and one recorded after are both
	// forty hex characters and resolve in different places.
	CompilerUpstream string

	// GoToolchain is the Go version that compiled this binary.
	GoToolchain string

	// SourceTreeModified reports whether the tree this binary was built from had uncommitted
	// changes.
	//
	// Read from Go's own `vcs.modified` stamp rather than computed, because the linker observes the
	// tree at build time and nothing later can. It answers the question a version number cannot: a
	// version identifies what was released, and this says whether that release is reproducible from
	// a fresh clone at all.
	//
	// It matters here more than in most projects because the source lives in a worktree several
	// members edit at once. A release staged while someone is mid-flight embeds their uncommitted
	// work, ships, installs, and reports a version that no commit produces. Measured: eight modified
	// files in the tree at the time this was written.
	SourceTreeModified bool

	// FormatterIdentity is a hash of the formatter's inputs: its printers and everything they import,
	// the pass loop that decides what formatted means, and the toolchain. The format record is keyed
	// by it rather than by the binary, so a cohere commit that leaves the formatter alone leaves the
	// record valid. Empty when the build was not stamped, and then the record keys by the binary.
	FormatterIdentity string

	// Platform is the operating system and architecture this binary was built for, as "os/arch".
	//
	// It is read from the runtime rather than stamped, because the runtime cannot be wrong about it
	// and a stamp can. A binary that reports the platform it was supposed to be built for, rather
	// than the one it was, would turn the most common packaging mistake into an invisible one.
	Platform string
}

// Current returns this binary's provenance.
func Current() Provenance {
	return Provenance{
		Version:            version,
		SelfCommit:         resolveSelfCommit(),
		CompilerCommit:     resolveCompilerCommit(),
		CompilerUpstream:   compilerUpstream,
		GoToolchain:        resolveGoToolchain(),
		SourceTreeModified: resolveSourceTreeModified(),
		FormatterIdentity:  formatterIdentity,
		Platform:           runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// IsDevelopment reports whether this binary came from a local build rather than a release.
//
// The distinction matters for a bug report: a development binary's rules are whatever was on disk
// at the time, so its version number identifies nothing and should not be trusted as if it did.
func (provenance Provenance) IsDevelopment() bool {
	return provenance.Version == "dev"
}

// String renders the provenance as one line per fact.
//
// Every fact prints on its own line, including the ones that are unknown, because a missing line
// and an unknown value read identically in a pasted bug report while meaning very different things:
// one is an old binary that predates the stamping, the other is a build that lost its stamp.
func (provenance Provenance) String() string {
	lines := []string{
		"cohere " + provenance.Version,
		"  platform:       " + provenance.Platform,
		"  go:             " + provenance.GoToolchain,
		"  compiler:       " + provenance.describeCompiler(),
	}
	// Printed whenever the build carried a stamp, including a release, because a published version
	// still does not say which commit produced it and a bug report quoting both is easier to chase.
	if provenance.SelfCommit != "" {
		lines = append(lines, "  commit:         "+ShortCommit(provenance.SelfCommit))
	}
	// Printed only when true, because "built from a clean tree" is the ordinary case and a line
	// asserting it on every release would be noise that hides the one time it matters.
	if provenance.SourceTreeModified {
		lines = append(lines, "  source:         built from a tree with uncommitted changes, so no commit reproduces this binary")
	}
	if provenance.IsDevelopment() {
		// A local build of a named commit with nothing uncommitted in it is reproducible from that
		// commit, which is what the launcher builds every gate from. Saying "whatever was on disk"
		// there would be a disclaimer that is no longer true, and a reader who learns the disclaimer
		// is decorative stops reading the one that is not: a `--dev` build, or any build from a
		// dirty tree, still carries it.
		if provenance.SelfCommit != "" && !provenance.SourceTreeModified {
			lines = append(lines, "  note:           a local build of the commit above, with nothing uncommitted in it, so that commit reproduces it")
		} else {
			lines = append(lines, "  note:           a local build, so the rules are whatever was on disk when it was compiled")
		}
	}
	return strings.Join(lines, "\n")
}

// describeCompiler renders the vendored compiler as a repository and a commit in it.
//
// The repository is printed alongside the hash rather than assumed, because the vendored compiler
// has already moved once — `microsoft/typescript-go` was archived and the pin is now against
// `microsoft/TypeScript`. Two pins from either side of that migration are both forty hex characters
// and resolve in different places, so a hash alone is not enough to look one up.
func (provenance Provenance) describeCompiler() string {
	if provenance.CompilerUpstream == "" || provenance.CompilerUpstream == "unknown" {
		return provenance.CompilerCommit
	}
	return provenance.CompilerUpstream + "@" + provenance.CompilerCommit
}

// resolveGoToolchain prefers the toolchain recorded in the binary over the stamped value.
//
// `runtime.Version()` is written by the linker that actually built this file, so it cannot disagree
// with reality. The stamp is a fallback for the same reason the platform is not stamped at all: a
// value a pipeline computes can be wrong, and a value the toolchain writes about itself cannot.
func resolveGoToolchain() string {
	if toolchain := runtime.Version(); toolchain != "" {
		return toolchain
	}
	return goToolchain
}

// resolveCompilerCommit prefers the stamped commit, falling back to the module's own VCS data.
//
// The stamp is authoritative because the vendored compiler is a git submodule rather than a Go
// module dependency, so the build info records this repository's revision and never the submodule's.
// The fallback exists so that a `go build` with no stamping still says something true about which
// checkout it came from, instead of the flat "unknown" that reads like a packaging failure.
func resolveCompilerCommit() string {
	if compilerCommit != "unknown" && compilerCommit != "" {
		return compilerCommit
	}

	information, available := debug.ReadBuildInfo()
	if !available {
		return "unknown"
	}
	for _, setting := range information.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			// Named for what it actually is. Reporting this repository's commit under the
			// compiler's label would be a true fact wearing a wrong name, which is worse than
			// the honest "unknown" because a reader would act on it.
			return fmt.Sprintf("unknown (built from cohere %s)", ShortCommit(setting.Value))
		}
	}
	return "unknown"
}

// resolveSelfCommit reads this repository's commit from the stamp Go's linker writes.
//
// Go records `vcs.revision` on any build inside a version-controlled tree, so this needs no
// `-ldflags` and works for a local `go build` as well as a release. A launcher build from a committed
// snapshot stamps selfCommit instead, and that wins, because it was stated deliberately by the only
// process that knows which commit the snapshot holds. Empty when neither is present, which is the
// honest answer for a build made outside a repository or with `-buildvcs=false` and no stamp.
func resolveSelfCommit() string {
	if selfCommit != "" {
		return selfCommit
	}
	information, available := debug.ReadBuildInfo()
	if !available {
		return ""
	}
	for _, setting := range information.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return ""
}

// resolveSourceTreeModified reports whether the build tree carried uncommitted changes.
//
// Go's linker writes this stamp itself, so it observes the tree at the moment of the build and
// cannot be wrong about it the way a value computed afterwards could. A binary with no build info
// at all reports false rather than true: absent evidence is not evidence of a dirty tree, and
// claiming otherwise would put a warning on every binary that predates this field.
func resolveSourceTreeModified() bool {
	information, available := debug.ReadBuildInfo()
	if !available {
		return false
	}
	for _, setting := range information.Settings {
		if setting.Key == "vcs.modified" {
			return setting.Value == "true"
		}
	}
	return false
}

// ShortCommit trims a full hash to the length people actually read, the one home for that length.
//
// Exported because the launcher checks a freshly built binary by finding its own commit, shortened
// this way, on the `commit:` line this package prints. Two copies of the length would have to agree
// by coincidence, and the day they did not, every committed build would be refused.
func ShortCommit(commit string) string {
	const shortLength = 12
	if len(commit) <= shortLength {
		return commit
	}
	return commit[:shortLength]
}
