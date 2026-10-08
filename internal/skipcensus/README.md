# Source-local skip declarations

Each owned Go test Skip, Skipf or SkipNow call has a declaration on the immediately
preceding line. The call occupies its own line; no blank line or unrelated comment
may intervene. Comments and string literals containing example calls are ignored.

Grammar (the literal prefix is `// census:`):

```text
annotation = "// census: " class [ " " description ]
class = "measurement" | "not-applicable" | "required-input" | "opt-in-lane"
```

Examples:

```go
// census: measurement
 t.Skip("profile is opt-in")
// census: not-applicable Linux-only LeakSanitizer witness.
 t.Skip("platform")
// census: required-input ADAMIC_TYPESCRIPT_SOURCE: setup supplies pinned TypeScript v6.0.3 source.
 t.Skip("compiler source absent")
// census: required-input PATH: setup supplies clang and Node.
 t.Skip("tool missing")
// census: required-input Gate tree supplies ir.Program.Regexps and native regex lowering.
 t.Skip("source dependency missing")
// census: opt-in-lane ADAMIC_GCC_LANE=1 runs the separate GCC verification shard.
 t.Skip("lane is off")
```

Measurement may omit its description. Not-applicable and opt-in-lane must state
a reason. Required-input must state a provider and name an uppercase environment
variable (including PATH) or an actual qualified source field such as
ir.Program.Regexps. Source dependencies must not invent environment variables.
This preserves all four existing classes. A skip reached after an opt-in guard
must still be required-input.

The scan is the declaration source; there is no skips.json and no joined-caller
identity. IDs use the lexical enclosing function and condition; discovering a
new caller cannot re-key the site. Callers remain runtime-attribution metadata.
The degraded-input rule is not on this base yet. When it lands, its matching log
sites must use exactly this preceding-line declaration grammar.

Integration continues to run from the repository root:

```sh
go run ./internal/skipcensus/cmd <log.jsonl>
go run ./internal/skipcensus/cmd -audit
go run ./internal/skipcensus/cmd -scan
go run ./internal/skipcensus/cmd -table
```

`-table` exports the same JSON array as `-scan`, ordered by file and line.
Each row includes test, callers, class, provides, file and line. It audits the
annotations before exporting; integration can replace its copied skips.json
with this command.

Scan returns an error for missing or invalid declarations, naming file, line and
function. CheckLog consumes its rows; missing required inputs and unknown runtime
skips fail closed. Keep the ordinary test verdict as well as the skip verdict.
The gate runner now calls Scan directly instead of opening the deleted table.
No result cache is added; ADAMIC_GATE_UNCACHED=1 computes the same inventory.

The AST migration command is `go run ./internal/skipcensus/cmd/annotate -table
<old-table.json>`. It matches historical rows by file, lexical function, printed
condition and message, locates actual calls through go/ast positions, inserts
comments, formats Go and verifies every original class and provider afterward.
It rejects unmatched rows instead of guessing a classification. Inventory and
Load exist only for this migration; production callers use audited Scan.

## Conversion, performance and validation

Base: f8dc89f215febd29f2b523edfab987428b82776a. The migration converted all 75
historical table rows and verified class and provider equality for every site.
Only comments were added to the 75 original skip sites; test behavior is intact.
IDs now retain the lexical helper name when its caller set changes.

Measured on the same working tree and box, alternating the two compiled binaries
for three rounds. The baseline binary was built before the scanner edit, at the
same base commit. The new binary performs annotation validation too. All three
new runs were below 0.245 seconds. There is no claimed speedup.

| Loop | Before, best of 3 | After, best of 3 | Instrument |
| --- | --- | --- | --- |
| AST scan and inventory JSON | 0.195472 s | 0.195727 s | `/tmp/census-audit-before -scan`, `/tmp/census-audit-after -scan`, stdout to separate files, timed with time.perf_counter |

Build-flags: base SHA above; nproc 5; cpu.max 400000 100000; Go 1.27.1
linux/amd64; clang 20.1.8; Node v24.19.0; no census cache; compiled binaries;
load before 0.03 0.51 0.35, after 0.35 0.57 0.37. Full samples and flags are
in [timings.json](testdata/annotation-proof/timings.json). Ordinary and
ADAMIC_GATE_UNCACHED=1 inventories were byte-identical. No cache was added.

Both affected package suites pass uncached:
`ADAMIC_GATE_UNCACHED=1 go test -v -count=1 ./internal/skipcensus/... ./cmd/adamic-gate`.
Vet passes for both packages. The historical plain log still reports exactly
17 required-input skips among 33, including TestWholeCompilerAgrees.
The reproducible mutation command is `python3 internal/skipcensus/testdata/proofs.py`:
a scratch new skip without an annotation, a bad class, and a required-input
without its variable each makes the real TestCensus fail, naming file:5 and
TestAnnotationWitness. Unit tests additionally cover adjacency/orphans, helper
aliases, stable IDs after adding callers, opt-in classifications and an inline
AST conversion. Proof outputs are in [annotation-proof](testdata/annotation-proof).

Setup exited 1 after verifying module dependencies: the base has a compiler API
mismatch in internal/metamorphic/source.go:439 (`parseName(path)` is string where
RootedFilePath is required, SourceFileParseOptions has no Path field, and
tspath.Path is undefined). This unit does not change that unrelated package.
The full live gate was not run. Existing installed tools built the census and
gate consumer; setup's log is preserved with its own timing lines.

The integration export proof is `python3 internal/skipcensus/testdata/table_proof.py`.
It invokes `go run ./internal/skipcensus/cmd -table` twice and compares all 75
rows with the historical table at f8dc89f2, matching by file, lexical function,
condition and message. Class and provides are byte-for-byte equal; output order
and required integration fields are checked. Disabling the table export makes
this proof fail. This is a historical migration proof, not a permanent census
row-count constraint on future work.

After merging 1db5ab2b1506e699233022eefba0ee7110fe24ce, the export proof
defaults to that historical table and checks all 76 rows. An optional commit
argument selects another historical table; the comparison has no fixed count.
The new optional-widening inventory skip is annotated measurement with its
original provider text. The central table remains deleted.

Merged-tree validation: all census packages pass uncached, and repository-wide
`go vet ./...` passes. The incoming metamorphic API fix resolves the setup
compilation failure recorded above.
