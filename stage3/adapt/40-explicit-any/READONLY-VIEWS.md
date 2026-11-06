# Readonly enum views, task 52zja36

Base: origin/main 63e14a0b2b61cf251ad74292e42d5fd09af804b3.
The 21 requested replacements are implemented, plus the cache-key type repair
explicitly authorized by the brief. No cast, statement or runtime expression is
added. Every edit is under adaptation 40.

Rules 72 through 89 and 91 now use
`Readonly<Record<"K", Readonly<Record<string, string | number>>>>`.
Rules 69 and 71 give formatEnum and getEnumMembers the parameter type
`Readonly<Record<string, string | number>>`. The outer Readonly protects the
namespace field, and the inner Readonly describes the enum table read view.

Rule 70's identity Map key also uses
`Readonly<Record<string, string | number>>`. The first after latent census found
two new NotYet diagnostics at cache.get and cache.set: the readonly parameter was
still being passed to mutable record storage. The cache never writes through
its keys. Changing its key annotation to the same readonly view removes both
stops. See [cache-key-reason.json](evidence/readonly-views/cache-key-reason.json).
The final scope audit is [rules-scope-final.json](evidence/readonly-views/rules-scope-final.json):
only the replacement fields of these 22 rules differ; every other rule and field
is unchanged. class-proof.cjs's projection and Boolean-map mutant use the new
parameter spelling. Previous evidence remains historical.

## Check 1: full apply and paired replay

Both baseline and final full applies with --build-product exit 0. The final apply
runs every later adapter in order over 40's new debug.ts. Therefore there is no
later owner conflict. Per-adapter Git objects show these later adapters change
debug.ts:

- 41-explicit-any-remaining: getFunctionName.
- 76-truthful-casts: finite assertion-key declaration and cache-use checks.
- 99-cli-debug-assertion-level: assertionCache, shouldAssertFunction and setAssertionLevel.
- 99-correctness-axolotl-n: formatControlFlowGraph.hasNode.
- 99-correctness-axolotl-t: DebugTypeMapper.__debugToString, enableDebugInfo and formatControlFlowGraph.hasAntecedents.

[later-debug-adapters.json](evidence/readonly-views/later-debug-adapters.json)
records the exact directory list. The evidence distinguishes active pins from
99n's proposed-pins.json, whose enum-owner proposals are not applied.

The requester's original replay condition was corrected. Adaptation 30's
addRange refusal at core.ts:992:21 occurs on main's own completed tree because
later adaptations change that owner. Earlier adapters need not be idempotent on
the final composed tree. The historical stop and partial census are retained.

The corrected replay runs every adapter in order on disposable copies of both
complete trees and continues after refusals. It records exact stdout, stderr,
exit status, file-write calls and files written. Both copies use the same locked
dependencies. Only the tree prefix is normalized. The initial 21-rule replay has
identical records for all 80 adapters. The final cache-aware replay also has identical records for all 80 adapters,
including 40. Both records and the empty diff are retained under
[pieces/paired-replay-final](evidence/readonly-views/pieces/paired-replay-final).

## Checks 2 through 6

| Check | Result |
| --- | --- |
| 2: adaptation 40 enum proof | Pass: zero new consumer diagnostics, all nine JavaScript files identical, 396 declarations compared, idempotence and eight mutants caught |
| 2: stage3 consumers | generate_shards --check passes; lane tests pass, 129.051 seconds |
| 3: adamic types | Final tree, all 79 sorted compiler roots together: exit 0, zero diagnostics |
| 4: normal product | All nine emitted JavaScript files byte-identical; typescript.d.ts and tsserverlibrary.d.ts byte-identical; no public line moves |
| 5: cast-specific latent stops | All 19 change from writable-record Refused to namespace-value NotYet |
| 5: ordinary census | Both 79-root runs: 73 Refused, 4 NotYet, 2 accepted; zero changed first stops |
| 6: inner-only mutant | Pass: restoring only SyntaxKind's inner-only Readonly still refuses through invariance.go:673 |

