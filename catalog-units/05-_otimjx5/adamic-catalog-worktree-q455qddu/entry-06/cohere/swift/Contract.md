# The engine contract

What an external engine prints and how the Go front door renders it. cohere-swift is the first
external engine; the TypeScript engine stays in process and never speaks this. The contract exists
so a Swift run reads exactly like a TypeScript run: same finding lines, same phase line, same
coverage notes, same accounting, same exit code. Someone who knows how to read one already knows
how to read the other.

The fixtures in `Contract/` are the agreement made executable. The Go renderer's tests and the Swift
encoder's tests both read them, so neither side can drift without the other's test failing.

| Fixture | What it pins |
|---|---|
| `Clean.jsonl` | Every phase ran or was reused. One file is excluded with its reason. Exit 0. |
| `Findings.jsonl` | A compiler warning with its group, and a rule finding whose message has a newline. Lint still runs after a warning. The binary was built from a modified tree. Exit 1. |
| `TypesBail.jsonl` | Under `--no-fix`, a compiler error cuts lint off as `notReached`. The run is incomplete. Exit 1. |
| `CrashWithoutSummary.jsonl` | The stream stops after the fix phase. The front door must say nothing was checked and exit 1. |
| `Unreadable.jsonl` | One file the engine could not read: named beside the excluded files, every phase ran, nothing found, and still incomplete with exit 1. |
| `LintOnly.jsonl` | `--lint`, a real run: fix and types skipped as `not requested`, lint ran and found two, and the summary calls the run complete, which it is. Exit 1, for the findings. |
| `Unused.jsonl` | `--unused`: after lint, each rule's `unused` records then its `unusedCoverage`, `unused-import` first and `unused-declaration` second (the declaration section is from a real run on ahraos-presence), the unused phase ran, and the summary counts no findings and exits 0, because the report is not a gate. |

**Version 2** (2026-10-02) added the `unreadable` record. Before it, a file the engine could not read was a
note on stderr: the summary said the run fell short and the front door believed it, but could not say
which files. Both sides moved to 2 in one commit, and each refuses the other version.

**Version 3** (2026-10-03) removed the summary's `nothingToCheck`. It named why a run had nothing to look
at, which only `--changed` over an unchanged tree ever produced, and `--changed` is gone from both
engines (ff50e94, 68e37a8). This engine had written it empty since, so the field, its fixture
(`NothingChanged.jsonl`) and the front door's `0 files checked` rendering went together. A summary over
a run that skipped every phase is now the gap it is, and one that calls itself complete is refused.

## Why records and not text

The engine could print cohere's text lines itself. It does not, because the lines are the front
door's to keep. `edit.Summary.String` carries a comment about a refusal breakdown that drifted
away from the count it qualifies. `printRuleDiagnostic` carries one about message newlines that
silently dropped four findings. Each printer carries rules like these, learned the hard way. A
second implementation of those printers in Swift would have to relearn every one, and the two
would disagree the first time either changed. So the engine reports what happened as data, and
the front door feeds that data to the printers it already has. A `phase` record becomes a
`pipelineReport.record` call. A `fix` record becomes an `edit.Summary` printed by its own `String`.

## The transport

- **stdout:** one JSON object per line, UTF-8, nothing else. A line that does not parse as a
  record is a protocol error. The front door stops and says the engine is broken; it does not skip
  the line, because a skipped line is a finding nobody sees.
- **stderr:** prose for a human, passed through untouched. Notes that the TypeScript engine prints
  to stderr (a submodule it could not read, a file lost from scope) go here. A file the engine could
  not read is not one of them: it is an `unreadable` record, so the front door can name it.
- **Records stream in pipeline order.** The front door prints each one as it arrives, so a slow
  types phase shows its fix line first, the way the TypeScript engine does.

## The invocation

The front door resolves the project root (the nearest directory holding `Package.swift` or
`tsconfig.json`, nearest wins), then runs:

```
cohere-swift --contract 2 --root <absolute package root> [flags] [paths]
```

`--contract` is the version the front door speaks. An engine speaking a different version refuses
the run, prints a sentence on stderr naming both versions, and exits 2. The first record the engine
writes is `provenance`, which carries its own `contract` field. The front door refuses a mismatch
there too, because the record and the flag are two independent ways for the versions to disagree.

## Records

