# 99u proofs and commands

Landing base: main a3f9467233562ef1d8834b064dddf1a149dd036c. Complete Node and 79-root census
measurement base: b0efce6dce1cd7445ed8a570aecac387a4a7d684. All executable adapter
and owner-pin inputs are frozen in provenance.json. Later evidence archives
and prose change the keyed product's inputs, not the source transformation.
Only this new adaptation directory is committed. No other adaptation, compiler
policy, oracle, lane checker or generation script is edited.

| Proof | Observed result |
| --- | --- |
| Kind-pin gate before census/build | Stock 6.0.3, zero diagnostics in each owner |
| Complete stock before/after checks with pinned Node ambient declarations | 79 roots, zero diagnostics on both sides |
| Adamic types, ordinary main binary | 79 roots, exit 0 and zero diagnostics before and after |
| Count-first unmodified full oracle | 106,366 pass, one sanctioned failure, no pending |
| Count-first identity versus uncounted main | Raw and normalized multisets exact; zero additions/removals |
| Planted counter controls | Eleven exact-owner paths; each counter fires once and fails once |
| Shipped case conditions | Ten conditions, each nonzero coverage and zero failures |
| Rejected default isExpression condition | 26 runs, 12 failures; original cast retained |
| Final built source full oracle | 106,366 pass, same sanctioned failure, exact title/outcome identity |
| Shipped assertion negative witness | Ten conditions fail at levels 0, 1, 2 and 3; forty expected failures |
| Owner/import contract | 43 cases pass; 38 no-write refusals; unrelated edits and LF/CRLF accepted |
| Full stage3/apply.sh | Exit 0 |
| Keyed second apply | Exit 0; compared source payload byte-identical |
| Direct adapter replay | files:0, edits:0; no writes |
| Four exported internal predicates | Signatures unchanged; both public declarations byte-identical |
| Complete source scope | 746 files compared; only two files, five pinned owners and required imports change |
| Syntax audit | Five Debug.type calls removed; no added casts; existing default cast unchanged |
| Real adaptation sort | 99u is after every 99-correctness-axolotl directory on the base |
| generate_shards --check | Exit 0 |
| Stage3/lane Go fixtures | Exit 0, 93.717 seconds on b0efce6d; focused latest-compiler check also passes |
| Ordinary per-file compiler census | All 79 roots before and after; 73 Refused, four NotYet, two accepted on each side |
| Meter full-body owner observations | All five complete owners on both sides, all 79 roots loaded, zero checker diagnostics, both no-output guards pass |
| Kind-pin mutant | Drop only Parameter label; exactly one stock TS2345 diagnostic at the kind pin |
| Debug.type lowering mutant | Restore one line in first owner; checker-clean; exact failed-call Boundary carries a NotYet dependency cause |

## Concrete commands

Tool setup, compiler builds and every run write logs to files. The source pin
is `/workspace/scratch/99u-upstream`. The baseline applied tree is
`/workspace/scratch/99u-main`; the accepted full apply is
`/workspace/scratch/99u-adapted`; the independently adapted measurement source
is `/workspace/scratch/99u-measured`. comparison.json verifies that all source
bytes in the latter two trees are identical.

```sh
source /workspace/adamic-tools/env.sh
GOPROXY='https://proxy.golang.org|direct' timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/99u-build-setup.log 2>&1
go build -o /workspace/scratch/99u-bin/ ./cmd/adamic ./stage3/census/tool
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/99u-upstream stage3/apply.sh /workspace/scratch/99u-adapted > /workspace/scratch/99u-full-apply.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/99u-upstream stage3/apply.sh /workspace/scratch/99u-replay > /workspace/scratch/99u-full-replay.log 2>&1
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-u/contract.cjs /workspace/scratch/99u-main
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-u/adapt.cjs /workspace/scratch/99u-adapted
python3 stage3/lane/generate_shards.py --check > /workspace/scratch/99u-shards.log 2>&1
GOMAXPROCS=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/99u-upstream go test -count=1 -parallel=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/ > /workspace/scratch/99u-lane.log 2>&1
```

The sorted 79 paths under each src/compiler directory are passed together to
`adamic types`. types-before.json and types-after.json retain root counts,
exit codes and timings; the corresponding compressed logs are empty. The
stock checker uses the same strict settings, ES2024, Bundler resolution,
noEmit and erasableSyntaxOnly false, plus the existing pinned Node ambient
declarations that supply stock TypeScript's host globals. All exact options
and the sole mutant diagnostic are in stock-checks-and-kind-pin-mutant.json.