The only declaration differences are formatEnum's readonly parameter in
compiler/debug.d.ts and typescript.internal.d.ts. No public declaration exception
or sanction is needed. [final-artifact-comparison.json](evidence/readonly-views/final-artifact-comparison.json)
records the exact source scope, sizes and hashes. Both ordinary compiler builds
pass. Sources differ only in debug.ts and emitter.ts, exactly as the rule audit
predicts. The cache key and all casts are erased by emission.

The rule-consumer search, current enum class proof and stage3 lane tests are
retained in the evidence. Historical brand and classification snapshot scripts
describe older transformations; the current enum test uses the real
40-after snapshot with only the 23 enum-display edits undone, builds before and
after, and checks the existing class-proof.cjs assertions and mutants. The final
run includes the readonly cache key. The lane test command follows 99w's proof,
excluding only separately managed product and shard cases:

```sh
python3 stage3/lane/generate_shards.py --check
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source go test -count=1 -parallel=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/
```

## The 19 sites and measurement limits

[SITE-STOPS.md](evidence/readonly-views/SITE-STOPS.md) lists every site and owner.
[site-stops.json](evidence/readonly-views/site-stops.json) retains exact reasons,
locations and phase-specific findings. Every before cast has the writable-record
Refused from internal/lower/invariance.go:673. Every final cast has the expected
namespace-value NotYet from internal/lower/module_namespace.go:27. No final cast
has a Refused scan finding. Successful native lowering is not claimed.

The before and final full latent pieces are restricted to debug.ts. Each has
105 units. The final cache repair removes the two additional storage stops from
the initial after piece. The nineteenth site is emitter.ts:5790:141 in makeName;
it uses the same refusal scan without walking every unrelated emitter body.
The scan-only variant matches all 57 before-debug refusal-scan findings exactly,
keeps the no-output guards, and changes only the scratch census driver in
[scan-only-units.go.txt](evidence/readonly-views/scan-only-units.go.txt).
Production compiler Go is unchanged.

The emitter single-unit lowering probe independently hits an earlier boundary,
emitter.ts:5790:84 reading autoGenerate, on both trees. This is retained rather
than presented as successful cast lowering. The full preliminary emitter walk
was stopped when the focused evidence identified this boundary. The ordinary
census remains the separate production-compiler first-stop measurement.

Every latent piece is bounded by timeout 1490s with a 10s termination grace,
so no piece exceeds 25 minutes. Pieces were committed and pushed as completed.
The final mutant is an isolated copy of final src: only one cast's outer Readonly
is removed. Its patch and source audit have filenames containing mutant.
The mutant reproduces the refusal even with the repaired cache key.

## Reproduction and retained failures

Use TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8 and the setup
recorded by 99w's proof. Source /workspace/adamic-tools/env.sh, install stage3/api
from its lockfile, and set ADAMIC_TYPESCRIPT_SOURCE. The baseline full apply is
from the base SHA; the final full apply uses the current adaptation 40 rules.
Every check's stdout and stderr are logged to files, with compressed copies
retained under evidence/readonly-views. No log was piped into head or tail.

The retained scripts are paired-replay.py, replay-trace.cjs,
compare-artifacts.py, locate-sites.cjs, report-sites.py, report-ordinary.py,
focused-proof.cjs, and run-class-proof.py. Run the class proof reproducer with
the final applied tree and a fresh scratch directory as its two arguments.
The existing latent overlay is built by stage3/census/latent/make_overlay.py;
scan-only-units.go.txt adds only an early emit/continue after its existing scan.

Exploratory failures are retained: a missing NODE_PATH, an initial comparison
that matched a cache annotation instead of a named parameter, a snapshot clone
without custom tree objects, and the scan variant's initially omitted os import.
They were corrected and the affected checks rerun. Historical initial-after
results are labeled separately from final cache-aware results.

The completed ordinary comparison is in [ORDINARY-CENSUS.md](evidence/readonly-views/ORDINARY-CENSUS.md)
and ordinary-comparison.json. Every first stop is identical after normalizing
only the tree root and omitting timing. The final debug latent delta has zero
new Refused findings (202 before findings, 184 final findings). All requested
checks are complete. Started processes were completed or terminated; the final
cleanup audit is retained with the evidence.
