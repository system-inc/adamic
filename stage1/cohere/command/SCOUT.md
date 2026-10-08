# Step 42: checked settings and native formatter wiring; real Path I/O awaits the host API

Task **#2eq9rwz**, current branch **scout/42-native-settings**, parent
**f2a5a82377a933d18a674d949c4fbee4411de485** (`scout/42-settings-resolution`).
The first linked command was **1ad13f7cc64b78598e2e677b89238568083b53eb**.
The original lint-area base was **9156bf5c579a44d687c9955d13e44f9ad8bbb6f8**.
Go oracle: cohere **7945d102a6c18dd36adf9114a758ce646e8b2359**;
its TypeScript checkout is **d92d9bfee114c80be2c375d72edae966176e3a4f**.

This package owns `main.a`, `format.a`, `settings/`, its tests, fixtures and evidence only.
No shared lint harness, parser, bridge, runtime or compiler file is changed.
The four original host questions are resolved by the user-supplied rulings below.
The integer representation question is resolved by the later ruling below.
Historical findings retain their original scope; the latest follow-up is at the end.

## Formatter composition inherited from the parent scout

`main.a` accepts cohere's `--format-only`, `--format-all`, `--format`,
`--no-cache`, `--directory` and explicit file arguments. It formats files at
Prettier defaults, in place, through JSON, CSS, YAML, GraphQL and the existing
partial TypeScript printer. A single source graph is emitted to C, compiled
into an object, archived as `cohere.a`, and linked with the existing runtime
archive. Adamic `.a` source and C static `.a` archives are distinct artifacts;
this experiment exercises both, without introducing a module ABI.

This is a formatting composition, not a complete replacement command.
It has no repository settings resolver, tree walk, format record, fix/type/lint
pipeline, JSON finding report, command summary, cache replay or rename verb.
It accepts a cold explicit-file run at fixed Prettier defaults. Its CLI comparison
uses an explicit Nexus `format: {}` configuration; zero-config Go cohere uses
the house defaults (width 120, indent 4, single quotes), and is not this scope. `--no-fix`
and other unsupported flags fail explicitly. Error exits remain Adamic panic
70, not cohere's 1. Successful formatting bytes are the parity claim; stdout,
stderr and exit behavior of the full CLI are separate outstanding contracts.

## Source evidence and shortest boundaries

| Boundary | Go cohere | Port | Shortest useful input / consequence |
| --- | --- | --- | --- |
| Command flags and phase selection | `cohere/command/cohere/main.go:66`, `:85`, `:95`, `:160`, `:371`, `:421`, `:443`, `:544` | `main.a:4`, `:8`, `:16` | `--format-only --no-fix x.json`, with text `0`: Go reports drift, exits 1, writes nothing. The scout cannot implement that exit through today's runtime. |
| Normalization shared by all file printers | `cohere/internal/format/native/native.go:212`, `:225`, `:239`, `:255` | `format.a:9`, `:10`, `:38` | BOM + `0`, named `.json`; CR after `0`. Preserve BOM, normalize CR/CRLF, preserve exact final output bytes. |
| Filename-dependent JSON parser | `cohere/internal/format/native/json.go:11`; `cohere/internal/format/javascript/printer.go:66`, `:73` | `stage1/cohere/json/formatter.ts:5`, `format.a:14` | `{"a":[]}` as `package.json` versus ordinary `.json`: different layouts. Passing a generic name loses the stringify route. |
| CSS file and embedding routes differ | `cohere/internal/format/native/css.go:15`, `:18`, `:21` | `format.a:17`, `stage1/cohere/css/print.ts:262` | `a{b:c}` exercises two-space indentation. `.scss` is an embedding parser in Go, not a standalone file printer. |
| YAML defaults and wrapper | `cohere/internal/format/native/yaml.go:16`, `:22`, `formatoptions/resolve.go:91` | `stage1/cohere/yaml/format.ts:8`, `format.a:22` | `a:1` or BOM-only `.yaml`; exact bytes, including empty bodies, belong to the file wrapper. |
| GraphQL defaults | `cohere/internal/format/native/graphql.go:13`, `:19` | `format.a:27` | `{a}`; width 80, indent 2, bracket spacing true. |
| Three-pass convergence, not one formatting call | `cohere/command/cohere/format.go:106`, `:152` | `format.a:40` | Every input is formatted until stable; still-changing third pass is a failure. No shortest nonconvergent source is claimed from this fixture set. |
| Partial TS files | `cohere/internal/format/native/typescript.go:19`, `:24`, `:26` | `stage1/cohere/tsprinter/files.ts:6`, `:7`, `:9`; `literals.ts:46` | `//`: Go formats it, source Node and native refuse `comment-attachment`. Numeric `1+2;` works as `.ts` and `.a`. Whole-file quotes/configuration are not certified. |
| Markdown lacks a composed file entry | `cohere/internal/format/native/markdown.go:28` | `stage1/cohere/markdownblocks/GAPS.md:3`, `:7`; `mdastCompile.ts:30`; `document.ts:16` | `a`, named `.md`: Go has a formatter; the command has no `.md` route. Do not substitute a leaf printer or unchanged input. |
| Imported package drivers execute I/O | their `main.ts` bodies | `json/main.ts:61`, `css/print_main.ts:57`, `yaml/main.ts`, `graphql/printer/main.ts`, `tsprinter/main.ts`, `lint/main.ts` | Import `../json/main.ts` with no arguments: it demands its own protocol and panics. Import APIs, not existing executable drivers. All driver state is listed in `testdata/entry-state.json`. |
| Whole lint graph needs the checker library even for a syntax-only request | `cohere/internal/lint/rules/core/no_debugger_test.go:18` proves `debugger;` fires | `lint/lint.ts:59`, `lint/checker_bridge.a:6`; stage 0 `internal/lower/tsgo.go:27` | The retained 15-line whole probe selects only `no-debugger`, reports one finding on Node, then formats `{}`. Stage 0 refuses an unlinked library call and explicitly asks for `--tsgo <checker archive>`. |