Every record has `kind`. Fields not listed are not part of the contract and the front door ignores
them. Durations are integer milliseconds. Paths are absolute, as `sourceFile.FileName()` is for
TypeScript.

### `provenance`, always first

```json
{"kind":"provenance","contract":3,"engine":"cohere-swift","version":"0.1.0","commit":"<40 hex or dev>","sourceTreeModified":false,"toolchain":"swiftlang-6.4.0.34.1","swiftSyntax":"604.0.0","swiftFormat":"604.0.0"}
```

`sourceTreeModified` drives the same warning the TypeScript run prints: this binary was built from a
modified tree, so no commit reproduces these findings. `--version` prints exactly this record and
nothing else.

### `project`, the Swift counterpart of `graph built in`

```json
{"kind":"project","root":"/…/ahraos-macos","package":"AhraOs","elapsedMilliseconds":412,
 "filesInPackage":152,"filesOurs":149,"filesInScope":149,"scopeDescription":"",
 "targets":[{"name":"AhraOsCore","kind":"library","files":14,"languageMode":"6"}],
 "excluded":[{"file":"/…/Scratch.swift","reason":"ignored by git"}]}
```

- `filesInPackage` is every `.swift` file in a root target. `filesOurs` is the subset that is
  not ignored by git and not excluded. Dependencies, including local ones like `Vendor/SwiftTerm`, are
  never counted, because they are not root targets.
- `filesInScope` equals `filesOurs` on a whole-package run. A run narrowed by named paths sets it
  lower and describes the request in `scopeDescription`, which the front door prints in the
  parenthesis of `N in scope (…)`.
- `elapsedMilliseconds` is the time spent describing the package. The accounting line bills it as
  the graph, because it is the step nothing can skip.
- Written on every run that checks anything, since every such run describes the package first.

Rendered (whole package, then scoped):

```
package described in 412ms — 152 Swift files in the package, 149 of them ours
package described in 412ms — 152 Swift files in the package, 149 of them ours, 3 in scope (Sources/AhraOsIpc)
```

Excluded files print as coverage notes under the lint line, grouped by reason. Up to five files of a
reason are named, `  not checked: <file> (<reason>)`; more are counted, `  not checked: <n> files
(<reason>)`, because 107 vendored files named one per line buried the verdict on the first real run.

### `unreadable`, between `project` and the first `phase`

```json
{"kind":"unreadable","file":"/…/Latin1.swift","error":"it is not valid UTF-8 or could not be opened (…)"}
```

One per file the engine could not read, so nothing in it was checked. `error` is a sentence. The front
door prints each under the lint line with the excluded files, grouped and counted past five the same
way, as `  not checked: <file> (could not be read: <error>)`, and counts it as a gap: a summary calling
the run complete with any `unreadable` record is a protocol error. An `unreadable` record before
`project` or after any `phase` is refused, so every file a run could not read is known before a phase
reports on the files it could.

### `finding`

```json
{"kind":"finding","source":"rule","file":"/…/ByteRing.swift","line":159,"column":61,"endLine":159,"endColumn":62,
 "severity":"error","rule":"cohere-swift/force-unwrapping","messageId":"forceUnwrap",
 "message":"…why, not only what…","fixes":[],"suggestions":[]}
```

- `source` is `compiler`, `rule`, or `format`.
- `line` and `column` are 1-based. Columns count UTF-8 bytes, which is what the TypeScript printer
  already prints (`GetECMALineAndByteOffsetOfPosition`) and what swift-syntax, sourcekitd and the
  compiler's serialized diagnostics all use. `endLine` and `endColumn` are exclusive and optional.
- `severity` is `error` or `warning`. Both are findings: both count toward the exit code. The
  severity is kept so the line can say which one the compiler called it.
- `fixes` hold edits as `{"start":<byte>,"end":<byte>,"text":"…"}`, using 0-based UTF-8 byte
  offsets into the file as the engine read it. The engine applies fixes itself, in its own fix phase.
  Fixes in a record describe what a `--no-fix` run withheld; they are never instructions for the front door.
- `suggestions` are `{"message":"…","fixes":[…]}`: repairs offered and never applied.

Rendered, one finding per line. Newlines in `message` are collapsed to spaces, as
`singleLineDescription` does:

