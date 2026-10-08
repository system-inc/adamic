# Step 43 (#csz98j8): an executable byte-agreement scoreboard, not a parity claim

Branch: `scout/43-scoreboard`, based on `origin/area/stage1-lint` at
`9156bf5c`. Oracle: cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`;
its TypeScript submodule is `d92d9bfee114c80be2c375d72edae966176e3a4f`.
Only this package is changed. Generated, ignored lint registry files are produced
through the existing public `registry.Generate` API; no shared source is edited.

## Research and evidence

The quiet hundred is Kirk's Mac corpus, not a repository manifest. The user's
23 public snapshot SHAs were all shallow-fetched and checked at their exact
commits, without installs or source-repository scripts. `public-fetch.json`
records each actual identity. `corpus.py` inventories Git blobs, rather than
inventing a hundred-file substitute. Its first-run selection is explicitly one
shortest nonempty, non-declaration `.ts` file per snapshot, with lexical ties.
This is a probe run, not all files in those repositories and not the quiet hundred.
`first-run.manifest.sources.json` records the selected paths, hashes, sizes and
full supported-file denominator for each snapshot. `--all` produces the complete
tracked supported-file manifest for those same snapshots. The actual 242,883-file
input list is included as `public-all.manifest.json.gz`; it has not been scored
in full. `fetch-public.py` recreates the read-only source snapshots in scratch.

The driver runs the existing port on source Node and builds the actual Go
oracle with cohere-internal imports through an overlay. It compares **stdout
bytes**, including human descriptions, byte ranges, message ids, fixes and
suggestions. No whitespace, filenames, order or text is normalized away. In a
full registry run, per-rule slices retain their human finding blocks and all
following edit/suggestion records. An additional `lint/all-fixes` cell holds the
complete answer, including cross-rule order, overlap decisions and converged
fixed source. Per-rule cells alone cannot certify cross-rule fixes.

Relevant implementation sites:

| Boundary | Go cohere / independent oracle | Port |
| --- | --- | --- |
| public rule universe | `cohere/internal/lint/registry/registry.go:48`, `rule/register.go:183` | `stage1/cohere/lint/registry/registry.go:98` discovers owned descriptors; `:257` renders dispatch |
| exact human bytes and stable position order | `cohere/internal/lint/report/report.go:48`, `:62` | `stage1/cohere/lint/main.ts:60` UTF-16 to UTF-8 table and subsequent human/range serialization |
| actual diagnostic/edit protocol | `stage1/cohere/lint/testdata/oracle.go:75`, `:114`, `:173`, `:136` calls real rules, `report.Write` and edit engine | `stage1/cohere/lint/main.ts:13` runs the actual Linter and serializes findings |
| debugger witness and safe removal | `cohere/internal/lint/rules/core/no_debugger_test.go:18`, `no_debugger.go:19`, `:54` | `stage1/cohere/lint/rules/no-debugger/rule.ts:9` |
| required options versus inert rules | `cohere/internal/lint/rule/register.go:43`, `registry/registry.go:94` | each discovered `rules/<slug>/oracle.go`; manifest decoded JSON field 5 |
| checker availability | oracle `testdata/oracle.go:181` builds file context, requires program for typed rules | `stage1/cohere/lint/main.ts:54`, `:204` recording/replay; `oracle/adamic.mjs:137` live Node bridge unavailable |
| formatter wrapper semantics | `cohere/internal/format/native/native.go:203`, `:218` strips/restores BOM and normalizes CR | direct port formatter APIs; `scoreboard/format.a` carries their actual results |
| full TS program printing | actual native TypeScript printer | `stage1/cohere/tsprinter/expressions.ts:1814` `formatProgram`, with explicit unsupported-family result |

The independent catalog runs Go's complete registry, not the port list, once per
file/rule. It records findings and finding-bearing files for every unported rule,
sorted by findings descending then rule name. `status != complete` means a
**lower bound**, not a zero-finding rule: required options, missing checker,
parse diagnostics, decoder errors and rule panics remain named. Defaults for
path-sensitive decoders are anchored to each snapshot root. This is a default
rule census, not the project's resolved ESLint/cohere policy, suppression or
module-resolution environment. The registered port adapter options are decoded
internal structs; the global registry census options use its registered external
JSON decoders. They are distinct dialects. Non-default option ranking needs a
separate external-option manifest; this piece does not assert their equivalence.

Hard-case inputs and what is actually held:

* `debugger;\n`: a real Go test. Go and source Node must produce the exact
  finding and deletion/fixed answer. The clean-running Node mutant changes the
  human message, and must become `diverge`.
* `\ufeffvar x=10;\r\n`: verbatim TypeScript compiler `bom-utf8.ts`. This
  exposes byte/UTF-16 coordinates, BOM and CR handling, and a declaration-shaped
  full-program formatter input. The fixture-run report retains both answers.
* `skipped typed/rule no program\n`: shortest protocol witness for false green
  coverage. Equal missing-checker records must be `blocked`, not agreement.
* `refused format/typescript function-types\n`: equal formatter coverage
  refusals must also be blocked. Process failure and cancellation are held
  separately; two failed executions cannot certify bytes.
* A one-character change in a suggestion's replacement (`===` to `==`) must
  survive per-rule slicing and be detected. Another rule's bytes stay identical.

The relevant READMEs/GAPS were read before choosing these boundaries: lint,
parser, TypeScript printer, JSON, YAML and formatfiles, plus `docs/0.1.md`,
`docs/memory.md`, `docs/lint-registration.md` and stage1 progress guidance.
The historical lint parser-EOF report is not assumed current: this baseline's
`parser.ts:989` already uses list terminators/recovery. No parser repair is made.
At this pin, the compiler corpus is actually under
`cohere/TypeScript/tsc/testdata/tests/cases`, not `cohere/TypeScript/tests/cases`.

### First measured scoreboard

`first-run.json`: **23 files / 642 source bytes**, from a **242,883-file**
tracked supported-source denominator across the 23 snapshots. **93 registered
port rules**, of which **3 require a checker**. **2 lint findings** match Go and
Node. **2,185 cells: 2,072 agree, 1 diverges, 112 are blocked**. The 112 are
69 typed-rule cells, 23 all-fixes cells lacking complete typed coverage, and
20 formatter refusals. Exactly one observed public formatter byte divergence:
Prisma's `packages/1-framework/3-tooling/cli/test/fixtures/empty-module.ts`
loses its comment. Its parse-valid deletion-minimal input is **`//`**.
Go formats the comment; the program port emits empty text. The cause is the
unguarded whole-program printer path at `tsprinter/expressions.ts:1814`, which
has no comment attachment. The expression entry's comment refusal at `:1786`
does not protect callers of this separate exported entry. No shared printer fix
is attempted.