Go's file formatter is held independently of the port by the overlay adapter in
`testdata/oracle.go`, using `native.Formatter` and `PrettierDefaults`.
The real Go command is additionally built and run with its own flags on the
same source files, with an explicit empty Nexus format block. Zero-config Go
uses the house format: `formatoptions/resolve.go:179` and `:182`, not
`PrettierDefaults()`. Inputs are never removed after oracle or parity failures.

## Every name, helper and state site

The complete machine-readable inventories are part of this package; the counts
below refer to actual source graphs, not imagined linker symbols.

* Formatter command: **97 modules**, **50 repeated top-level names at 130 sites**,
  **83 top-level const/let declarations**; `testdata/inventory.json` includes every
  location and declaration. The repeated names include `format`, `Parser`,
  `Printer`, `Documents`, `SettingsOptions`, `ResultType`, `stringWidth`,
  `numberText`, `stringText`, `parse`, `tokenize`, `written` and `defaults`.
* Formatter + registered lint scratch graph: **371 modules**, **93 descriptors**,
  **82 repeated names at 407 sites**, **248 top-level const/let declarations**;
  `testdata/whole-inventory.json`. Factories use the existing registry's named
  aliases; listener classes do not require a global source-name namespace.
* Byte-identical modules: JSON/TS `width.ts` (two files), and JSON/TS/YAML
  `widthTables.ts` (three files). These are **two groups / five files**. Other
  repeated names are candidates, not assertions that their algorithms are the
  same. In particular, the separate document engines are different contracts.
* Go command + all formatters + lint: **1,905 package variable declarations**,
  **170 repeated free-function names**, **33 text-identical free-function groups**.
  `testdata/go-inventory.json` is built from Go ASTs and records every variable,
  function location and declaration. Constants are not called mutable state.
* Eight standalone drivers have **39 top-level const/let declarations** in
  `testdata/entry-state.json`; importing those drivers would start multiple CLIs.

State that matters to composition: CSS has two module regex objects at
`css/print_doc.ts:5` and `:6`; the global regex's `lastIndex` is reset at `:20`.
YAML shares compiled schemas at `yaml/composer.ts:15` and `:16` and a width
lookup at `yaml/width.ts:102`. TS shares an operator map at
`tsprinter/syntax.ts:24`; scanner keyword and punctuation maps are at
`typescript/scanner/tokens.ts:2` and `:91`. The JSON/GraphQL/TS/YAML width tables
are module constants. Per-file parser, tree and document state is created by
its existing API and remains owned by that call. The command owns arguments,
its file list and two mutable flag locals; none points back into a parser.

On the Go side, native's printer, parsed-printer, embedded-doc and raw-doc maps
and mutex are `native/native.go:48`; parser-to-name routing maps are `:82` and
`:100`. Extension duplication is a startup panic (`:165`), not last-writer-wins.
Command cache flags, process start, run-cache state, format walk-ahead, profiles
and reporting modes are all retained in the Go AST inventory. They are not
ported by this formatter-only entry.