```
/…/ByteRing.swift:159:61 - force unwrapping crashes on nil … [cohere-swift/force-unwrapping/forceUnwrap]
/…/Broken.swift:3:23 - error: cannot convert value of type 'String' to specified type 'Int'
/…/Warning.swift:2:9 - warning: variable 'neverMutated' was never mutated; consider changing to 'let' constant [#VariableNeverMutated]
/…/Unformatted.swift:12:1 - not formatted the house way; a run without --no-fix rewrites it [cohere-swift/consistency-require-formatting/notFormatted]
```

The compiler line is the TypeScript shape `file:line:col - error TS2322: message`. Swift
diagnostics have no number, so the compiler's warning group takes the place of the code, written
the way swiftc writes it, `[#Group]`. A compiler finding with no group has no bracket. For compiler
findings, `rule` holds the group name without the `#`, or `""`.

### `unused` and `unusedCoverage`, only under `--unused`

```json
{"kind":"unused","file":"/…/Main.swift","line":1,"column":1,"endLine":1,"endColumn":18,
 "rule":"cohere-swift/correctness-no-unused-import","messageId":"unusedImport","message":"…why…","subject":"import Foundation",
 "suggestions":[{"message":"Remove `import Foundation`","fixes":[{"start":0,"end":18,"text":""}]}]}
{"kind":"unusedCoverage","rule":"cohere-swift/correctness-no-unused-import","filesChecked":217,
 "filesNotChecked":{"it has #if, and the index describes only the configuration the build compiled":1},
 "checked":471,"skipped":{"re-exported with @_exported, which is API":0},"found":15,"elapsedMilliseconds":2525}
```

The `--unused` report: code that was written and is never used. It is the Swift counterpart of the
TypeScript unused report and keeps its rule, a report and not a gate. An `unused` record carries a
finding's place and words, but it is not a `finding`: the summary does not count it and it never
moves the exit code. Failing a build over code that is safe to remove and never breaks one would make
the report something people route around rather than read.

- Each `unused` record names one thing to remove. `subject` is that code as written and short
  (`import Foundation`), for the report's one line per item. `suggestions` hold the removal, offered
  and never applied: the report does not rewrite.
- Exactly one `unusedCoverage` record follows a rule's `unused` records, and its `found` equals how
  many came. It states the population beside the result, so a report that checked nothing never reads
  like a report that found nothing: `filesChecked`, `filesNotChecked` by reason (a file the index
  cannot vouch for), `checked` (the items judged), and `skipped` by reason (items never reported by
  design).
- Both come after the `lint` phase record and before the `unused` one. An `unused` phase that `ran`
  without an `unusedCoverage` record is refused, as is a count that disagrees.

Rendered after the lint line, in the TypeScript report's shape:

```
unused: a report, not a gate — nothing here fails a build

  imports nothing uses — 2
    /…/Main.swift:1:1 — import Foundation
    /…/Tally.swift:2:1 — import Combine
  looked at 2 files and 5 imports (cohere-swift/correctness-no-unused-import)
  not checked for unused imports: 1 files (it has #if, and the index describes only the configuration the build compiled)
  not checked for unused imports: 1 files (the build has not compiled it as it stands)
  never reported: 1 imports (re-exported with @_exported, which is API)
```

Files not checked and items skipped are counted by reason, sorted by reason. The section header and its
lines are left out when nothing was found.

