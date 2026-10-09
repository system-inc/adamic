# Stage 1 progress dashboard

`go run ./cmd/adamic-stage1-progress --pending` lists every production Go package
under the pinned `cohere/internal`, its non-test physical Go line count, credited
stage 1 lines, recorded status and GAPS.md evidence. It reads Git snapshots
without checking out, merging or changing branches. `--json` emits the same
inventory and totals for automation. `--ref <ref>` selects one snapshot.

This is a dashboard of recorded slice evidence, not a fresh run of every port
oracle. "Ported and held byte-identical" means the mapped package has complete
source credit and no recorded package scope exclusion. The underlying slice
GAPS.md/test reports establish the native/Node/Go comparison on their stated
corpora. It does not mean all possible inputs have been exhaustively tested.

## Measurement on main and pending branches

Measured 2026-10-06 after fetching the full requested ref patterns. All measured
snapshots pin cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`. Denominator:
**382,804 physical lines in 1,455 non-test Go files, 121 packages**.
An independent filesystem census reproduced all three counts.

| Snapshot | Commit | Credited Go lines | Percent | Complete / partial / not started packages |
| --- | --- | ---: | ---: | --- |
| origin/main | 1a4e299 | 8,308 | 2.1703% | 4 / 9 / 108 |
| origin/codex/stage1-css | e8e03a8 | 8,888 | 2.3218% | 5 / 5 / 111 |
| origin/codex/stage1-css-printer | 6476d7a | 12,089 | 3.1580% | 5 / 5 / 111 |
| origin/codex/stage1-markdown-blocks | 12912b9 | 6,662 | 1.7403% | 4 / 7 / 110 |
| origin/codex/stage1-markdown-escape | f65e4e0 | 6,361 | 1.6617% | 4 / 4 / 113 |
| origin/codex/stage1-yaml | 0d694d7 | 8,308 | 2.1703% | 4 / 9 / 108 |
| origin/codex/typescript-scanner | ed2477e | 6,981 | 1.8236% | 4 / 4 / 113 |
| Main plus pending evidence union | Main pin | **14,075** | **3.6768%** | **5 / 12 / 104** |

[Main: every package and scope](stage1-progress/main.txt).
[Pending snapshots: every package and scope](stage1-progress/pending.txt).
[Machine-readable reports with full commit pins](stage1-progress/snapshots.json).

A branch's percentage describes its own complete tree. Older branch bases lack
ports subsequently integrated into main, explaining the lower Markdown/scanner
numbers. The union counts covered upstream line identities once; it adds 5,767
unique mapped lines to main. It is an **evidence union, not a successful merge or
executable integration**. In particular, the raw CSS branch still records native
composition as blocked, while the CSS-printer branch records closure and native
parser/printer parity. No branch's old refusal is silently replaced with main's
compiler behavior.

Already integrated refs, excluded from the pending set by Git ancestry:

* `origin/codex/stage1-css-numbers` at `8909b74`.
* `origin/codex/stage1-css-selector` at `c717c5d`.
* `origin/codex/stage1-formatter-slice` at `f69f6a5`.
* `origin/codex/stage1-json-format` at `67796a5`.

The saved census predates creation of this dashboard's own
`codex/stage1-progress` branch. Later `--pending` runs also list that branch until
integration; it carries the same stage 1 coverage as its main base and contributes
no new cohere lines. Saved reports retain the original measured ref pins.

These integrated slices still contribute to main. The scanner's latest branch
has new commits and is pending even though its earlier work was integrated.

## What the percentages mean

The denominator includes comments, blank lines, generated Go and all platform
variants. A final line without a newline counts as a line. `_test.go`, directories
named `testdata`, and hidden directories are excluded, following Go's package
inventory convention. Bundled JavaScript, assets and TypeScript submodule source
are outside this Go denominator. Build-ignored Go generators remain included:
this is physical source size, not a build-specific executable instruction count.

The numerator is a conservative **reviewed source footprint within the slice's
recorded contract**, not the percentage of cohere functionality implemented.
`coverage.go` maps each known stage 1 directory to complete upstream files or
complete named functions. Partial functions, missing mappings and composition
that has only Node evidence receive no line credit. This deliberately undercounts
working JSON doc/shared-JavaScript paths, selector parsing, Markdown layout,
quote-prefix recognition and other partially mapped code. It never credits an
entire parent package merely because one helper has a working port.

Full-file credit includes its comments/imports/blank lines. Function credit covers
its parsed start/end lines and attached documentation; unrelated declarations
and surrounding file scaffolding are excluded. Credit is a set of upstream
file/line identities, so reused functions, shared slices and multiple branches
are counted once. An alternative logical-SLOC count would give different weights;
this report consistently uses the requested physical non-test Go lines.

Status is intentionally stricter than line coverage:

* **Ported and held byte-identical:** every source line is mapped, required port
  and parity-test files are present, and no package-level scope exclusion is
  recorded in the reviewed mapping. Main's four are gitignore, mediaquery,
  values and directive grammar.
* **Partly ported:** a slice maps to the package but is incomplete, has an explicit
  scope exclusion, or lacks the required production/native evidence. GraphQL's
  parser is mapped but its printer is not. Suppression has all four files mapped
  for the ordinary Index contract, but GAPS.md excludes nil-Index methods and
  concurrency, so it stays partial even with 658/658 source-footprint credit.
* **Not started:** no mapped Adamic implementation is present. The YAML directory
  records only a Go/original-library baseline audit; its mere GAPS.md does not
  earn implementation credit or change YAML packages to partial.

The dashboard reads each existing directory's whole GAPS.md, displays its first
paragraph and records its scope note. Composition credit also requires the
specific recorded native gap closure. Required source/test names are checked
against the same Git tree. Missing mapped Go files or renamed/ambiguous mapped
functions fail, rather than silently reducing or inflating credit. A new unknown
stage1/cohere directory fails with an instruction to add a reviewed mapping.
The registry supplies semantic source attribution that directory names and free
prose cannot reliably infer. Keep it updated when a slice's contract expands;
do not treat arbitrary positive wording in GAPS.md as a certificate.

`stage1/typescript/scanner` and `parser` are reported separately with their GAPS.md
summaries. Their upstream implementation lives in
`cohere/TypeScript/tsc/internal`, not `cohere/internal`. They are important shared
prerequisites but add **zero cohere-internal Go lines** to this dashboard. Their
own package/function completeness is not surveyed here.

## Commands and validation

```sh
# Discover all refs, not only this checkout's original narrow fetch set.
git fetch origin '+refs/heads/codex/stage1-*:refs/remotes/origin/codex/stage1-*' '+refs/heads/codex/typescript-scanner:refs/remotes/origin/codex/typescript-scanner'
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic-stage1-progress --ref origin/main > docs/stage1-progress/main.txt 2> /tmp/stage1-progress-main.stderr
go run ./cmd/adamic-stage1-progress --pending > docs/stage1-progress/pending.txt 2> /tmp/stage1-progress-pending.stderr
go run ./cmd/adamic-stage1-progress --pending --json > /tmp/stage1-progress.json 2> /tmp/stage1-progress.stderr
go vet ./... > /tmp/stage1-progress-vet.log 2>&1
go test ./cmd/adamic-stage1-progress ./cmd/adamic-meter ./internal/load -count=1 > /tmp/stage1-progress-touched-final.log 2>&1
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(strings|collections|exceptions)\.a$' -count=1 -timeout 30m > /tmp/stage1-progress-oracle.log 2>&1
```

Dashboard commands exit 0 with empty stderr. Vet and formatting checks pass.
Final touched tests pass: progress 0.088s, meter 0.516s, loader 0.635s.
The filtered native/Node oracle passes in 10.867s. No compiler/runtime or slice
file changed. The complete gate was not rerun for this read-only CLI addition:
the immediately preceding unit's gate took over nine minutes in unrelated
exhaustive Unicode scans; this unit ran the touched packages and filtered oracle.
Individual port corpora were not rerun; the displayed statuses retain their
existing documented evidence and coverage limits.

Fixtures verify physical line endings, generated/platform inclusion, test/fixture
exclusion, bounded declaration credit, fail-closed renamed functions, overlapping
line union, incompatible cohere pins, Git-snapshot reads despite a dirty worktree,
missing parity evidence, audit-only YAML, unknown slices and pending-ref ancestry.
Three actual Go-source mutants were run through file overlays, leaving production
source unchanged; all compile and fail only their metric/evidence assertions:

| Mutant | Catch |
| --- | --- |
| Double credited line counts | `TestUnionDoesNotCountSharedLinesTwice`: wrong shared-line total; Git snapshot fixture also rejects excess line credit; exit 1 |
| Bypass the required parity-test file check | `TestGitSnapshotAndEvidenceBoundary`: removed test still receives complete status and line credit; exit 1 |
| Count `_test.go` as production | `TestPhysicalInventoryBoundaries`: test source admitted; independent Git fixture denominator also becomes wrong; exit 1 |

Setup succeeded: Go 0s, clang 0s, Node 1s, submodules 1s, build-cache warm 78s,
total 78s. `nproc` reports 5; cgroup quota is 4 CPUs, memory 17.6 GB.

The generated table renderer was corrected in a follow-up commit to omit blank
trailing evidence columns. Both saved tables have no trailing whitespace; the
package tests pass after that presentation-only change. Source credit, branch
pins and the saved JSON metrics are unchanged.