Independently emitted executable objects each define **`main`**. The package
contains a real failed two-object link proving that clash. Source names from
separate modules stay distinct in one emitted program; they must not be fixed
by concatenating source or renaming another lane's helpers. The whole-program
archive here is one C translation unit, not separately compiled reusable
Adamic package archives. No stable cross-package ABI is claimed.

The TS AST helper inventory (`testdata/ts-helper-inventory.json`) additionally
records **980 callable sites**, **11 raw-identical declaration groups** and
**17 raw-identical body groups**, including class methods. Matching bodies do
not settle imported bindings, parameters or nominal class identity; they are
exact duplication candidates, not permission to consolidate another lane.

## Fixtures and checks

The quiet hundred lives on Kirk's Mac. The user supplied 23 public repository
commits; **all 23 were shallow-fetched at those exact SHAs**, with no checkout,
dependency installation or execution of repository scripts. `testdata/corpus.json`
records repository, commit, original path, saved fixture path and SHA-256.
Normal verification uses the saved bytes offline.

The selected real sources are **23 root package.json files**, **six CSS files**
(the first two tracked CSS paths in Actual, TanStack Query and tRPC), and **three
TypeScript 6.0.3 compiler cases**: `2dArrays.ts`, `APILibCheck.ts`,
`APISample_Watch.ts`. These are a sample from the supplied quiet-hundred pins,
not all files or all 100 repositories. The TypeScript cases are held as explicit
whole-file refusal proofs against Go, Node and native; they are not credited as
successful formats. Their source pragmas/comments prevent the partial TS entry
from claiming arbitrary whole-file coverage.

Go cohere's own `native/core_test.go:19` supplies six unchanged source texts:
ordinary JSON, package JSON, CSS, GraphQL, YAML and YML. Each is tested raw,
with BOM, with CRLF, and with BOM+CR: **24 cases**. Six short controls add ordinary
JSON, CSS, YAML, GraphQL, numeric TS and numeric Adamic: **30 controls total**.
Together with the **29 quiet JSON/CSS files**, that is **59 successful file-byte
comparisons** on the Go API, source Node, release native and sanitized native.
The real Go CLI produces **54 matching formatted outputs**, including **26 quiet
files**, and **five exact pinned format/unchecked findings** with unchanged inputs.
Backstage, Excalidraw and shadcn-ui package files retain a legacy `prettier` key;
two BOM-bearing package controls are invalid as settings-discovery JSON. The
command does not resolve or reproduce those configuration refusals; they are
explicit boundaries in `testdata/cli-refusals.json`, not successful CLI formats.

Three compiler source files plus `//` form **four explicit
unsupported TS proofs**. Go formats `2dArrays.ts` and `//`; it refuses the two
raw multi-file API tests at offsets 253 and 204 because their embedded JSON is
not one TS source file. `testdata/proof-go.json` pins every exact Go outcome.
`0` in a JSON file proves Go's no-fix exit 1 and unchanged
source. The archive link and the intentional duplicate-main link are real builds.

Three owned-code mutants must build, exit successfully and differ on file bytes
on both native and source Node:

1. Drop the BOM restoration (`format.a`).
2. Always choose an ordinary JSON filename (`format.a`).
3. Change CSS indentation from two spaces to four (`format.a`).

No test uses `t.Skip`; the corpus and its pins are mandatory, and the only new
top-level Go test calls `t.Parallel`. The artifact environment variable merely
retains the release files; it does not gate coverage.

## Design within today's rules

Keep one `.a` entry module and named imports from pure package APIs. Keep routing
and the file wrapper in this package. Use readonly discriminated results and
readonly option objects. Mutation is confined to arguments, current text and
existing parser/document arenas. Existing numeric tree/doc indexes keep ownership
acyclic; module initialization follows the language's dependency order. There is
no dynamic module loader, source concatenation, mutable global formatter registry,
new syntax or collector in this port.

Go command behavior must be ported as behavior: resolve options per file, create
one lint context per file, converge fixes before formatting, offer the already
parsed tree where supported, keep refusal distinct from unchanged output, and
collect reporting/exit status after processing all files. The current command
cannot claim these phases by silently returning the input or by pretending panic
70 is cohere's exit 1. Existing checker linking uses its approved bridge and
archive; its Go runtime is an external dependency, not a collector added to Adamic.

The first piece needs no ruling because it invokes already supported APIs and
links a single whole program. Remaining language/library changes stop at the
shared-file boundary. Formatting-option completeness belongs to the printer
lanes; the command should pass a resolved option object once those APIs accept it.
A complete Markdown file adapter likewise awaits its owning lane.

