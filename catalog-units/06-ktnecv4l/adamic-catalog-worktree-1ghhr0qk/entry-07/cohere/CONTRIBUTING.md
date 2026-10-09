# Contributing to cohere

How cohere is built, and why it is built that way. For installing and running it, see
[README.md](README.md).

## Why

Writing code is becoming free; knowing it is correct is not. A gate that is slow gets skipped, and
a skipped gate enforces nothing. The goal is unlimited rules checked in about a quarter second, so
that encoding a judgment stops needing justification.

Measured on a 3,416-file, 24.7 MB TypeScript codebase, the existing toolchain costs 9.3 seconds —
517x slower than running a regex over the same bytes. Almost none of that is compute. It is
boundaries: rules crossing into a JavaScript isolate, a Tailwind plugin making a blocking
cross-thread round trip per class literal, and a type checker rebuilding a 9,530-file program from
scratch on every run, once per shard.

`cohere` removes the boundaries by standing in the type checker's own language.

## The name

`cohere` is a transitive verb. You cohere something: it takes an object and changes it, which is
what this does. It lints, autofixes and formats, and `ReportNodeWithFixes` is the heart of the rule
system, so the most consequential surface here is the one that rewrites your source.

From _cohaerēre_, to stick together. The physics sense is the useful one: coherent light is not
light that is similar, it is light that is phase-locked, and nothing is added to make it so. The
photons stop cancelling each other. That is the problem this solves for a codebase written by many
hands at once.

## Coherent is not the same as correct, and `verify` is the program that will say so

Logic has the pair:

- **coherent** means the parts hang together, with no internal contradiction
- **correspondent** means the whole thing matches the world outside it

A perfectly coherent story can be entirely false. Coherence is a relation among the parts and says
nothing about whether they touch reality, which is why a green run over a probe that never fired
feels exactly like a green run over a clean tree.

Those are two jobs rather than two features of one program. `cohere` answers *is this well-formed
and in phase*, from the graph alone, which is what lets it be fast enough to leave on. A separate
`verify` will answer *did it actually do the thing*, which cannot be answered from inside the code
and needs running it and reading what came out. The oracles belong there: the fixture pairs, the
mutation sweeps, the known-dirty controls, the differential against the gate this replaces. Today
they live here, in `internal/lint/rules/tools/` and beside the rules they hold.

## Built on