`unused-import` reads the build's index store. An import is used when something written in the file
resolves into its module, or into a module it re-exports, or when a declaration the file refers to
names that module in its signature (the index leaves out some member references, so a value of the
module's type reached through a closure parameter is caught by the signature that hands it over). An
import that is used is still reported, with `messageId` `redundantImport`, when the file's other imports
already bring everything it brought: `import AppKit` beside `import SwiftUI`, which re-exports AppKit
whole. That check believes only re-exports read from each module's own text: a `.swiftinterface`'s
`@_exported import` lines, and for a Clang module the headers its map re-exports and what they import on
lines the build certainly compiled, a whole module or one header at a time. It tries the widest import
first so the narrower one stays, and judges each removal against what the ones before it left, so every
import reported can go together. A file the index cannot vouch for is not checked: one the build has
not compiled as it stands, one with `#if` (the index describes the configuration the build compiled),
one with a reference no module claims. `@_exported` imports are API and never reported.

`unused-declaration` reads the same index, after `unused-import`, and writes its own `unusedCoverage`
record. A declaration `private` or `fileprivate`, or inside a type or extension that is, is judged
against its own file's record, since nothing outside the file can name it. Any other declaration is
judged against every record of the package, its tests included; one declared public in a library
product's target, or in a target a library product re-exports, is API and never reported. A declaration
is used when a record holds a reference to it from outside its own text (a function that only calls
itself is not used), when another declaration overrides or witnesses it, when a call of its struct's
synthesized memberwise initializer passes it, or, for a property wrapped by an attribute, when its
`$name` or `_name` is referenced. `subject` is the keyword and the name (`func after(_:_:)`,
`var stopping`), and the suggested removal takes the declaration's lines with the comments written above
and trailing it. Only the outermost of nested unused declarations is reported, so every one reported can
go together. What the index cannot see, or what removing would change without breaking the build, is
counted under `skipped` by reason and never reported: overrides and witnesses, what the Objective-C
runtime reaches or can find by name, XCTest's test methods, entry points and previews, members the
compiler calls by name, declarations an attribute may register and the types that hold them, Codable's
coding keys, initializers, a stored property of a type whose conformances reach `Encodable` or
`Decodable` or what the index cannot see into, of a SwiftUI view (SwiftUI compares them to decide when to
draw), or of a type descending from another package's (which may read it by reflection), an instance's
stored property whose initializer runs code, a field of a struct of plain numbers (its bytes may be a
shader's constants), cases an enum's conformances or raw values may reach, and a name that appears inside
`#if` or in a file the build has not compiled as it stands. Files are left unchecked for the reasons
`unused-import` gives.

### `fix`, the fix and format phase's summary

```json
{"kind":"fix","filesConsidered":149,"filesRewritten":2,"fixesApplied":5,"fixesRefused":1,
 "refusalsByReason":{"overlaps another fix":1},"filesReformatted":2,"filesNotFormatted":0,
 "notFormattedReasons":{},"formatScope":"Sources/AhraOsCore",
 "changedFiles":[{"file":"/…/Pane.swift","fixedBy":{"cohere-swift/private-over-fileprivate":3},"formatted":true},
                 {"file":"/…/Ring.swift","fixedBy":{"cohere-swift/legacy-constructors":2},"formatted":true}]}
```

The front door fills an `edit.Summary` from it and prints that summary's `String()`, then
`format scope: <formatScope>`. Fixes and formatting are separate counts, as they are for
TypeScript. Every file in scope is formatted the house way. A file the formatter could not format
(the formatter failed on it, or a file in scope does not parse, so nothing was rewritten) counts in
`filesNotFormatted` with its reason. A formatter that formatted nothing therefore never reads like
a formatter that found everything already correct.

`changedFiles` names every file the phase wrote, once each. Each entry carries `fixedBy`, the fixes that
landed in it counted by the rule that proposed them (empty when only the formatter changed it), and
`formatted`, whether the formatter changed it. Its length is `filesRewritten`. A run that writes nothing,
under `--no-fix` or with nothing to change, writes it empty. The front door reads it with #ytqqv8v's
output: it lists these files and counts them as cohered, and refuses a stream whose list length differs
from `filesRewritten`. Until then it ignores the field, as below.

The field is additive within contract 3 (@system_cohere_build's ruling on #s4k5yb9). A field this
document does not list is ignored by the front door, so an older front door loses only the list, never a
count or a verdict. The counts stay what the verdict reads, and the list never changes an exit code. An
engine from before the field writes no `changedFiles` at all. The front door reads that as *not listed*,
never as *no files changed*, and renders the counts alone. `Clean.jsonl` is such a stream.

### `types`

```json
{"kind":"types","diagnostics":1,"files":149,"elapsedMilliseconds":2187,
 "filesWithoutRecord":[],"build":"swift build, scratch <root>/.cache/cohere/swift"}
```

Rendered as `types: 1 diagnostics over 149 files in 2.187s`. Each entry in `filesWithoutRecord`
prints `  types: no compiler record for <file>, so its diagnostics are unknown`, and any such file
makes the run incomplete. The types phase reads each file's serialized diagnostics (`.dia`), not
the build's stdout, because an incremental build does not reprint warnings in files it did not
recompile. So a file with no `.dia` is a file the engine cannot vouch for.

`build` says in words how the records were had: a build into the scratch, the last build's records
reused, or edited files checked in process. A run whose builds all succeeded may end it with what the
scratch prune removed, as `; the scratch prune removed 6.9 MB: 226 index records no unit names`. The
prune runs after such a build, at most once an hour per scratch. It removes the index store's records
no unit names and the scratch of a package that is no longer local, each unused for an hour, and never
touches a build in flight or the build descriptions, whose cache the build system bounds itself.

**Only compiler errors cut lint off.** In TypeScript every diagnostic the checker reports is an
error, so "any type diagnostic bails" and "any type error bails" were the same rule. In Swift they
are not. A warning (`variable was never mutated`) is a finding and fails the run, but the code
still means what it says, so lint runs and its findings are not noise. An error means the semantics
are wrong, so lint is recorded `notReached` with the detail
`types bailed: N type errors — lint findings against wrong semantics are noise`.

The types phase always covers the whole package, even on a scoped run. In Swift a change in one
file changes what other files in the module mean, and the compiler checks the whole module
regardless. Narrowing what is reported to the scope would hide findings the edit caused. The
front door prints `  types: covered the whole package, not only the N files in scope` when the two
differ.

### `lint`

```json
{"kind":"lint","findings":3,"rulesRun":15,"filesWalked":149,"nodesVisited":412233,"elapsedMilliseconds":96,
 "reusedFrom":"","rulesSilent":["cohere-swift/consistency-no-print"],"rulesWatchedAndQuiet":9,
 "crashes":[{"file":"/…/X.swift","error":"…"}],
 "rulesScopedOff":{"cohere-swift/consistency-no-print":12},"rulesNotConfigured":[],"configNote":""}
```

Rendered by the existing lint line and coverage printers:
`lint: 3 findings — 15 rules over 149 files, 412233 nodes visited, in 96ms`, followed by the rule
coverage, crash, and config notes. `reusedFrom` names the phase whose walk was reused (`fix`), and
the line then says so instead of printing a duration.

### `phase`

```json
{"kind":"phase","name":"types","outcome":"notReached","elapsedMilliseconds":0,"detail":"1 type diagnostics — lint findings against wrong semantics are noise"}
```

`name` is `fix`, `types`, `lint`, or `unused`. `outcome` is `ran`, `skipped`, `notReached` or
`reused`; these are the phases.go outcomes, spelled in camelCase. The engine writes exactly one
`phase` record per phase, every run, in pipeline order, including phases that did not run. A
missing phase record is a protocol error, for the reason `markRemainingNotReached` gives: a phase
absent from the report reads the same as one the reporter forgot.

A phase the caller did not ask for is `skipped` with the detail `not requested`, exactly, or for
unused `not requested — ` and how to ask for it. That detail is the only one that means nobody
asked, and it is not a gap: `--lint` asked for lint and got all of it. Any other `skipped` detail on
a phase that reports findings is a gap. The front door reads the command line too, the way the
TypeScript run does (no phase flag asks for fix, types and lint; `--fix`, `--types` or `--lint`
only for those named; any `--unused` flag for unused as well). So a phase it asked for that comes
back `not requested` is a protocol error, not a smaller run.

### `rule`, only under `--rules` and `--rules-enabled`

```json
{"kind":"rule","name":"cohere-swift/force-unwrapping","severity":"error"}
```

`--rules` prints names, sorted. `--rules-enabled` prints `name<TAB>severity`. The development-build
note goes to stderr, as it does for TypeScript.

### `summary`, always last

```json
{"kind":"summary","findings":4,"complete":true,"exitCode":1}
```

- `findings` is the sum of every `finding` record. The front door counts the findings it received
  and refuses a summary that disagrees. Two counts that should agree and do not are a defect.
- `complete` is false when any file, or any phase the caller asked for, that could have produced a
  finding did not run. A phase skipped as `not requested` is not one of them. The
  front door prints `this run did not check everything` from its own phase records anyway. A summary
  calling the run complete over a gap its records show (a requested phase not run, a file without a compiler
  record, a rule crash, an `unreadable` file) is a protocol error. The other direction is believed: an
  engine that says it fell short where no record shows it is taken at its word.
- `exitCode` is what the engine will exit with.

## Exit codes

| Engine exit | Meaning | Front door |
|---|---|---|
| 0 | summary written, `findings` 0, `complete` true | exits 0 |
| 1 | summary written, findings or an incomplete run | exits 1 |
| 2 | the engine could not run: bad flag, contract mismatch, no package, toolchain missing, refused config (a leftover `.swift-format`, a `format` key) | prints the engine's stderr, then `cohere: the Swift engine did not run, so nothing was checked`, exits 1 |
| other, or killed | crash | `cohere: the Swift engine exited <code> without a summary, so nothing was checked`, exits 1 |

An exit of 0 or 1 without a `summary` record is a crash. An exit code that disagrees with the
summary is a protocol error. The front door owns the final exit code, and there is no path through
it where a run that did not finish prints green.

## Flags

| Flag | Swift meaning in phase 1 |
|---|---|
| `--no-fix` | Writes nothing to the project: no source, no `.build`, no `Package.resolved`. Builds go to the engine's cache outside the project, and package resolution is disabled. The fix phase runs in check mode, recorded as `ran`: each file the formatter would rewrite becomes one `format` finding at its first changed line. |
| `--fix` | Fix and format only. |
| `--types`, `--lint` | That phase alone, as for TypeScript. |
| `--format` | Formatting is on by default for Swift, unlike TypeScript, where it waits on parity with the existing gate. For Swift that parity is measured: swift-format 604.0.0 as a library is byte-identical to `xcrun swift-format` on 952 of our files. The flag is accepted and changes nothing. |
| `--format-all` | Accepted and changes nothing. Swift formats every file in scope by default, because formatting all of Presence costs a few seconds, not the minutes that made TypeScript format only changed files. Named paths still narrow the scope. |
| `[paths]` | Narrow fix, format and lint to the scope. Types still reports the whole package (see `types`). |
| `--lint-config <file>` | The `CohereSettings.json` whose `swift` block configures rules. |
| `--fix-passes <n>`, `--single-threaded` | As for TypeScript. |
| `--rules`, `--rules-enabled`, `--version` | `rule` and `provenance` records. |
| `--unused`, `--unused-all`, `--unused-deep` | The `--unused` report: `unused` and `unusedCoverage` records after lint, never counted as findings (see those records). Today it holds `unused-import` and `unused-declaration`, each with its own `unusedCoverage` record. It reads the index the types phase writes, so on a run without types it reads the last build's and counts every file that build did not compile as it stands as not checked. A bail before it (a file that does not parse, a type error) records `unused` as `notReached`. `--unused-all` and `--unused-deep` ask for nothing more yet and run the same report. |
| `--timing`, `--explain <file>` | Not implemented for Swift yet. The engine refuses with exit 2 and names the flag. A flag that is accepted and ignored reads as a run that did what was asked. |
| `--tsconfig` | Meaningless for Swift. The front door refuses it against a Swift root. |
| `--directory` | Resolved by the front door into `--root`. |

## Config

Rules are configured in the `swift` block of `CohereSettings.json`. Keys are `cohere-swift/<rule>`,
and values are severities and strictness options, never allow lists or ignore names.

**Formatting is not configured.** Swift has one house format, built into the engine
(`Sources/CohereSwift/Format/HouseSwiftFormat.swift`, ruled 2026-10-03 in #mg4dgjm), the way
TypeScript's Prettier options live in one block of the Nexus tier. Every file the engine owns is
formatted with it, and nothing a repository carries changes it. Two things are refused by name, each
with exit 2 before anything is checked, as a bad flag is:

- A `.swift-format` at or above any file the engine owns, up to the repository root (the package
  itself outside git). The refusal names each file and says to delete it, and to bring any setting it
  needs to the house format instead. A `.swift-format` above the repository root belongs to someone
  else and is not looked at.
- A `format` key in the settings file, at the top level or in the `swift` block. The refusal names
  the file and the key and says to remove it, and to bring any setting it needs to the house format.

**Where the config lives.** The engine reads `CohereSettings.json` from beside `Package.swift`, or
the file `--lint-config` names. When there is none, every house rule runs at `error`, and every run
says so in one line naming the defaults in force and where a config would go:
`  config: no CohereSettings.json beside Package.swift, so every house rule ran at error — put one at <root>/CohereSettings.json to change that`.
That is not the permissive default the TypeScript loader refuses. The TypeScript loader refuses to
lint with nothing configured because the output would look like a clean run. Here, every rule runs,
and the line says why. Approved by @system_cohere, 2026-10-01.

The line travels in the `lint` record as `"configNote"`. It is empty when a config was read, and the
front door prints it under the lint line whenever it is not empty.