## Historical host questions from the parent scout (resolved below)

1. What supported host API should let a command finish with cohere's exit **1**
   or flag-parser exit **2**, while preserving cleanup and flushed output?
   `panic` is 70; neither source Node's Adamic runtime nor the prelude exposes a
   general command exit. Relevant shared files: `internal/load/prelude.d.ts`,
   `oracle/adamic.mjs`, and the runtime's process/cleanup implementation. No
   process API or fake panic mapping is introduced here.
2. What supported stdout/stderr byte-writing API should reproduce cohere's
   diagnostics and summaries, including empty strings and strings without final
   newlines? `console.log` is a one-string line operation; opening `/dev/stdout`
   or `/dev/stderr` is not proposed as a portable language answer. The same
   prelude/oracle/runtime boundary needs a ruling.
3. For exact Go file/path behavior outside this UTF-8 sample, what byte-I/O
   representation is supported? JavaScript/Adamic strings expose UTF-16, while
   Go can retain non-UTF-8 file bytes and names. No encoding substitution is
   blessed by this scout; expanding the command corpus requires the host-I/O
   contract in those shared files.
4. If this step means **separately compiled reusable Adamic package archives**,
   rather than one program whose entry is `.a`, what is their approved symbol,
   class identity, module-initialization and ownership ABI? Today's program-local
   emission plus a whole-program archive does not settle that language/toolchain
   question. The compiler/native lane would own the implementation.

## Validation and next scout

Measured Linux amd64 release builds:

| Program | Generated C bytes | Program/archive bytes | Binary bytes | Final link |
| --- | ---: | ---: | ---: | ---: |
| 97-module formatter command | 14,432,303 | 8,307,758 (program object archive) | 4,384,408 | 0.202 s in the final verification run |
| 371-module formatter + 93-rule lint probe | 21,578,328 | 12,891,950 (50 C objects, program + runtime) | 43,427,320 | 0.770 s |

The second program also links the unchanged **49,291,344-byte Go checker
archive**. No live checker call is made by the syntax-only debugger control;
this is a link/composition proof, not new type-checker parity credit. Node and
native both print `1\n{}\n\n`; Go's own `TestNoDebuggerFires` holds the one finding.
`testdata/whole.sh` reproduces the scratch graph and archive measurement.
The 50-object build preserves stage 0's C flags and uses whole-archive for those
objects, preserving the original monolithic link's runtime inclusion. Its binary
size and output match the ordinary `--tsgo` build. C compilation and final link
are timed separately, not presented as one link measurement.

The only observed stage-0 refusal in this composition is the unlinked checker
call at `checker_bridge.a:6:33`; the existing `--tsgo` path closes it. There is
no new syntax refusal, source-name clash or compiler workaround in the composed
native program. The independent-entry experiment fails on the exported `main`.

The package test passed with no skips: all 59 formatter cases matched on source
Node, release native and sanitized native; the real CLI held 54 exact outputs
and five exact refusals. The BOM, JSON-filename and CSS-indentation mutants
changed 12, 19 and eight cases respectively on both Node and native. `go vet`
and `gofmt -l` are clean. Complete outcomes are retained in
`evidence/package-test.log`, `evidence/vet.log` and `evidence/gofmt.log`.
Reproduce with README.md; inventory tools contain their complete selection rules. No shared source needs a change to reproduce the implemented
piece.

Take the explicit-file composition and mandatory real-source corpus first.
Implement settings resolution before widening CLI parity: the zero-config house
format, legacy Prettier refusal and BOM package-discovery boundary are measured
dependencies, not defaults the command may silently choose.
Then carry cohere's command reporting and `--no-fix` semantics after the host
ruling; wire settings/walk and the existing lint/checker lifecycle through owned
adapters. Coordinate the TS and Markdown completeness dependencies with those
lanes. Do not edit shared harness/parser/bridge/compiler/runtime to make this
package green. If their API contract needs changing, identify the exact file
and stop at that boundary.


## Settings follow-up: implemented scope and source evidence

`settings/resolve.a` discovers from each file's directory upward, stopping at the
nearest `CohereSettings.json`. At each directory it reads `package.json` first,
then checks all **19** legacy config names in Go's order, before accepting settings.
Unreadable package manifests are ignored as Go ignores them; successfully read
malformed manifests are refused. A legacy key is refused even when null or false.
No repository config or downloaded JavaScript is executed.

