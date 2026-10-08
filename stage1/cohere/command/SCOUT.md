# Step 42: one entry can compose the printers; a cohere command still needs its reporting and host contract

Task **#2eq9rwz**, branch **scout/42-linked-command**, base
**9156bf5c579a44d687c9955d13e44f9ad8bbb6f8** (the fetched `area/stage1-lint`).
Go oracle: cohere **7945d102a6c18dd36adf9114a758ce646e8b2359**;
its TypeScript checkout is **d92d9bfee114c80be2c375d72edae966176e3a4f**.

This package owns `main.a`, `format.a`, its tests, fixtures and evidence only.
No shared lint harness, parser, bridge, runtime or compiler file is changed.
All language/API questions below are open for **@system_adamic**.

## What is implemented

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

## Open questions for @system_adamic

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
