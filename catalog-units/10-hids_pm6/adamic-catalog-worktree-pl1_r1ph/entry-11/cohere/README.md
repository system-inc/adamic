# cohere

cohere type-checks, lints, fixes and formats a TypeScript codebase in one process. The TypeScript
compiler builds the program once, and that one type graph is what TypeScript's own diagnostics and
every lint rule read, the rules in one walk of each file. The fix engine and the formatter run in the
same process. Fixes re-read each file from disk and re-lint it until they stop landing, so they never
apply to a stale copy. They land only in files the lint walk checked: a file the formatter reaches that the
program does not hold, such as a bundle beside a tsconfig that does not include it, is formatted and never
fixed. The formatter parses each TypeScript file it formats with the same TypeScript
parser, and prints every file it formats with cohere's own printers. Keeping the rules on one graph in
one walk is what lets a codebase carry hundreds of them without each one costing another pass over the
code.

## Install

```sh
pnpm add -D @system-inc/cohere
```

npm and yarn work the same way. The package installs two commands, `cohere` and its short form `c`,
and pulls in one prebuilt package for your platform: macOS, Linux or Windows, each on arm64 or x64.
No Go toolchain is needed. If no binary matches your platform, `cohere` exits with an error naming
the platform rather than doing nothing.

## Swift, on macOS only

A directory with a `Package.swift` at its root is checked as a Swift package, by `cohere-swift`, an
engine the two macOS packages ship beside `cohere`. It runs on macOS 15 and later, and types the
package with the Swift toolchain already on your machine. On Linux and Windows there is no Swift
engine, and `cohere` refuses a Swift package by name and exits 1 rather than checking nothing.

## Running it

Run it anywhere in a repository. It starts from the project you're in, the nearest `tsconfig.json` or
`Package.swift` at or above you, or from the repository's root when there's none, and checks that
project and every project below it: each TypeScript program and each Swift package, each with its own
engine and its own settings, all at once. A repository holding more than one gets one report, a section
per project, a line per project and one summary, and the exit code is the worst of theirs, so it's green
only when every project is. A run that finds no project fails, and so does one whose every project was
skipped, such as a Swift package alone under `--format-only`: it checked nothing.

Finding projects never enters what `.gitignore` ignores, at any level, nor `node_modules`, `.build`,
`.cache` or `testdata`, nor a directory that is a repository of its own: a submodule stays part of the
program above it, as it always was. `--directory`, `--tsconfig` or a path narrows the run to one
project; `--lint-config` doesn't, and every project runs under the settings it names.

Each file is checked once, by the nearest `tsconfig.json` that includes it, the one your editor opens it
under. Another tsconfig that includes it still types against it and doesn't report it, and each project
formats its own directory, leaving the directories of the projects below it to theirs. A solution
`tsconfig.json`, `"files": []` with `references`, includes nothing to check: what it references is
checked instead, whatever those tsconfigs are named.

```sh
cohere                  # type-check, lint, apply every available fix, and format, in one call
cohere --fix            # apply fixes and format, running no other phase
cohere --no-fix         # write nothing: report what fixing and formatting would change, and exit on it
cohere --lint           # report lint findings only: no fixes, no TypeScript diagnostics
```

A bare `cohere` formats the files not on record as formatted at their current bytes, so a warm run
formats what you edited and nothing else. Add `--no-format` to any of the first three to leave
formatting out.

A run prints the files it rewrote, then its findings, then one line:

```
🪄💅 app/os/SessionRow.tsx      prefer-const ×2, prefer-nullish-coalescing
  💅 modules/pensieve/Recall.ts
app/os/Session.ts:12:5 error nexus/consistency-no-abbreviated-identifier `ctx` is an abbreviation.
✗ ☠️ 0.8s • 1 finding • 2 cohered (480 rules • 3 checked • 3,923 cached)
```