The resolver reads and validates the complete `extends` graph, detects cycles,
keeps a shared base at its first position and chooses the most derived format
block over Prettier defaults. A block replaces its base's block; its missing keys
do **not** inherit the base's options. No block means the embedded house format.
Chains reaching `cohere:system-inc/` may carry a block only at `cohere:typescript`.
The two ignore lists and their provenance/declaration bit travel with the options.
Directory answers, refusals included, are cached for one resolver's lifetime.
New resolver instances see changed snapshots; no process-global cache is added.

Per-file formatter overrides here mean the nearest configuration and the selected
chain block. Go's lint `overrides` have only `files`, `rules` and `reason`; they
are not formatter option overrides. The raw shapes and unknown keys are validated
as `SourcesOf` validates them. This component does not resolve effective lint
rules, selectors, severity, version pins, reasons or command `output` settings;
those belong to the subsequent full configuration/command integration.

| Behavior | Go evidence | Owned port evidence |
| --- | --- | --- |
| Upward search, package-first legacy refusal | `cohere/internal/format/formatoptions/resolve.go:138`, `:165`, `:402` | `settings/resolve.a:262`, `:282` |
| House versus own block, whole-block precedence | `cohere/internal/format/formatoptions/resolve.go:278`, `:330`, `:356` | `settings/resolve.a:220`, `:228` |
| Source graph order, deduplication, cycle/error wrapping | `cohere/internal/lint/configuration/configuration.go:623`, `:670`, `:719` | `settings/resolve.a:150` |
| Unknown top-level/override keys and typed raw fields | `cohere/internal/lint/configuration/configuration.go:964`, `:1036`, `:1096` | `settings/resolve.a:60`, `:121` |
| Sorted option checks, null/default handling, LF-only refusal | `cohere/internal/format/formatoptions/resolve.go:464`, `:483`, `:495` | `settings/resolve.a:187` |
| Tool error status | `cohere/command/cohere/main.go:44`, `:49` | `settings/host.a:8`, `:13`: refusal status is returned, not an abrupt exit |

`settings/json.a` is a strict data reader for Go's `encoding/json` grammar,
not the JSON formatter's JSON5 grammar. It retains raw numeric spellings and
numeric tree indexes, decodes duplicate keys as the relevant Go target does,
normalizes invalid UTF-16 surrogates, and reproduces the pinned **Go 1.27**
syntax/type diagnostics. Its `json.RawMessage` alias appears as `jsontext.Value`
in Go's error strings. Key sorting uses UTF-8 byte order, rather than JavaScript's
UTF-16 default sort; quoting uses a generated snapshot of Go's `unicode.IsPrint`.
All options/results exposed by the resolver are readonly; local construction
uses mutable arrays/maps with no ownership cycles or collector.

## Host rulings supplied by the user for @system_adamic

These replace the four questions above, for `.a` programs on the `adamic` module;
the tsc Node surface remains step 29's ruling.

1. **Main's return value is the exit code after normal teardown.** Cohere uses
   that route for 1 (findings) and 2 (usage), preserving leak checks. Explicit
   `exit(code)` is abrupt like Node's `process.exit`: flush pending output, run
   no cleanup, and count the leak check as intentional.
2. **Exact writes accept `Uint8Array` or a string.** Bytes go to stdout/stderr
   byte for byte; strings go as exact UTF-8 with no added newline or re-encoding.
   `console.log` retains JavaScript's added newline.
3. **Non-UTF-8 paths are branded `Path` byte sequences.** Every `adamic` file call
   must round-trip them losslessly. String conversion is explicit and checked
   for non-UTF-8 paths unless the caller explicitly requests a named lossy form.
4. **No reusable package archives or package binary ABI for now.** Compile the
   entire program closed-world for monomorphization/soundness; speed comes from
   build caching. Module initialization follows ECMAScript module order. Revisit
   reusable archives only if build time becomes the long pole.

The implementation represents a refusal as an immutable message, status **1**
and legacy-refusal flag. `SettingsHost` supplies reads and existence checks;
`CommandOutput` supplies string writers. `reportRefusal` returns the status and
passes an exact captured string to the adapter. The fixture adapter is a read-only
filesystem snapshot, and its protocol uses `console.log` only for test transport.
It does **not** claim the operating-system exit or final cohere CLI report stream.
A complete CLI must select Go's fatal-error or finding reporter at its phase.
There is no panic-to-exit mapping, `/dev/stdout` trick or fake branded Path.