`fixture-run.json`: **2 files, 6 cells: 5 agree, 1 diverges**. The real Go debugger
fixture produces **1 identical finding** and deletion/fixed-source answer.
The TypeScript BOM/CRLF fixture exposes a missing-BOM formatter result. Its
parse-valid deletion-minimal input is the single **U+FEFF** character. Go's file
wrapper preserves it; the port's whole-program printer loses it. Both reduced
engine answers are retained in the reports, separately from the original answers.

The full Go registry has **492 rules**, so **399 are unported** on this branch.
Of those, **178 have a complete default probe census** and **221 are incomplete**:
194 lack a program, 24 lack decoded options, and 3 explicitly require options.
The three nonzero unported probe counts are one each for
`@typescript-eslint/no-non-null-assertion`, `id-length`, and
`structure/consistency-require-organized-imports`. This sample is too narrow to
rank work on the quiet hundred; the report does not present these counts as that
ranking.

## Design and limits

The scoreboard is host Go orchestration, like existing oracle/test tooling; it
contains no replacement lint rule. The owned formatter adapter is valid
TypeScript/Adamic, imports concrete existing ports and uses readonly existing
option interfaces. There is no new syntax, runtime service, GC, compiler change,
or substitute implementation for a language gap. Reports are append-only values
until aggregation; source inputs are read-only, and mutant/reduction files live
in scratch. All new top-level Go tests call `t.Parallel`.