Each oracle command below uses the base's unchanged stage3/oracle/run.sh with
no additional test flags, filters, timeout changes or preparation shortcut.
The preload observes identity without changing Mocha's original events.

```sh
NODE_OPTIONS=--require=/workspace/scratch/99u-census-preload.cjs CLAIM_IDENTITIES=/workspace/scratch/99u-reference-identities.jsonl stage3/oracle/run.sh /workspace/scratch/99u-main /workspace/scratch/99u-reference-oracle > /workspace/scratch/99u-reference-oracle.log 2>&1
NODE_OPTIONS=--require=/workspace/scratch/99u-census-preload.cjs CLAIM_IDENTITIES=/workspace/scratch/99u-counted-final-identities.jsonl CLAIM_COUNTS=/workspace/scratch/99u-guard-counts.jsonl stage3/oracle/run.sh /workspace/scratch/99u-counted /workspace/scratch/99u-counted-oracle-final > /workspace/scratch/99u-counted-oracle-final.log 2>&1
NODE_OPTIONS=--require=/workspace/scratch/99u-census-preload.cjs CLAIM_IDENTITIES=/workspace/scratch/99u-adapted-identities.jsonl stage3/oracle/run.sh /workspace/scratch/99u-adapted /workspace/scratch/99u-adapted-oracle > /workspace/scratch/99u-adapted-oracle.log 2>&1
```

Install and build exit 0 in all three runs. Tests/oracle exit 1 solely for the
sanctioned public API baseline; baseline_diffs is exactly
`["api/typescript.d.ts"]`. Reports and all raw identity streams are retained.
compare-identities.py requires every raw multiplicity to be even, compares
both exact raw and normalized multisets, and preserves repeated titles.

## Compiler observations and controls

CENSUS.md describes the unmodified ordinary per-file driver and the supported
meter full-body selector. Both observation sides load all 79 roots with no
checker diagnostics. The generated meter overlay disables ordinary Load and
Lower output, and both guards are checked by the .go.txt observation drivers.
No working compiler is built from that overlay, and it has no backend output.
The ordinary Adamic and census binaries have no overlay.

mutant-owner.go.txt records the lowering control through that same attempt
path. It validates the exact restored call's failed-statement Boundary and
requires its text to equal a NotYet finding in the same selected unit. The
root coordinate is in binder.ts, not falsely reported at the caller. The
existing replay API's signature-mismatch metadata is retained because its
selector coordinate is the caller, while the error originates in a dependency.
The adapted owner has a later child-predicate dependency boundary, not that
restored-call boundary. debug-type-mutant.json retains the entire trace and
explicitly distinguishes these coordinates.

The first owner is used for the single-line lowering control because restoring
HasInferredType makes the default's node.kind ill-typed after its false
assertion narrows node to never. That rejected trial is retained and is not
counted as a lowering result. The separate kind-pin mutant changes only one
hasInferredType case label, produces exactly one stock diagnostic, and is bound
to the accepted owner by its recorded source and options.

## Pilots kept separate

- The first counted-copy preparer hit Node's default output-buffer limit while
  reading the baseline diff. Its premature uninstrumented build was stopped.
  Preparation was fixed and required to succeed before the final counted run;
  that final run uses fresh identity/output files. pilot-counted-prepare.log.gz
  records the discarded attempt.
- A stock checker invocation without Node ambient declarations reported the
  same 59 host-environment diagnostics seen in the initial isolated gate.
  The completed stock checks supply the already installed pinned Node types;
  they have zero diagnostics before and after, and one only in the kind mutant.
- The first runtime-level witness used stock Debug while transpiling at changing
  assertion levels and encountered stock's existing assertion-cache behavior.
  The completed witness uses the fully built adapted compiler's Debug and guard
  functions, so it exercises main's existing assertion-level adaptation. All
  forty intended failures occur. No other adaptation was edited to do this.
- Two exploratory whole-selected-file scans were stopped incomplete after the
  ordinary full per-file census and all requested complete owner observations
  finished. full-file-pilots.json marks them incomplete; none of their partial
  headers is claimed as a completed census or acceptance result.

Pre-push main refresh, final full apply and source identity are recorded in
landing-results.json. Proof archives may be added after an apply without
changing the frozen adapter or pin bytes.

## Main advanced during landing