The current shared prelude still lacks the ruled APIs, and native emission
still ends the module body with `return 0`. Required owning-lane files are
`internal/load/prelude.d.ts`, `internal/lower/input.go`, `internal/native/emit.go:76`,
`internal/native/runtime/` and `oracle/adamic.mjs`. They need the ruled main-return,
exact-write and branded-path contracts. UTF-8 fixture paths remain strings in the
snapshot interface; a future real adapter must use the approved Path API.
No shared file is edited here.

## Settings corpus and validation

`testdata/settings_sources.py` reads source blobs at the **same 23 exact public
commits** as the parent scout: every recognized root config and the first two
nested configs per pin. There were **42 recognized config paths**, of which
**20 files** were saved. All 23 root package manifests are also held. Missing
configs are recorded as absence, not skipped pins. `settings-corpus.json` records
all discovered paths and every selected file's origin and SHA-256; normal tests
use the saved bytes offline. Fetching runs no repository code and installs nothing.
The script also regenerates the exact **eight embedded Go sets** in `sets.a`.

`settings-cases.json` contains **249 control sets / 252 queries**, drawn from
Go's `formatoptions/resolve_test.go` and shorter shape/grammar reductions.
The 23 pinned trees add **31 queries**: **283 exact answer lines**, comprising
**77 successful resolutions**, **206 refusals**, and **47 legacy-config refusals**.
Go and source Node agree byte for byte on each entire answer, including options,
source, ignore lists, refusal message, intended status and captured strings.
No failing input is dropped from this comparison.

Four source settings mutants are caught: replace the house answer with Prettier
(**61** changed answers), ignore legacy config (**47**), stop the upward walk
(**270**), and invert chain precedence (**one**). Native settings mutants are not
credited: the native resolver is blocked as described below.

The independent strict reader compares **311 inputs** to Go on source Node,
release native and sanitized native, all exact. Two successfully built reader
mutants (accept BOM as whitespace; ignore trailing input) disagree with Go on
both native and Node. Its release binary is **467,440 bytes**, program archive
**249,774 bytes**, generated C **403,973 bytes**; final link **80 ms** in the final
run. The report-interface probe matches four captured strings/statuses on Node;
its native build has the independent refusal below.

Go's complete `formatoptions` settings suite passes with
`COHERE_RESOLVE_TREES` set to the entire saved fixture tree: **588 directories
identical, 429 refusals**, no skips. All four new top-level tests call `t.Parallel`.
The targeted settings suite passes, and gofmt/vet are clean. The unchanged
formatter command's complete parent-scout suite and evidence remain on the branch;
this follow-up does not claim a new full linked-command test run.

## Historical settings boundaries (superseded where stated below)

1. **Native settings composition refuses a call spread.** The actual entry is
   `settings/main.a`; stage 0 reports `settings/resolve.a:251:55`:
   `stage 0 can't lower a SpreadElement yet`, for
   `patterns.push(...layer.ignores)`. Valid call spread is in `docs/0.1.md`.
   `internal/lower/object.go:1131` currently accepts only a single push value,
   while its argument lowering reaches an unsupported SpreadElement first.
   The compiler lane must implement the documented call form. No source loop,
   concatenation substitution or generated flattening works around this gap.
2. **Native reporting probe refuses JSON literal shorthand.**
   `settings/report_main.a:12:34` reports
   `stage 0 can't lower JSON.stringify a literal with spread, shorthand or methods yet`.
   `internal/lower/library_json_stringify.go:113` admits only PropertyAssignment.
   The compiler lane must handle the valid shorthand. The source remains the
   failing proving program; no expanded property rewrite is substituted.
3. **Go int64 options exceed Adamic's exact number domain.** Shortest retained
   input: `{"format":{"tabWidth":9007199254740993}}`. Go accepts and resolves
   **9007199254740993**; source Node resolves **9007199254740992**. The separate
   boundary test requires this mismatch and grants no parity credit. Lexical
   validation handles out-of-int64 refusals exactly, but it cannot make accepted
   int64 values become exact float64 numbers.
4. **JSON and YAML option plumbing is owned elsewhere.**
   `stage1/cohere/json/formatter.ts:474` and `stage1/cohere/yaml/format.ts:8` expose
   fixed-default entries. Their owners must accept resolved options through the
   printer/document contracts before the linked command can apply house defaults
   or arbitrary own blocks. No output reindentation, text replacement or copied
   printer is added here. CSS already accepts its option object.

Shortest retained reductions are `settings/gaps/call_spread.a:3` (Node `x\n`)
and `settings/gaps/json_shorthand.a:2` (Node `{"status":1}\n`); both load as
valid TypeScript and retain their exact stage-0 refusal.

