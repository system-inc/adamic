# scan-all — task #et209ny

**Records refusal-scan stops only, plus the first declaration, body or tail stop
reached; blind past that first stop. A zero on a body-phase or tail rule is never
a pass.** This is scout design C, packaged from `0d078a8d`, using current main's
orchestration and the scout's pinned independent bind and record rule heads. It
is measurement tooling, never a compiler acceptance result, and never an all-stop
lowering implementation.

One invocation loads **one root's import closure**. It continues the refusal scan
at the scout's recovery seam, then runs production declaration, body and tail
lowering until the first stop. A rule/preflight may itself return only its first
error; only the first winning registry rule is reported per visit to a node, and a rule
can prune the remaining rules at that node. The continuing walker still visits
children. Deferred scan fallback is not an additional independent finding. No
recovery is attempted after a declaration, body or tail stop, including a record
stop before bind checking. Retries can record the same triple again.

The header and every output table repeat these limits. Events and distinct
`(rule, site, words)` triples are separate counts. File buckets are parsed from
recorded `site` values (`file:line:column`); missing sites are explicitly unlocated.
The returned production diagnosis may replace the underlying abort with a deferred
host-write diagnosis. The report preserves that underlying boundary separately;
it is not counted twice. `Completed` means the measurement pipeline completed
after scan findings, never that the source compiles.

## Build outside the checkout

Use the repository's pinned Go toolchain and initialized recursive submodules.
Set `GOTMPDIR`, `TMPDIR`, `GOCACHE` and `ADAMIC_BUILD_CACHE_DIR` to your scratch.
From the repository root, with `$scratch` outside the repository and its parent
already present:

```sh
go run ./stage3/census/scan-all/make-overlay . "$scratch/overlay"
go build -overlay "$scratch/overlay/overlay.json" -o "$scratch/scan-all" ./stage3/census/scan-all/tool
"$scratch/scan-all" /absolute/path/to/one/root.ts > "$scratch/summary.md"
```

The generator requires a **new** outside scratch directory, resolves its parent
and the repository's symlinks, and rejects production destinations. It reads
`internal/lower`, writes copies only to scratch, and generates an overlay map.
Both public `Lower` APIs unconditionally return nil IR plus an error in this
binary; the driver and all controls check both guards. The measurement
entry returns diagnostics and completion only, with no IR or emit/cache handle.
Nothing writes production source. The carried `heads/*.go.txt` are exact scout
artifacts: bind from `ba619e28`, and cycles/graph-records/records from `b529db90`.
The generator also inserts their `graphRecords` field into the copied lowering
struct. These are independent rule heads, **not a claim they are merged into
main**. In particular main at the example base lacks the b01 declared-parameter
rule; substituting main's bind file makes two required controls red. The report
header names these overlays so results cannot be mistaken for plain main.
There is no native backend in the driver.

Every rewritten literal anchor must match exactly once. The production
`lower.go`, `refusals.go` and all four rule-head replacement bases are also
SHA-256 anchors: *any* drift refuses
generation and names the file anchor. The latent refusal transformer checks its
AST anchors too. Updating an anchor requires reviewing the new orchestration and
re-running the controls; never update hashes just to make a build pass.

## Tests and first-error mutants

```sh
go test -overlay "$scratch/overlay/overlay.json" -count=1 -timeout=60s ./stage3/census/scan-all/...
```

An ordinary build of `tool` refuses to run and directs you to the overlay
instructions. Both the generator package and generated compiler driver carry
explicit compiler consumer markers and dependency declarations.

The ordinary package tests check current-main anchors, drift rejection and the
scratch boundary. Without the overlay they explicitly skip the eight measurement
controls: production lower deliberately has no measurement API. **For proof, use
the overlay invocation above**, which replaces only the test bridge and enables
all eight tests. Each test calls `t.Parallel()`, loads a small checked fixture,
and calls the measurement API directly; tests run no build subprocesses. The
measurement entry serializes its recorder and checker work, so parallel tests
cannot mix their events.

All fixtures begin with an earlier `@ts-check` scan stop. Four retain the scout's
b01 bind stop, record stop and their clean twins. Four combine their positive and
clean variants: a record stop prevents observation of a later bind stop. Clean
twins must complete, rather than merely lack the target words.

Each fixture ships with the same explicit behavioral mutant: pass its name via
`SCAN_ALL_MUTANT_FIRST_ERROR` to restore the untouched production `refuse` for
that test. The target tail stop (or the clean twin's completion) then disappears,
so **the same named test must turn red**:

- `bind-stop`: `TestScanAllCensusBindStop`
- `bind-clean`: `TestScanAllCensusBindClean`
- `record-stop`: `TestScanAllCensusRecordStop`
- `record-clean`: `TestScanAllCensusRecordClean`
- `combined-bind-stop-record-stop`: `TestScanAllCensusCombinedBindStopRecordStop`
- `combined-bind-stop-record-clean`: `TestScanAllCensusCombinedBindStopRecordClean`
- `combined-bind-clean-record-stop`: `TestScanAllCensusCombinedBindCleanRecordStop`
- `combined-bind-clean-record-clean`: `TestScanAllCensusCombinedBindCleanRecordClean`

For example:

```sh
SCAN_ALL_MUTANT_FIRST_ERROR=bind-stop go test -overlay "$scratch/overlay/overlay.json" -count=1 -timeout=60s ./stage3/census/scan-all -run '^TestScanAllCensusBindStop$'
```

`TestScanAllCensusMainAnchorsStillMatch` guards main's current source;
`TestScanAllCensusDriftNamesItsAnchor` changes each snapshot and requires a named
failure with no partial overlay; `TestScanAllCensusScratchCannotBeProduction`
checks the destination boundary.
`TestScanAllCensusRuleHeadBaseDriftFailsLoudly` mutates each wholesale replacement
base and checks its named rejection.

## Example provenance and isolation

`examples/` contains summary tables only, from one sys.ts closure run and a second
root closure, with wall time, peak RSS and source pins. No raw observation stream
is checked in. Prepare the adapted TypeScript input using main's `stage3/apply.sh`
with `ADAMIC_TYPESCRIPT_SOURCE` set to TypeScript v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8` and `npm ci` in `stage3/api`.

No gate consumes these reports, their counts or completion. To audit references,
from the repository root:

```sh
rg -n 'scan-all|ScanAll|SCAN_ALL' cmd cloud internal stage3/meter stage3/fixtures
```

Only the declared compiler dependency is expected. This package is a compiler
consumer for dependency selection; its output is not gate input. Do not submit
this work to Loom.