- [microsoft/typescript-go](https://github.com/microsoft/typescript-go) — Apache 2.0. The TypeScript
  compiler in Go. Vendored as a submodule and reached through generated shim modules.
- [typescript-eslint/tsgolint](https://github.com/typescript-eslint/tsgolint) — MIT. The shim
  generator and the type-aware rule catalog. We do not fork it: its architecture is one-shot and
  stateless by design, which is precisely what we are replacing.
- [web-infra-dev/rslint](https://github.com/web-infra-dev/rslint) — MIT, copyright Bytedance Inc and
  typescript-eslint contributors. Its `internal/utils/cfg` is vendored at
  `internal/lint/ecmascript/control_flow_graph`: a basic-block control-flow graph with dominators, reachability, and
  correct `try`/`catch`/`finally` edges, built on the same typescript-go AST we use. Vendored rather
  than depended on, because it is a package rather than a library and we change it when we change a
  shared decision. The header on `internal/lint/ecmascript/control_flow_graph/cfg.go` names the commit and what was
  changed.

## Development

```sh
git submodule update --init TypeScript
go build ./...
go run ./command/cohere
```

`cohere` cannot cohere itself — it is a Go program and its phases check TypeScript. This repo is
gated by Go's own toolchain: `gofmt -l .`, `go vet ./...`, `go test ./...`, `go build ./...`.

Heavy Go work goes through `cohere-dev`, which takes a token from the machine's pool and runs with Go
held to the token's share of the cores, niced (#qhg0ntb): `go run ./command/cohere-dev test ./...` for the
landing gate, `test <package>` or `test --fast` while editing, `build` and `vet`, and `exec -- <command>`
for anything else heavy, a bench or a cohere run. A run that finds every token held waits its turn in
arrival order and names who holds them; `cohere-dev status` shows the pool and the line. Plain `go test`,
`go build` and `go vet` run unbudgeted on every core, and a dozen of them at once held the machine at load
110 to 180, made a 27s gate take 7 to 10 minutes, and failed tests on their deadlines. Inside a token's run
nothing more is held back: its packages and `t.Parallel` tests use the whole share.

Work lands on main through `go run ./command/cohere-dev land`, run in the worktree being landed, with
its work committed (#fz6xejy). It takes the machine's one land lock in arrival order and holds it from
merge to fast-forward: it merges current main in, gates the result (`go vet ./...` and the whole
module's tests, through a pool token), and fast-forwards main to it, checkout and all. Nothing landed
this way moves main during another's gate. A fast-forward of main by hand does, and when one happens
mid-gate, land merges it and gates again. A merge that conflicts is undone and its files named. On a
loaded morning a branch gated green eight times without landing, and every one of those gates compiled a
merge that never existed again.

The Go build cache is held under a cap, 40 GB unless `COHERE_DEV_CACHE_GB` says otherwise (#jc6ca7r).
When a token's run ends, at most once every ten minutes for the machine, `cohere-dev` looks at the
cache in the background, and when it is over the cap it takes every token, trims the least recently
used entries to three quarters of the cap, and gives the tokens back. `cohere-dev status` shows the
last look. Go alone keeps entries for five days, and the house refilled 200 GB in two and a half hours.

A gate run never passes `-count=1`, and the wrapper refuses one for the whole module. Go's test cache
skips every package whose inputs did not change, and after a one-file change `-count=1` cost 81% more
CPU for the same verdict (1,180 against 2,136 core-seconds). A flaky test is fixed rather than re-run
around; name its package to re-run it alone.

## Porting a rule

Five things, each of which caught a defect the others could not. They are ordered by what they cost
and what they find, not by importance.

**Read the upstream corpus first.** oxc ships its pass and fail cases inline in every rule file, sixty
to eighty per rule, at `~/Projects/system/oxc/crates/oxc_linter/src/rules/`. Reading them costs two
minutes. They encode the cases a porter does not think of, **by definition**: a porter who had thought
of them would have written them. A rule shipped past a twenty-case fixture pair and a three-guard
mutation sweep, and oxc's own passing case found the false positive in ten minutes.

It is also the thing that tells you what the rule **operates on**, which is what a neighbouring rule
most easily misleads you about. `no-nonoctal-decimal-escape` looks like a regex rule and is not:
`/\8/` is a pass in its corpus. Ported by analogy from the rule finished an hour earlier, the code and
its fixtures would have shared one wrong belief about the rule's surface, and every fixture would have
passed.

**Port from oxc, not from another implementation.** oxc is what the gate runs, so the differential
compares against it. A rule ported from `@typescript-eslint` or from rslint can be correct and still
read as a difference, which makes the harness less meaningful for no gain. Cite the others as a second
opinion; where they disagree, oxc wins and the disagreement earns a line in the commit.

**Measure the precondition before writing.** Count what the rule is about on the real tree, and prefer
a two-sided key: "3 violations, each carrying a disable comment" proves far more than a bare zero,
because a zero is equally satisfied by a rule that cannot see. Re-measure your own count when it moves;
a number large enough to feel like evidence stops you asking whether you counted the right thing.

**Ship a fixture pair and run a mutation sweep.** Both directions, same commit, and every mutant
confirmed to compile before it is trusted: a mutant that fails to build emits no failure line and reads
exactly like a passing sweep. A survivor is one of three things and only the first wants a fixture:

- the fixtures do not measure that branch, so write one
- the branch cannot change an answer, so delete it
- one behavior is held redundantly by two guards, so mutate them together
- the fixtures measure the right branch but assert the wrong layer

The fourth is the expensive one. A rule's flag was forced always-on and thirty fixtures stayed green,
because every fixture covering that input asserted **which rule fired** and never **what it offered**.
A finding with a wrong repair reads as a correct finding. For a rule carrying fixes or suggestions,
assert the repair rather than the id: the id is satisfied by a correct detection with a wrong fix.

**Run against the real tree before committing, not after.** Two false positives shipped past a full
fixture pair because the fixtures were written from the same wrong belief as the code. The tree is the
only check that does not share the author's assumptions.

### When the sweep itself is the thing that lied

Four faults have produced a clean sweep from a probe that never ran, and every one reads identically
to a mutant nothing catches. Three appeared twice, in different members' harnesses, within one night.

- **A mutant that does not compile.** `go test` reports `[build failed]` and zero failing tests. Assert
  the mutant builds before reading its result.
- **A mutation that did not apply.** A replacement string that does not match the source changes
  nothing and fails nothing. Assert the file changed before running tests.
- **A filter that cannot match.** `grep -c` counts matching lines per file and `grep -o` counts
  occurrences; a total built from the wrong one reports zero against output that has failures in it.
- **A restore that silently did not run.** An unset variable in the cleanup leaves the mutant in the
  tree, and a restore that did nothing looks exactly like one that worked. Cohere with `cmp`.

The shape is always the same and it is the same shape the rules themselves fail in: **the probe did not
run, and nothing said so.** A sweep is what everything else here is trusted on, so it is the one
instrument whose own failure modes have to be checked rather than assumed.

### What a fix costs

`ReportNodeWithFixes` rewrites source, and the text is only half of it: a correct replacement over the
wrong range writes the right characters into the wrong place. Pointing one deletion at an enclosing
node instead of a statement removed an entire function with every existing assertion still passing. So
`ruletest.ExpectFixedSource` is required for any rule that proposes fixes, and the fixture-pair guard
enforces it.

A port is not obliged to carry a defect it can see. Two rules here report without the fix their
original ships, because those fixers drop type annotations, lose `async`, or replace a whole
`VariableDeclaration` while reporting per declarator, which silently deletes a component. Reporting
without a fix is the subset you can show correct.

### Fidelity

A port is faithful to what a rule **decides**, not to how it **obtains what it needs**. The originals
walk the filesystem and read `process.cwd()` because ESLint hands them one file at a time and gives
them no program. `cohere` has the program. Reproducing a workaround for a constraint we do not have is
not fidelity.

Where a rule has no tree exposure and the upstream gives no reasoning, fidelity is the only available
authority, and your own sense of what the rule should do is the thing most likely to be wrong. Record
the intuitive reading beside the actual one so the next reader does not correct it back.

## The dispatcher

Rules are compiled in rather than loaded, which is what makes them free to run. The cost is that
adding a rule means rebuilding, so `command/cohere-dispatch` pays that cost automatically: it names
the binary for the checkout's HEAD commit, looks for `.cache/cohere/bin/cohere-<platform>-<hash>`,
and execs it when present or builds it first when absent. A new commit costs one rebuild; every run
after it is a stat and an exec.

**It builds from the committed tree, never the working tree.** The checkout is shared, and a gate
that built from disk ran whatever any member had half-written: on 2026-10-02 one member's
uncommitted rule work turned 22 findings on in ahra before it was classified. So a build extracts
HEAD with `git archive`, extracts the compiler at the commit HEAD pins, and builds that. The
compiler is kept under `.cache/cohere/compiler/`, one per pin, because it is 66,000 files. Before a
binary is cached, its own `--version` must name the commit, or it is refused.

```sh
cohere                    # build HEAD's committed tree if it has no binary yet, then exec
cohere --dev              # build the working tree instead, uncommitted edits (anyone's) included
cohere --dispatch-verbose # say so when a rebuild fires
```

A default run's output is cohere's alone. While the launcher builds, a terminal shows one dim line that
clears before cohere prints, and a pipe or a file gets nothing; what a prune or a trim removed goes to
`.cache/cohere/prune.log`. `--verbose`, which the launcher reads on its way to cohere, or
`--dispatch-verbose` print each step and each removal in full. A failure always prints.

Anything the dispatcher does not own is forwarded to the real binary untouched.

`--dev` is the explicit way to run uncommitted work, and it says so on every run. It builds to a
stable path because Go caches package compilation but not linking, so a hash-named binary is a new
filename and therefore a full link on every change. Measured on a comparable binary: a leaf rule edit
costs 2.03s, while a genuinely unchanged tree at a stable path costs 0.29s. It records its hash beside
the binary and compares it, so a stable name never means a stale binary. `go run ./command/cohere`
in your own checkout works too, and its `--version` reports the uncommitted changes.

The build flags are fixed at `-trimpath -ldflags="-s -w"`. They make the link marginally faster and
the binary 29% smaller, and `-trimpath` is a build-input change rather than a link flag: turning it
on invalidates the entire compile cache, measured at 33s on a warm 2.0 GB cache. Flipping it per run
would pay that repeatedly, so it does not vary. The launcher builds through the same Go cache every
`go build` and `go test` on the machine uses (#3kr3x59). The house sets `GOFLAGS=-trimpath`, so a
member's test run compiles the packages the launcher's next build needs, and a private cache only
stored them twice: it had reached 40 GB beside the default cache's 81 GB. The launcher still builds
with cgo off, which compiles only the few packages that use cgo differently, measured at 46 MB.

**Go does not bound the cache, so the launcher does.** Go trims by age alone, five days
unused, and developing cohere filled the default cache far faster: on 2026-10-03 it regrew 223 GB in
nine hours and filled the disk (#3sgjy0h), most of it `TestEveryReleaseTargetCompiles` building cohere
for six platforms on every `go test ./...`, until `3dc6e70` made it type-check them instead.
`internal/release/gocache` trims a cache of its least recently used entries once it is over a cap,
down to three quarters of it. On any run, at most every ten minutes, the launcher starts a background
copy of itself that trims the cache at 48 GB. The run pays one stat for that and never waits on the
walk. What it removed is logged to `.cache/cohere/prune.log`.

**No trim removes an entry used in the last 90 minutes**, so a burst can stand above the cap by what it
wrote inside that window. Go marks an entry used at most once an hour, so within the hour, age cannot
tell an entry nobody wants from one a running build is reading, and removing one of those fails the
build with `could not import ... no such file`, which reads as a compile error. A gate that trimmed
young entries to hold a burst under the cap did exactly that.

**A missing binary is a loud error, never a fallback.** With no Go toolchain the dispatcher runs the
binary shipped for the platform, and if there is none it exits non-zero naming the platform and
everywhere it looked. There is deliberately no path where it execs something that might do nothing:
the gate this tool replaces printed green over zero files for days because a resolver found no
binary, fell through to a bare command name, and an empty file list is indistinguishable from a
clean tree.

cohere decides what changed from content hashes and reads ignore files itself, so its own code starts
no git process. Two exceptions sit outside the checking: the dispatcher extracts the committed tree
with `git archive`, and SwiftPM clones and describes a Swift package's dependencies.

## Releasing

`cohere` reaches a machine as an npm install. One thin dispatcher package resolves a per-platform
binary package, the same shape oxlint and tsgo use:

```
@system-inc/cohere                 the package you install; a Node launcher, no binary
@system-inc/cohere-darwin-arm64    ~43 MB Go binary, stripped
@system-inc/cohere-darwin-x64
@system-inc/cohere-linux-arm64     for CI and containers
@system-inc/cohere-linux-x64
@system-inc/cohere-win32-arm64
@system-inc/cohere-win32-x64
```

```sh
pnpm add -D @system-inc/cohere
```

The command is still `cohere`, with `c` beside it: npm's `bin` names the command, not the package.
The scope is there because the bare name `cohere` is taken on npm. The platform packages carry the
dispatcher's name as a stem rather than a scope of their own, the shape biome and rollup use,
because `@system-inc` holds other tools and `@system-inc/darwin-arm64` would not say whose binary it
is. Each one stages in a directory named for its package without the scope, so `dist/` reads the
same way `node_modules/@system-inc/` does.

Every size here names the build that produced it, because the same commit measures 43.1 MB stripped
and 61.9 MB from a plain `go build`. An 18 MB spread under one label is not a measurement, it is two
measurements sharing a name, and two people quoting sizes from different builds will both be right
and still disagree.

Most of that is architecture rather than catalog. Attributed by symbol on an unstripped build,
measured at 06:23: 23.2 MB of Go runtime and type metadata, 7.2 MB more of `go:` metadata, 9.1 MB of
typescript-go, 1.54 MB of goja, and 0.18 MB of our rules. Against the 56 rules the binary reported
running in that same measurement, that is about 3.3 KB per rule, so porting the remaining hundred
adds under half a megabyte. Embedding a whole type checker and a whole JavaScript interpreter is what
costs, and both were decided long before the rules were.

Two traps live in that paragraph, and both cost someone time tonight. `go tool nm` reports every
symbol as size zero on a stripped binary, so an attribution run against a release build sums to
nothing rather than failing — the numbers above come from an unstripped build for that reason. And a
standalone program measures a dependency's marginal cost including everything it drags in, which is a
different quantity from its share of a binary that already paid for most of that: goja measures
1.54 MB by attribution and 8.47 MB as the marginal cost of adding it to an empty program. Both
are correct; quoting one to answer the other's question is not.

The platform packages are `optionalDependencies` pinned to the exact version, and npm picks one by
matching the `os` and `cpu` fields against the machine. Those fields are generated from the same
target that cross-compiles the binary, because npm's spelling and Go's disagree in two places —
`win32` against `windows`, `x64` against `amd64` — and a package published under the Go spelling
installs correctly and is never found, which on the machine is indistinguishable from a platform we
never shipped.

```sh
go run ./command/cohere-release --version 1.0.0 --output dist
```

One command builds all six from one machine, in about two minutes, and it stages rather than
publishes. It refuses a release it cannot complete: a target that fails to build fails the whole
run, because a version missing one platform resolves to nothing there and gets reported as a bug
against a release that looked fine everywhere else.

The launcher is Node rather than Go, which is the one surprising choice. A Go dispatcher would have
to be cross-compiled per platform, making it a seventh platform package and defeating the point of
installing one thing. Node is present by construction in an npm install, and `require.resolve` asks
the package manager where a package actually is instead of modeling pnpm, npm, and yarn layouts by
hand.

`cohere --version` reports the version, the platform, the Go toolchain, and the vendored compiler as
`owner/name@commit`, so a bug report names what was running rather than "latest". The stamps go in
at link time; an unstamped local build says `dev` and says why.

The compiler's repository is stamped rather than written down, because it has already moved once:
the pin was against `microsoft/typescript-go` until that repository was archived, and is now against
`microsoft/TypeScript`. Both spellings are forty hex characters and resolve in different places, so a
hardcoded label would have survived the migration while quietly becoming false.

### Versioning and the changelog

cohere is semver, starting at 1.0.0, and `release.Build` refuses any version below that or not semver
(a prerelease like `1.0.0-rc.1` is allowed). What each part means, as Kirk set it:

- **Major** breaks a configuration or a command line that worked: a key, a flag or a set name removed or
  reinterpreted, so an existing `CohereSettings.json` or script no longer means what it did.
- **Minor** can fail a project that passed: a rule turned on in a set, a severity raised, a rule's
  options changed, a setting like an ignore pattern changed. A project takes that by taking the minor.
- **Patch** can only relax or fix: a rule turned off, a severity lowered, a bug fixed.

The rule sets are where most of a release's visible change lives, and they can be read exactly, so their
part of `CHANGELOG.md` is generated:

```sh
go run ./internal/release/tools/changelog --version 1.1.0 --previous 1.0.0
```

That diffs every set between the tag `v1.0.0` and `HEAD` and prints the release's section, with one
heading per changed set listing each rule turned on or off, each severity and option change, and each
other setting that moved. It also refuses a version smaller than the diff allows: a set that turned a
rule on can't ship as a patch, and a set that was removed can't ship without a major. The section goes
above the last one in `CHANGELOG.md`, and what changed in cohere outside the sets is written under the
same heading by hand. Each release is tagged `v<version>`, which is what the next one diffs against.

The Swift engine's contract is not a compatibility surface. The front door and `cohere-swift` ship in
one package, and the release refuses to stage a pair that speak different contract versions, so a
contract change needs no version of its own.

### The formatter carries no JavaScript

A released binary formats with cohere's native Go printers and embeds none of the Prettier fork's
bundles, so `--version` names no formatter: there are no bytes in it for a stamp to vouch for. The
bundles remain in `internal/format/prettier` as the test oracle the printers are measured against,
and `prettier.DigestBundles` keys the oracle's cache so a rebuilt fork invalidates every cached
answer rather than comparing against output from bundles nobody loads.

`COHERE_BINARY=/path/to/cohere` points every `cohere` on the machine at a local build. A broken
override is fatal rather than a fallback, even when a good install is sitting right there: someone
who sets it has stated which binary they want, and quietly running a different one would hand them
results they would read as their own build's.

### macOS signing

Measured, because the answer decides how much this matters. An unsigned binary with no quarantine
attribute runs normally, and that is what an `npm install` produces — package managers extract
tarballs without setting `com.apple.quarantine`, so consumers installing from the registry are not
blocked. The same binary *with* quarantine set is killed by the kernel: exit 137, SIGKILL, and zero
bytes on both stdout and stderr.

That silent kill is the reason to sign. It is reached whenever the binary travels as a file rather
than as a package — a release asset from a browser, a binary copied out of CI — and it fails in the
worst available way, with no output to explain it. The launcher's signal handling turns it into a
loud 137 rather than a green nothing, and signing removes it entirely.

Signing is wired and credential-gated. It needs a **Developer ID Application** certificate
specifically; an Apple Development certificate signs successfully and is then rejected by the notary
service at the end of a release. Staging without credentials is supported and says so in its own
summary, because an unsigned release is fine and an unsigned release that looks signed is not.

Note that a notarization ticket cannot be stapled to a bare executable — only to a bundle, a disk
image, or an installer package — so a notarized Mach-O is validated by an online check on first
launch. A machine that is entirely offline the first time it runs a quarantined `cohere` is still
blocked. The npm path does not set quarantine, so this does not affect it.