The first two are implementation gaps in existing valid syntax, not requests for
new syntax. The host rulings are settled policy whose shared implementation is
pending. The additional open language question for **@system_adamic** is:
**What supported exact representation should carry Go's accepted 64-bit integer
settings through Adamic APIs and reports, beyond float64's exact integer range?**
No BigInt, clamping, new refusal policy or string-number substitution is decided
by this scout.

Take the resolver, exact fixtures and retained boundary programs next. The compiler
lane must first close the two proven syntax gaps; the runtime/compiler lanes must
land the ruled host APIs; the numeric contract needs its ruling. Then add a real
Path-backed adapter and wire resolved options into formatter entries through their
owners. Hold the actual cohere CLI's complete reporter/exit behavior as well as
resolution before claiming command parity. Stop at each named shared-file boundary.


Final follow-up evidence: `evidence/settings-tests.log` contains the complete
passing targeted suite, upstream audit, mutants and boundary proofs;
`evidence/settings-go.jsonl` and `settings-node.jsonl` retain every identical
answer at the measured fixture paths. `settings-vet.log` and
`settings-gofmt.log` are empty successful checks.


## Latest ruling and implementation: checked integers, #hnm8t56

The user supplied @system_adamic's ruling: a setting is a JSON number and an
Adamic number is a JavaScript number. Beyond the safe integer range it is a
checked **settings error naming the key and original value**, never silent
rounding and never a bigint or a string-valued option. Go cohere will adopt a
refusal too under **#hnm8t56**, already with @system_cohere according to the user.
No duplicate message or issue was sent.

`settings/resolve.a` checks the raw integer spelling after Go's integer/type
validation and **before** Number conversion. The implemented safe-integer range
is **[-9007199254740991, 9007199254740991]**, following the ruling's explicit
JavaScript safe-integer wording: exactly ±2^53 is also refused. The source token
is retained only by the JSON reader and diagnostic, not carried as an option.
An error has this form:

```
/CohereSettings.json: format option "tabWidth" value 9007199254740993 is outside the safe integer range [-9007199254740991, 9007199254740991]
```

`testdata/settings-integer-boundary.json` stays the shortest named divergence:
`{"format":{"tabWidth":9007199254740993}}`. Its task/provenance are named at
`settings-integer-divergences.json`, the test and every divergent matrix row:
**#hnm8t56**. The pinned Go oracle still accepts exactly 9007199254740993; Adamic
now refuses with key/value and intended status 1. The test requires that named
outcome, so a Go pin that changes it forces reconciliation rather than silently
retiring the divergence.

The **36-row** integer matrix holds both printWidth and tabWidth at zero, signs,
safe bounds, ±2^53, the first rounding case, int64 bounds, out-of-int64 values,
fractions, exponent spellings and null defaults. Source Node, release native and
sanitized native match every ruled result. **12 current Go/Adamic differences**
are named #hnm8t56; none is counted as oracle parity. Removing the check builds
and finishes with wrong output on both Node and native, proving that silent
rounding cannot pass. No Go cohere source or submodule pin is changed.

## Filed compiler gaps and expressly authorized workarounds

The user reports the call-spread gap filed as **#aqh658t** and JSON literal
shorthand as **#kqxkvg0** in compiler's tree, and expressly authorizes their
recorded workarounds so step 42 keeps moving. That instruction supersedes the
previous prohibition for these two cases only.

The unchanged shortest proving programs now live at:

