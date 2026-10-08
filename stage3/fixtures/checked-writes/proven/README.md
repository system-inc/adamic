# Writable views covered by static proofs

Basis: adaptation 71's classification at Adamic commit `9b8ebd77`, measured on
adapted TypeScript 6.0.3 source commit `12fef29628bfda6cf8bec4ba2a326fa4ec4a66d5`.
This branch starts from `origin/main` (`45487a809f89885a3fc651cd590e7dabf31362dc`);
it does not merge the adaptation branch or change compiler/source files.

Step 10 (#7hc82dv) requires every one of these **98 sites** to compile with **zero
writable-view runtime checks** once checked writes (#63x2441) lands. These are
91 closed receiver graphs with no writes (class a) and 7 closed graphs whose
writes fit their original holders (class b). They are selected by original IDs,
not deduplicated by location or by the census's first suspect field. For example,
a reason about readonly `pos` does not imply that a modifier write changes `pos`.
The other 517 of the 615 retained observations are outside this bucket.

[sites.json](sites.json) preserves each original selected record, including every
location, original type, expression, read, write, alias, call and unresolved list.
It adds explicit proof obligations, a witness, and the future compile/no-check
expectation. The census JSON SHA-256, adapted source hashes, and literal source
spans pin its basis. The `source_families` table preserves all **38** source-type
families; the ten executable witnesses group them by the receiving proof needed,
rather than repeating a truthiness fixture for every source type.

| Proof family | Sites | Witness | Static obligation |
| --- | ---: | --- | --- |
| Truthiness only | 82 | `01_truthiness.a` | The object is only the left operand of `&&`; no object identity transfers into its result. |
| Discarded condition result | 4 | `02_control.a` | The final FlowNode result is tested by an `if`, then discarded. |
| Kind reader | 3 | `03_kind_reader.a` | A resolved `isPrivateIdentifier` body reads `kind`; no write or escape. |
| Range readers | 2 | `04_range_reader.a` | Follow the resolved emitter helpers; only presence and `end` are read. |
| Diagnostic related information | 2 | `05_diagnostic_related.a` | Follow the forwarding call; initialize `relatedInformation` with `[]` and push compatible information. Do not widen/write `file`. |
| Modifier initialization | 1 | `06_modifiers_initialize.a` | `||=` stores a compatible array; `unshift` inserts a compatible static modifier. |
| Assertion clause | 1 | `07_assert_clause.a` | ImportAttributes fits the original AssertClause domain. |
| Property initializer | 1 | `08_property_initializer.a` | Parser recovery writes Expression or undefined into that exact domain. |
| Static block modifiers | 1 | `09_static_modifiers.a` | The modifier array or undefined fits the original domain. |
| Ambient module parent | 1 | `10_module_parent.a` | AmbientModuleDeclaration fits both original Node parent domains. |

The fixture excerpts retain the actual functions or statements verbatim. Small
surrounding interfaces expose only the fields needed to run each case. Factory,
parser, annotation and line-count dependencies are explicit deterministic driver
stand-ins, not claims to reproduce those helpers' entire implementation. The
range fixture keeps the full closing-line function and exercises presence/end
reads, with a stand-in line-difference helper. The diagnostic, kind-test, accessor
and static-block functions are retained whole. The other fixtures retain the
original receiving statements in a driver scope. Source coordinates name the
pinned **adapted** compiler tree, not pristine line numbers.

## Observations and limits

[status.json](status.json) records source Node stdout/stderr/exit and the exact
current-main stage0 diagnostic for each witness. All ten run on Node. Current
Adamic returns **2 NotYet and 8 Refused**, so no native binary was produced and no
native-output equality is claimed. Unrelated current limitations include value
`&&`, `||=`, a predicate proof and checked non-null assertions. No fixture was
rewritten to conceal those blockers. The compile/no-check entries in sites.json
are the ruling's requirements, not observations of an implementation that has
not landed. These scout fixtures are not added to the central native oracle;
[counts.md](counts.md) records this bucket's measurements.

The original audit is conservative about aliases, virtual/external calls,
callbacks, generic bounds and heap escapes. Its conclusions assume ordinary
declared properties and its modeled library operations. This unit verifies the
pinned classification and representative receiver behavior; it does not claim
runtime reachability in all compiler projects or a new whole-program proof.

## Checks and mutants

`receiving-audit.cjs` is the byte-identical audit from `9b8ebd77`; its SHA-256 is
checked by verify.cjs. Stock TypeScript **6.0.3** supplies the independent type
relations. `verify.cjs` compares all 98 rows to the original JSON in Git, checks
every expectation, source-type and proof-family count, and checks each witness
receiving graph. Each witness has zero stock-TypeScript diagnostics. With
`PROVEN_SOURCE_TREE`, it also verifies adapted source hashes and excerpt spans.

Its three child mutants must exit 1 specifically at their assertions:

- `--bad-write` adds `node.emitNode = undefined` to the real kind-test function
  in the standalone witness. The audit first establishes c/checked and then
  fails the original a/proven assertion; TypeScript still accepts the program.
- `--drop-row` removes one original ID. The exact pinned-identity comparison
  fails, so matching totals alone cannot hide a dropped site.
- `--node-output` changes the driver autoGenerate input from 7 to 8. The receiver
  remains class a, but the source Node stdout byte comparison fails.

[real-mutant.json](real-mutant.json) goes further: real-mutant.cjs creates two
compiler programs over the **whole pinned adapted src tree**, changing only the
real nodeTests.ts receiver through an in-memory host. Original site **420**
(emitter.ts:5461:90) moves **proven -> checked** because undefined does not fit
its original `EmitNode & { autoGenerate: AutoGenerateInfo }` field in **both**
GeneratedIdentifier and GeneratedPrivateIdentifier. The pinned audit conservatively
returns d for the union; this runner separately uses stock TypeScript's assignability
relation to prove rejection by each original member before routing to checked.
It does not modify the audit or treat an unresolved graph alone as proof. The receiver's
stock-TypeScript diagnostics are unchanged. `--expect-proven` then exits 1 at the
original proven-site contract. This tests routing, not a native runtime check
which remains pending #63x2441. Source files on disk are unchanged.

Reproduce from the repository root, after fetching the pinned classification ref:

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules"
PROVEN_SOURCE_TREE=/tmp/step10-proven-tree node stage3/fixtures/checked-writes/proven/verify.cjs > /tmp/step10-verify.log 2>&1
python3 stage3/fixtures/checked-writes/proven/observe.py > /tmp/step10-observe.log 2>&1
node stage3/fixtures/checked-writes/proven/real-mutant.cjs /tmp/step10-proven-tree --expect-proven > /tmp/step10-real-mutant.log 2>&1
```

The last command must exit 1 with `original proven-site contract rejects the real
out-of-type write`; an incidental runner/build failure does not count. Rebuild the
adapted tree with the apply.sh at `9b8ebd77` in a detached scratch worktree. Setup
reported 8.544 seconds, Go 1.27.1, clang 20.1.8, Node 24.19.0, `nproc` 5 (quota 4).
Apply and every test wrote a log; no full package or full gate ran.

## Checking the compiler implementation

After checked writes lands, the compiler worker should measure every original
site on this pinned tree, including all aliases and writes in its graph. Supply
`PROVEN_COMPILER_RESULTS=/path/to/observations.json` to verify.cjs. Its format is:

```json
{
  "adamic_sha": "<40 lowercase hex digits identifying the measured compiler>",
  "adapted_source_commit": "12fef29628bfda6cf8bec4ba2a326fa4ec4a66d5",
  "records": [
    {"id": 30, "where": "src/compiler/binder.ts:3288:21", "compiles": true, "writable_view_runtime_checks": 0}
  ]
}
```

All 98 ordered IDs/locations are mandatory; the one-row example is incomplete
and fails. Counts must come from the compiler's actual plan/emission evidence,
not copies of expected values. Only writable-view checks are counted here;
ordinary bounds, unwrap and other checks remain separate obligations. This unit
provides the complete assertion contract; implementing the compiler observation
producer is pending the checked-writes compiler interface.

The observation-consumer controls also ran:

```sh
node stage3/fixtures/checked-writes/proven/compiler-contract-controls.cjs > /tmp/step10-compiler-controls.log 2>&1
```

A complete synthetic 98-row stand-in passes. An inserted writable-view check,
a refused site, and a dropped site each fail their respective assertion (exit 1).
[compiler-contract-controls.json](compiler-contract-controls.json) labels these
synthetic controls explicitly; they are not compiler implementation evidence.
