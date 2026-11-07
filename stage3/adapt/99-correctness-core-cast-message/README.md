# cast keeps its value and drops only the predicate name

Complete, JavaScript-changing. Adamic task b5w3ycg, Nov 7 22:06:52Z,
Compiler accepts under the delegation; the user subsequently revised the
mutant proof to run on head 2. Base: 2a95448b8fadbcd30549a148bd28da5f8b75f907.
The published stopped checkpoint is retained as history, with a normal
continuation commit. No force-push is used.

The chosen wording is `Invalid cast. The supplied value ${value} did not pass the test.`
Only the space, quoted predicate name and Debug.getFunctionName call disappear.
The period stays. `${value}` belongs to #09dvwny and is unchanged. The
condition, return, Debug.fail call and its throw implementation stay byte for
byte. Current source: core.ts:1789; the earlier forecast called it 1791 before
the elementAt declaration cut removed two preceding lines.

## Ownership and adapter

No earlier active whole-owner or exact-message pin covers cast. The audit of
224 adapter/pin inputs includes 99q, 99r and the earlier core adaptations.
40 and 41's old whole-cast text is historical evidence, not an active guard.
The new directory pins the complete before/after owner, refuses drift and
performs zero writes on an already-adapted tree. No earlier adapter changes
in this head. Debug and getFunctionName are byte-identical to the baseline.

## Counts and witnesses

| Measurement | Real executions | Firing control |
| --- | ---: | ---: |
| Original getFunctionName expression in cast's failure message, cited fcb70da5e | 0 | 1 |
| Cast failure branch on this exact cut, fresh whole lane | 0 | 1 |

This does not claim cast itself is never called. Its failure branch is zero.
The fresh lane covers all 398 units, 106366 passes, one expected failure and
zero skips, with all identities and baseline guards preserved. The full raw
pieces and independent reconciliation live in head 2's move directory;
whole-lane-pair-reconciliation.json here retains the checked result.

The actual completed normal builds are used by normal-cast-witness.cjs.
Six failing value classes preserve Error class, condition evaluation count,
value interpolation and exact throw point. Four passing cases retain identity;
a predicate-thrown sentinel retains its identity. Only the shorter text differs.

## Revised mutant, from head 2

cast is an uninstantiated generic in the core census. Its before/after core
censuses have identical site sets, no added site, and zero diagnostics. No
lowering success is inferred from deferral. As the user ruled, the mutant is
instead run on the finished moved tree: restore the original complete cast
owner while Debug no longer exports getFunctionName.

Stock tsc exits 2 and adamic types over all 79 roots exits 1. Each reports only
TS2339 at core.ts:1789:97: Property 'getFunctionName' does not exist on type
of Debug. This is the earlier 1791 site after the elementAt cut. The source
restoration hash also matches the complete original core.ts bytes.

Citation: typescript/core-getfunctionname-move,
stage3/adapt/99-correctness-core-getfunctionname-move/cast-restoration-mutant.json.
The exact result is copied here and pinned by SHA256 `f1d5fc34c58b5247c0cbaf4974b74f01a666758bac79b0c6071a156f687b2343`.
This replaces the superseded absent-core-site mutant requirement.

## Proof results

- Both full applies: exit 0, 542.291s baseline and 653.843s head 1.
- Post-apply checkpoints: pushed after each apply, recorded in PROOF.json.
- Paired replay: all 97 adapters; only cast's transformation differs.
- Adapter 49 rewrites core.ts unchanged: same exit, output, paths and four
  writes. Its raw no-op hashes inherit exactly our message difference. Both
  writes are identity operations; the raw hashes and proof remain visible.
- Both 79-root adamic types checks: exit 0, zero diagnostics.
- Both own compiler/harness builds: exit 0, zero diagnostics.
- Every JavaScript hunk: only one message deletion in each of _tsc.js,
  typescript.js and run.js; all other emitted JavaScript byte-identical.
- typescript.d.ts and tsserverlibrary.d.ts: byte-identical.
- Core full-mode census: identical sites, no additions, cast deferred.
- Actual normal-build Node witness: pass, as described above.
- Revised missing-member mutant on head 2: rejected by stock tsc and Adamic.
- getFunctionName and all other source files: unchanged.
- Ledger: records the ruling and the head-2 mutant proof.

See PROOF.json, all-emitted-javascript-hunks.patch, source-hunks.patch and the
raw archives. The original STOP.json and raw-stopped-proof.tar.gz are marked
historical. The old -9 applies were manually stopped, not killed by the
2-hour limit or an observed memory failure. Fresh applies all pass. Peak
child RSS is recorded; it is not represented as an aggregate process-tree peak.