* `internal/lower/testdata/step42-command-settings/call_spread.a` (#aqh658t).
* `internal/lower/testdata/step42-command-settings/json_shorthand.a` (#kqxkvg0).

Their selector tests run source Node, require valid TypeScript, and require the
specific stage-0 refusal. **They fail when stage 0 lowers either program**, so
landing a compiler fix forces removal/review of the workaround. The compiler
folder README records the issue IDs, scope, output and reproduction command.
Only new proof files/documentation in that folder are added; no shared compiler
implementation is modified.

The inherited-ignore site uses **single pushes**, with an #aqh658t comment.
The test report JSON literals write **status: status**, with #kqxkvg0 comments
at every changed site (`settings/main.a`, `properties_main.a`, `report_main.a`).
The report interface now agrees on Node and sanitized native for all four
captured strings and returned status 1, without inventing an OS exit adapter.

Native compilation then reveals a third refusal: the original full-answer
transport's `JSON.stringify(answer)` at `settings/main.a:31:41` is refused:
`JSON.stringify object references (structural types can hide fields and toJSON;
runtime shapes need complete value metadata)`. The shortest reduction is
`internal/lower/testdata/step42-command-settings/json_reference.a`.
`internal/lower/library_json_stringify.go` owns that implementation. This gap
has no supplied task ID, and **no production workaround is deployed for it**.
The original full-answer program and its refusal remain held.

An independent scalar property probe verifies the resolver's values. For the
fixture arrays it emits NUL-joined contents plus counts, which are compared to
Go's fixture projection; this is a test protocol, not a general lossless array
serializer or a replacement CLI JSON reporter. All **283** property outcomes
agree on Go, Node, release native and sanitized native. Four valid-TypeScript
settings mutants disagree on **72, 47, 270 and one** outcomes on both Node and
native. The original Go/Node whole-answer comparison remains exact as well.

## Resolved options linked to formatter APIs

`format_resolved.a` takes an immutable settings answer. It propagates settings
refusals, normalizes file text/BOM as the original wrapper does, and converges
up to three passes. CSS receives width, indentation, tabs, quote preference and
trailing comma through its existing PrintOptions; GraphQL receives width,
indentation, tabs and bracket spacing. Complete Prettier-default answers may
also use the inherited JSON/YAML/partial-TS routes. A nondefault answer for
those routes returns **NotYet** explicitly; it is not a Go configuration error.
No output reindentation, copied printer or silent option dropping is introduced.

The new API is a linked composition ready for a host adapter, not a claim that
the original command flags now resolve real filesystem settings. The original
`main.a` CLI remains its fixed-default slice until real Path I/O and the ruled
main-return/exact-write APIs are available.

The linked native archive passes **47 exact Go/Node/native outcomes**: six
pinned quiet-hundred CSS files across house/default/own options, 25 GraphQL
option/normalization controls, default JSON/YAML/numeric-TS controls and an exact
configuration refusal. Three separate JSON/YAML/TS nondefault probes agree on
an explicit NotYet owner boundary. The whole program still links all inherited
formatters with the resolver; no reusable package ABI is introduced.

Measured release composition: generated C **15,411,307 bytes**, program archive
**8,790,766 bytes**, binary **4,640,496 bytes**, final link **175 ms**. The native
settings property binary is **628,784 bytes**, with a **502,654-byte** program
archive. The linked graph inventory (`testdata/native-settings-inventory.json`) covers
**103 modules**, **55 repeated names at 141 sites**, **86 module-level
const/let declarations**, and the same two identical-file groups. One closed
program/one main linked without a new symbol clash. No formatter/parser/bridge/
runtime/compiler implementation owned by another lane was changed. Existing formatter implementation and its original tests are
unchanged; the added linked bridge has its own exact native test.

## Real Path adapter: concrete shared-lane blocker

The ruled branded Path API still does not exist in this checkout. The minimal
fixture `testdata/host_path.a.txt` imports only the approved type and fails:

```
error TS2305: Module '"adamic"' has no exported member 'Path'.
```

The Path selector fails as soon as that export appears, forcing adapter work.
Read-only checks of **origin/main 45487a809f89885a3fc651cd590e7dabf31362dc**
and **area/runtime 9bc8201f1c97c4c18a03d263a40ece933ffa752e** found no Path export
in their preludes either. Neither branch was merged.

A real adapter requires the approved Path type, lossless path operations and
Path-aware file calls in `internal/load/prelude.d.ts`, lowering in
`internal/lower/input.go` and the owned runtime/oracle implementations. The
ruled exact writes and main-return status likewise still need
`internal/native/emit.go` and `oracle/adamic.mjs` support. Adding a string brand
here would lose non-UTF-8 paths and violate the ruling. Therefore **no purported
Path-backed adapter is added**, and full CLI settings/refusal exit parity is
unfinished at this named shared-file boundary.

Take the checked resolver and linked formatter API next. Land the actual host
APIs, implement the real adapter, and wire the original CLI to this API. Let
JSON/YAML/TS owners expose their remaining resolved-option contracts. Track
#hnm8t56 at the oracle pin, and retire #aqh658t/#kqxkvg0 workarounds when their
selectors fail. The full-object serializer refusal must also be addressed before
claiming complete command JSON-report parity.


Latest evidence: `evidence/native-settings-tests.log` records the passing
settings/reader/report/selector/integer suite and all native mutants;
`evidence/resolved-format-tests.log` records the linked archive test;
`evidence/path-selector.log` and `host-path.log` retain the missing-Path proof.
`native-settings-vet.log` and `native-settings-gofmt.log` are empty successful
checks. These are targeted affected-package checks; the unchanged original
fixed-default CLI's parent validation remains preserved, rather than claimed as
a newly repeated full CLI suite.