The first complete measurements used 5ff432f9. The landing refresh found
7050c7075 (container/slot lowering) and b0efce6dc (99s). The branch was moved
to b0efce6dc and all primary apply, oracle, counter, checker and compiler
measurements are repeated on that stack. historical-5ff-summary.json retains
the original outcomes; those earlier raw streams are not mislabeled as the new
main measurements. The five 99u owner pins and executable edit are unchanged.

## Final Go-only advance and focused verification

The next refresh advanced main through 634544791 and a3f946723. Neither changes
TypeScript adaptation inputs, the source pin, API dependencies, oracle or lane
harness. Full apply on that tip is byte-identical to the refreshed oracle tree
over all 746 source files, and direct replay makes zero edits.

The complete ordinary 79-root census remains frozen to b0efce6d. Separate
latest-tip binaries rerun 79-root types, both edited-file ordinary census rows,
all five complete owner observations and the exact single-line lowering mutant.
All use the unmodified current compiler, the existing meter full attempt path
and both no-output guards. The owner findings are unchanged.
current-compiler-verification.json records binary hashes and every result.
The latest lane Go run exits 0 in 80.245 seconds.

No earlier classifier run is mislabeled as a latest-compiler run. The counts
and Node identities are bound to identical TypeScript source bytes; the
current compiler and source scope are explicit in provenance.json.

## Combined 99r/99s/99u stack and workspace recovery

The local commit survived as 4d808b00 on 31d7a2448. Full apply on that parent
passed, as did direct replay with files:0 and edits:0. combined-stack-preparation.json
describes the exact reversible construction of the combined main reference and
accepted instrumented scratch tree. The application adapter and pins did not
change. The default cast remains byte-for-byte unchanged.

The unmodified combined main oracle completed its upstream test process before
the executor disconnected. The wrapper's postprocessing report was absent.
recover-reference.py checks the actual completed log, every raw title/outcome
event, even host/worker multiplicities and every baseline file. The recovered
report explicitly leaves missing wrapper exit/timing metadata unclaimed.

The accepted combined-stack oracle completed normally: install/build exit 0,
tests exit 1 solely for the sanctioned API baseline. It has 106,366 passes,
one failure, no pending tests and baseline_diffs exactly api/typescript.d.ts.
The raw and normalized identity multisets exactly match the recovered combined
main reference. All 11 counter totals match the original pre-build measurement.

The new node-kind side channel records only actual false default evaluations.
All twelve are in the README with the exact active title and canonical numeric
kind. The names are independently verified against stock TypeScript 6.0.3;
the raw numeric kinds are authoritative. There are six MethodSignature and six
TypeReference nodes. This instrumentation is scratch-only and changes no guard,
cast, adaptation site or oracle harness.

The fresh ordinary census executable built at 31d7a2448 is byte-identical to the
focused a3f94672 binary. Relevant lower/load/IR/checker/meter sources are also
unchanged between those commits. combined-compiler-closure.json records that
comparison. Earlier 79-root compiler censuses remain labeled b0efce6d; no older
classifier observation is relabeled as a newer execution.

Combined-stack follow-up on 31d7a244808fa8092643926d21ab9d310a54cce3: `combined-types.json` records exit 0 over all 79 roots before and after. `combined-comparison.json` verifies the only source edits are the planned owners/imports and both public declarations are byte-identical. `combined-contract.json` repeats all 43 contract cases. `combined-files.json` records the fresh ordinary per-file compiler census; both files still stop at the existing core.ts:736:36 unchecked cast. The complete owner replay records are in `combined-owners-before.jsonl.gz` and `combined-owners-after.jsonl.gz`; the compiler closure is checked in `combined-compiler-closure.json`. These are additional measurements, not replacements for the explicitly based historical census.

The final combined source also passes stock TypeScript 6.0.3 on all 79 roots before and after. `combined-stock-kind-pin-mutant.json` records the one expected diagnostic when the Parameter case is removed from hasInferredType. `combined-guard-runtime.json` repeats all 40 expected assertion failures at levels 0 through 3 using the final built compiler and actual existing guards. `combined-sort.json` uses the real adaptation_inputs() implementation. The final pre-push fetch confirms origin/main is still 31d7a244808fa8092643926d21ab9d310a54cce3; no additional rebase is required.

Final combined-stack generate_shards --check and the documented stage3/lane Go fixture command both exit 0; their logs are combined-shards.log and combined-lane.log. All resumed verification completed before the plain push.