`agree`, `diverge` and `blocked` are separate outcomes, aggregated per rule,
upstream family and file. The report retains the complete original input for
every non-agreement. `--reduce` reaches a single-rune-deletion fixed point for
successful byte divergences. This is **not globally shortest**: syntax-preserving
replacement, reordering and shorter equivalent programs are not searched.
Reductions relocate the file while preserving its basename/script extension;
path-sensitive and typed-program reductions need stronger project-preserving
handling. Those limits are recorded in each reduction field. No globally-shortest
or project-preserving claim is made by this first piece.

The current adapter covers TypeScript/JS/Adamic program printing, JSON, YAML,
GraphQL and CSS/SCSS at the existing port's exposed options. Markdown has
component printers but no whole-file port composition API available to this
adapter; it produces a visible blocked cell, not fragment-as-file agreement.
The native Go wrapper is used for the oracle; the port gets defaults matching
that wrapper (JSON/YAML use Prettier defaults, other families use cohere defaults).
Custom formatter options and the resolved per-directory cohere configuration
are not yet wired into this host driver.

The exact shared-file dependency for configured cross-rule fixes is
`stage1/cohere/lint/main.ts` (the manifest passes one decoded JSON bag), with
`context.ts`/`settings.ts` retaining that same bag for every listener. Separate
selected-rule rows can compare findings and per-rule fixes with identical
options today. A simultaneous all-rule run with a distinct bag per rule cannot
be represented by this transport. The scoreboard stops at this boundary: no
shared transport, context or fix-engine edit is made, and configured cross-rule
fix parity is not claimed. Its owner must expose that input contract before
this package can hold those combined answers.


Typed findings require a program config and an already built native lint binary:
`--native` records the existing bridge facts and Node uses the existing replay
mechanism. Without both, typed coverage is blocked. The first public probe run
has neither installed project dependencies nor fabricated checker facts, so it
cannot certify typed findings. Go's independent census likewise records `no program`.

## Open questions for @system_adamic

Both questions below were resolved by @system_adamic. No new language question
is raised by this receipt.

* Host-tool boundary: **the scoreboard may remain Go**. It measures Adamic
  binaries through stdout, exit code and timing; none of it links into the
  product's Adamic build. Source Node remains a separately identified reference
  execution, not evidence of an Adamic binary run.
* Dependency-input boundary: **a snapshot without installed dependencies is a
  different input**. It must be labeled and compared only with another snapshot
  of that kind. Manifests now require `dependency_inputs.go` and
  `dependency_inputs.node`, each `installed` or `uninstalled`, and reject absent
  or differing labels before building or executing adapters. Reports retain the
  labels. `TestDependencySnapshotBoundary` plants both mismatch directions and
  missing/unknown labels; all must fail validation. Matching labels identify the
  kind, not identical installed dependency versions or checker configuration.

Stage 0 successfully built `format.a`; no new language refusal was encountered.
`native-node.json` holds identical native/source-Node output and exit codes for
the debugger and BOM witnesses and real `CohereSettings.json`. This is a native
compile/replay check, not native-versus-Go parity on the known formatter gaps.
Any future native adapter refusal must be recorded in the validation evidence. Its
owner must rule or fix the originating compiler/runtime/port gap. This scout
must not repair a shared compiler, runtime, bridge or parser to get green.

## Fixtures, mutants and reproduction

* 23 real public source probes, exact path/hash/pin: `first-run.manifest.sources.json`.
* Go's real debugger fixture and TypeScript's real BOM fixture:
  `testdata/extra-fixtures.json`; raw `.ts.txt` files stay outside the module graph.
* Synthetic protocol controls are clearly tests of the scoreboard, not source
  corpus: missing-checker, formatter refusal, cancellation, empty/duplicate
  manifest, suggestion edit mutation, and aggregation of a wrong-message result.
* The live mutant is an owned scratch Node wrapper. It runs the unchanged source
  port successfully, changes only `A debugger statement stops execution` to
  `planted divergence`, and requires the ordinary byte comparator to reject it.
  This proves output sensitivity; it does not claim a lint-rule implementation mutant.