See [Output](#output) for what that line says, and for `--verbose` and `--json`.

### Adamic's `.a` files

A tsconfig that lists `".a"` in a top-level `"sourceExtensions"` makes its `.a` files TypeScript source
under their own names: type-checked, linted and formatted exactly as the same files named `.ts`, and
printed as `.a` everywhere. Imports name the extension, `import { area } from './geometry.a'`. The key is
read by cohere's compiler, System, Inc.'s fork of TypeScript; `tsc` ignores it.

`.a` is also a static library's extension, so cohere never reads a `.a` file a tsconfig doesn't claim: a
`libfoo.a` beside your code is left alone. A claimed `.a` file that `.gitignore` ignores (C and C++
templates list `*.a`) is checked and not formatted, and the run names it and says so in the footer.

## CohereSettings.json

With no settings file, cohere applies the house stack, each set where the code shows it fits:

- `cohere:typescript` on every file. It composes two sets a configuration can also name on its own:
  `cohere:adamic`, the soundness set, whose rules hold types to the truth, and `cohere:style`, the taste
  set.
- `cohere:react` on each file that imports `react` or a `react-` package, or contains JSX.
- `cohere:next` on each file that imports `next`, and, once anything does, on Next's own files: everything
  under `app/` and `pages/`, and `middleware`, `instrumentation` and `next.config`.
- `cohere:tailwind` on every file when the root stylesheet the Tailwind rules find (`app/globals.css` and
  the others) imports `tailwindcss`. Without one, its rules are skipped by name.

package.json is never the evidence: a dependency the code never imports applies nothing. The run's
`sets:` line names each set and why. The house format applies too, so the first run rewrites files into
the house style.

To go your own way, write a `CohereSettings.json`. One that names no `cohere:` set keeps the detected
stack and applies its own rules and format on top:

```json
{
    "rules": { "no-continue": "off", "eqeqeq": ["error", "smart"] },
    "format": {}
}
```

An `"off"` there needs no reason, and `--coverage` names each one. A `format` block applies over
Prettier's defaults, so `{}` means Prettier's defaults. A file that names sets in `extends` chooses them
instead, and they apply to every file: `"extends": ["cohere:react", "cohere:next"]`.

A project that extends a `cohere:system-inc/*` set is one of ours, and stays strict: every off says why
under `reasons`, every departure from an inherited ruling says why under `departures`, and the format
comes only from the Nexus tier, so a `format` block anywhere else in the chain is refused.

The keys:

- `extends`: a rule set cohere carries (`cohere:<name>`), a path to another settings file, or a list of
  them, applied first and in order. A rule two sets that don't extend each other both configure is an
  error naming both.
- `rules`: each rule's severity, `"off"`, `"warn"` or `"error"`, with options where a rule takes them.
- `overrides`: a list of `{ "files": [...], "rules": {...} }` blocks that change rules for matching
  paths.
- `ignorePatterns`: paths cohere never checks, beyond what git ignores. Ignore matching follows
  gitignore(5): every `.gitignore` from the root down and `.git/info/exclude` (not `core.excludesFile`),
  case sensitive (`core.ignorecase` is not read). A `.gitignore` that is a symbolic link or larger than
  100 MiB is refused by name rather than skipped.
- `reasons` and `departures`: why this file turns a rule off, or sets it differently from the file it
  extends. Required in our tiers, optional elsewhere. cohere prints them, so a choice stays visible.
- `format`: the formatter's options, `printWidth`, `tabWidth`, `useTabs`, `semi`, `singleQuote`,
  `trailingComma`, `bracketSpacing`, `bracketSameLine`, `arrowParens`, `endOfLine`, and `ignore`, a list
  of paths the formatter leaves alone. An option outside that list is an error.

The full reference, every key with an example, is [schema/CohereSettings.md](schema/CohereSettings.md).
It is generated from the loader itself, along with two JSON schemas an editor can validate against
through a `"$schema"` key: [schema/CohereSettings.schema.json](schema/CohereSettings.schema.json) for a
project's file and [schema/NexusCohereSettings.schema.json](schema/NexusCohereSettings.schema.json) for
the Nexus tier. Both ship in the package, so the path works offline and matches the installed cohere:

```json
{
    "$schema": "./node_modules/@system-inc/cohere/schema/CohereSettings.schema.json"
}
```

`cohere --rules` lists every rule cohere implements for your project's language, and
`cohere --rules-enabled` lists the ones your settings turn on for one file (`index.ts` at the project
root unless you name another).

## Flags

| Flag | What it does |
| --- | --- |
| `--fix` | apply fixes and format, running no other phase (`--no-format` to leave formatting out) |
| `--fix-passes N` | how many times a file may be re-linted while fixes keep landing (default 10) |
| `--no-fix` | write nothing to your source; report what fixing and formatting would change, and exit nonzero if anything would |
| `--no-format` | leave formatting out of a bare run, `--fix` or `--no-fix`: fix, type-check and lint only |
| `--format` | format the files not on record as formatted, or the paths you name; a bare run, `--fix` and `--no-fix` already do, so naming it there changes nothing |
| `--format-all` | format every file, not only the ones not already on record as formatted (implies `--format`) |
| `--format-only` | format only, proposing no fixes and running no other phase (implies `--format`); with `--no-fix`, the commit gate's format check (see [Before you commit](#before-you-commit)) |
| `--lint` | run the lint rules only, reporting what they find without fixing it, and without TypeScript's diagnostics |
| `--types` | report TypeScript's diagnostics only, running no rules and fixing nothing |
| `--unused` | report code that is never used: unreferenced exports and unreachable statements |
| `--unused-all` | list the unused findings already marked `cohere-keep`, not only count them (implies `--unused`) |
| `--unused-deep` | also group the unused code into islands by what reaches what (implies `--unused`) |
| `--directory PATH` | the project root (default: the nearest `tsconfig.json` or `Package.swift` above you) |
| `--tsconfig PATH` | the tsconfig that defines the program (default: the nearest `tsconfig.json`) |
| `--lint-config PATH` | the settings file, relative to `--directory` if given, else to where you run it (default: `CohereSettings.json` at the project root) |
| `--no-cache` | read and write no cache, so every phase computes from source |
| `--stdin-filepath PATH` | with `--fix`, read one file from stdin and print the fixed text, writing nothing |
| `--explain FILE` | report what every rule did on one file, and why it ran or did not, writing nothing |
| `--print-config` | print each rule's resolved severity and options for one file (`index.ts` unless you name one), as JSON, and exit |
| `--rules` | print the rules cohere implements for your project's language, and exit |
| `--rules-enabled` | print the rules your settings turn on for one file (`index.ts` unless you name one), with severity, and exit |
| `--verbose` | print everything a run can say: each phase, the coverage summary, overrides, skips, notes, memory and the total |
| `--phases` | put where the time went (graph, fix, format, types, lint) first in the footer's parentheses |
| `--json` | print newline-delimited JSON for a program to read instead of the human view (see [Output](#output)) |
| `--coverage` | name every rule under the coverage fact that describes it, not only count them |
| `--adamic-readiness` | say what share of the files are Adamic-ready, in the footer and in `--json` (see [Output](#output)) |
| `--timing` | report what building the graph cost and the CPU each rule cost, most expensive rule first |
| `--single-threaded` | use one type checker instead of several |
| `--profile FILE` | write a Go CPU profile of the run to FILE |
| `--cache-dump` | print what this project's cache holds, and exit |
| `--version` | print the version, what this binary was built from, and the Swift contract it speaks, and exit |

## Output

A run prints three things, in order:
- **The files it rewrote.** Each line has a 🪄 if fixes were applied and a 💅 if it was formatted, then
  the path, then the rules whose fixes it took, with a count past one. Past 20 files it says how many
  more, and `--verbose` lists them all.
- **Its findings,** one per line, as `path:line:col severity rule message`.
- **Any crash,** one per line, as `path crash rule message`: a rule cohere could not finish on a file, or
  a file it could not check at all. That's a bug in cohere, not in your code, and it fails the run (see
  [Reporting bugs](#reporting-bugs)).
- **One footer line:** the verdict (✓ 💎 or ✗ ☠️), how long the run took, what it found, how many files
  it cohered, and in the parentheses how much it covered.

```
✓ 💎 2.4s (480 rules • 3,926 checked)
✓ 💎 0.7s • 2 cohered (480 rules • 3 checked • 3,923 cached)
✗ ☠️ 0.8s • 1 type error • 2 findings (480 rules • 3 checked • 3,923 cached)
✓ 💎 2.6s (488 rules • 3,926 checked) • 87% Adamic-ready (3,412 of 3,926)
```

The last line is a run with `--adamic-readiness`.

The words mean exactly this:
- **cohered:** files cohere rewrote, fixed or formatted, which are the files listed above the footer.
  Shown only when there were any.
- **checked:** files examined fresh in this run.
- **cached:** files the cache answered for, unchanged since a run that checked them. Checked and cached
  together are every file in scope.
- **rules:** the rules that ran.
- **Adamic-ready:** shown only with `--adamic-readiness`. The share of linted files whose types are true: no
  type error and no `cohere:adamic` finding, counted before any disable comment and whatever your settings turn on, since Adamic, which
  compiles TypeScript native, reads neither. It names the compiler options Adamic sets that your tsconfig
  leaves off (`strict`, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`), and says `not measured`
  with the reason when it could not measure, never a number it didn't. `--json` gives it as the summary's
  `adamic`, which is absent without the flag. Measuring runs the set's rules your settings leave off, at the
  set's own options, without reporting them, so a run that doesn't ask pays nothing for it.

A count is exact, with its thousands grouped (`3,923`); `--json` gives the plain integer. A count of zero
is left out, so a cold run shows no `cached` and a run with nothing changed no `checked`.

Anything the run did not check is in the footer even when it passes, so a green line never hides a
gap: `✓ 💎 2.4s (…) • 💅 formatting not checked`. The same goes for a phase that could not run, a rule
that skipped every file, a run narrowed to some of the files, and an Adamic `.a` file git ignores. A crash
is in the footer too, and it never passes.

On a terminal the verdict and time are bold, what was found is red, and the parentheses are dim. A
pipe, a file or `NO_COLOR` gets no color codes at all.

`--phases` puts where the time went first inside the parentheses: 🕸 building the graph (read, parse,
bind), 🪄 fixing, 💅 formatting, 🔷 the type check, 👑 lint, and 🧹 unused when `--unused` ran. A phase
that did not run is left out. Each file is formatted once its fixes settle, so the two interleave: 💅 is
the time a format was running and 🪄 the rest of the fix phase, and together they are its time. To make that a project's default, set it in `CohereSettings.json`:

```json
{ "output": { "phases": true } }
```

`--timing` is separate and still prints what each rule cost, as CPU read from each walk thread's own
clock, so a rule that waits or is descheduled costs nothing. Reading that clock around every call adds
overhead; the phase times are free.

`--verbose` prints everything a run can say: each phase and why any did not run, the coverage summary,
overrides, skips, notes, memory and the total. Its footer adds the syntax nodes the run walked, and
says when the whole run was replayed from the cache. A note, about the cache, line endings or a scope
widened to the whole tree, prints only there; one that names something left unchecked is a gap the
default footer says too.

`--json` is for a program to read. It prints newline-delimited JSON, one object per line, each with a
`kind`: a `finding` per finding, a `fixed` and a `formatted` per rewritten file, a `crash` per crash,
and a `summary` last.
The summary carries a `schemaVersion`, the verdict, the timings and counts, and `gaps`, everything the
run did not check, which a program deciding whether to trust a passing run should read too.
[schema/CohereOutput.schema.json](schema/CohereOutput.schema.json) describes every field. Read
`--json` rather than the human view, whose layout can change.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | no problems found |
| `1` | it found problems, cohere crashed on a file, cohere itself could not run, for example on a settings error, or there was nothing to check: no project found, or every one skipped |
| `2` | the command line was wrong, such as an unknown flag |

`1` covers both findings and failures, so in CI the output says which one it was.

## Before you commit

```sh
cohere --no-fix --format-only --format-all
```

This is the commit gate's format check, and its exit code answers one question: would formatting change
any file, in this repository or in the nested repositories (declared submodules) it reads? It runs no
fixes, no type check and no lint, so a lint finding has no say in it.

- `0`: no file would change, here or in any nested repository.
- nonzero: each file that would change is a finding, a nested one naming its repository. So is each file
  the formatter could not read, such as one that does not parse, and a format walk that failed. A run that
  could not check a file never exits `0`.

Only declared submodules are read. A git repository inside yours that `.gitmodules` does not name (a
clone in an ignored directory, say) is skipped and nothing in it is checked: the scope line lists it
under `skipped nested repositories`, and `nested repositories:` does not count it. Declare it as a
submodule, or run the gate inside it.

Drop `--format-all` to check only the files not on record as formatted at their current bytes, or name
paths to check only those. Without `--no-fix`, `cohere --format-only` formats and writes nothing else.

## Caching

cohere keeps its cache in `.cache/cohere` at the project root, so a run reuses what an earlier run
already established. When your tsconfig sets `incremental`, it also keeps TypeScript's build information
where the tsconfig says, and a `--no-fix` run leaves that file untouched. Add `.cache/` to your
`.gitignore`. `--no-cache` reads nothing from those caches and writes nothing to them; use it when
you suspect the cache, or to time a run from scratch. Leave the files in `.cache/cohere` to cohere: it
reads source text straight from the ones it mapped, so a tool that truncated one in place while a run
was reading it would end that run with a memory fault. Deleting the directory between runs is always
safe.

## CI

```sh
cohere --no-fix
```

`--no-fix` makes the run read-only, so CI reports what a developer's run would have fixed and
formatted instead of doing it on a machine nobody looks at. A nonzero exit fails the job, unformatted
files included. `--no-fix` does not stop cohere writing its own cache in `.cache/cohere`,
which is safe to persist between CI runs to speed them up, or to discard.

## Verifying a release

Every release is built by [release.yml](.github/workflows/release.yml) in GitHub Actions, from a commit
on this repository, and you can check what you installed came from it.

cohere checks part of this on every run. `@system-inc/cohere` ships a `SHA256SUMS` of every binary in
its platform packages, and before running the binary it checks the binary against it. A binary that
differs, from a corrupted download or cache or from tampering, is refused by name and never run. Setting
`COHERE_BINARY` to a build of your own skips the check, since no release describes that build.

To check it yourself, run three commands from your project's root, putting your version and platform in.
The first runs where you are; the other two run from the directory holding both cohere packages, which
the `cd` finds the same way for npm and pnpm installs:

```sh
npm audit signatures

cd "$(dirname "$(realpath node_modules/@system-inc/cohere)")"
gh release download v1.0.0 --repo system-inc/cohere --pattern SHA256SUMS --output - | shasum -a 256 -c --ignore-missing
gh attestation verify cohere-darwin-arm64/bin/cohere --repo system-inc/cohere
```

- `npm audit signatures` checks the registry's signature on every package installed, and the provenance
  each cohere package was published with, which names the workflow run and commit that built it. npm
  shows the same provenance on each package's page.
- The `SHA256SUMS` attached to the GitHub release is the same file as the one inside the package, but it
  reaches you without going through npm, and `shasum` checks the installed binaries against it. A line
  that ends `OK` is a match. It exits nonzero on any mismatch, and when it finds none of the files, so
  running it from the wrong directory fails rather than passing.
- `gh attestation verify` checks the binary against the build attestation the release made for it,
  signed through Sigstore by the workflow and stored on this repository. Every binary, the `SHA256SUMS`
  and the VS Code extension are attested.

The commands are written for a macOS or Linux shell.

## Reporting bugs

Report a bug at <https://github.com/system-inc/cohere/issues>. Include the output of `cohere --version`,
the rule, the smallest code that shows it, and what cohere printed. A crash line already says most of
that, so paste it whole.

## Contributing

How cohere is built, and why, is in [CONTRIBUTING.md](CONTRIBUTING.md).

## License

cohere is licensed under either of the Apache License, Version 2.0 ([LICENSE-APACHE](LICENSE-APACHE))
or the MIT license ([LICENSE-MIT](LICENSE-MIT)), at your option. cohere is built on the TypeScript
compiler, whose notice is in [NOTICE](NOTICE), and on the projects credited, each with its license, in
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md). Every package it publishes carries all four files.

Unless you explicitly state otherwise, any contribution you intentionally submit for inclusion in cohere,
as defined in the Apache-2.0 license, is dual licensed as above, without any additional terms or
conditions.