```sh
source /workspace/adamic-tools/env.sh
mkdir -p /tmp/adamic-gate
go test ./stage1/cohere/scoreboard -count=1 -v -timeout=30m
go vet ./stage1/cohere/scoreboard
python3 stage1/cohere/scoreboard/fetch-public.py /path/to/scratch/public
python3 stage1/cohere/scoreboard/corpus.py /path/to/scratch/public/fetch.json --out /tmp/probes.json
# --all produces every tracked supported file instead of first-run probes.
go run ./stage1/cohere/scoreboard --fixtures --manifest /tmp/probes.json --out /tmp/report.json
go run ./stage1/cohere/scoreboard --fixtures --manifest stage1/cohere/scoreboard/fixture-run.manifest.json --reduce --out /tmp/fixtures.json
```

A nonzero scoreboard exit preserves its report; it means a byte divergence,
blocked coverage or census failure. It is not a crashed package test. A canonical
Mac manifest must identify all 100 snapshots by repository/root/SHA; the driver
validates each pin and source and refuses to label the 23 snapshots canonical.
A canonical manifest still needs the actual Mac file-selection/options contract;
100 identities alone do not prove that every configured file was selected.

## Validation receipt

All **8 top-level Go tests** pass without skips, including the actual Go/source-Node
fixture and clean-running message mutant. `validation/tests.txt` retains the
complete final run; `validation/vet.txt` is clean. `gofmt -l` and
`git diff --check` are clean. Both scoreboard commands exit nonzero by design
because their reports contain named divergences/coverage blocks; these expected
nonzero receipts are retained. Stage 0 built the owned adapter successfully, and
its three native/source-Node replay checks are byte-identical. No full gate,
full public-source scoreboard sweep or full quiet-hundred sweep is claimed.

## What the next scout takes

Take the driver, the exact snapshot/source inventories, the first-run answers,
the public all-files manifest generator, the byte comparator/partition tests,
and the clean-running planted divergence. Obtain the actual Mac selections and
resolved options; run full available trees before claiming a frequency ranking
for the public portion, then the full quiet hundred. Supply project-specific
checker configs/native recordings; add whole-file Markdown composition only
through its owner. Finish project-preserving reductions and their failure
controls. Do not treat the probe ranking or deletion-minimal inputs as the
requested quiet-hundred ranking or globally shortest inputs.

## Full-tree continuation receipt

`requested-all.manifest.json.gz` enumerates the exact requested eight extensions
from all 23 pinned Git trees: **235,377 files**.
`requested-all.inventory.json` gives per-snapshot and per-extension counts and
bytes. This is an inventory, **not a full-tree scoreboard run**. An independent NUL-delimited `git ls-files -z` census agrees on that count,
including 51 symlinks. The earlier
242,883 denominator used a broader extension set; it is not this request's
denominator. The manifest explicitly labels both execution inputs uninstalled.
Tracked symlink contents are hashed as Git blobs during inventory rather than
following their targets; a future executor must preserve those bytes and path
semantics or report the input unavailable. The current per-file executor reads
the checkout and is not yet suitable for certifying those symlink inputs.

The existing executor also starts fresh Go/Node processes per file and stores
all cells and repeated input bytes in memory. Persistent workers, bounded
streamed reports, resume checks and project-preserving reductions remain
unimplemented. No all-tree agreement counts, missing-rule frequency ranking,
or all-divergence attribution is claimed in this receipt. These are the next
scout's concrete work, not permission or language questions.

`stage1/cohere/lint/main.ts:13` is the lint lane's option transport. A distinct
per-rule option table must be implemented there by that lane; this scout has
not changed it. The current scoreboard's adapter-options/census-options
dialect distinction remains material.

The named `comment-loss` and `BOM preservation` fixtures are retained in
`testdata/extra-fixtures.json` with hashes and provenance, and both are asserted
as successful-but-different Go/Node executions in the live package test.
`rulings-fixtures.json` records nine cells: seven agree and these two formatter
fixtures diverge, with dependency labels.
The BOM fixture keeps its original CRLF bytes; package-local `.gitattributes`
marks it binary for diffs. Earlier validation `.log` receipts were ignored by
the repository's global pattern; this receipt includes fresh tracked `.txt`
receipts instead.
