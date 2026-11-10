from pathlib import Path
import json,hashlib,shutil,tarfile,collections,re
s=Path('/workspace/scratch/each-node');repo=Path('/workspace/adamic');unit=repo/'stage3/adapt/99-correctness-core-each-node';base=json.loads((s/'base.json').read_text())['base'];lane=json.loads((s/'reconciled.json').read_text());build=json.loads((s/'build-comparison.json').read_text());static=json.loads((s/'static-census.json').read_text());assert lane['passed'] and build['passed'];assert lane['totals']=={'defined':1086412,'exposed':0,'throws':0};assert len(static['functions'])==26 and not static['unresolved'] and not any(f['canSeeSecond'] for f in static['functions'])
assert (s/'mutant/src/compiler/debug.ts').read_bytes()==(s/'census-mutant/src/compiler/debug.ts').read_bytes()==(s/'before/src/compiler/debug.ts').read_bytes()
summary={'compilerBase':base,'rule1SourceOnMain':(repo/'internal/lower/overload_callback_use.go').exists(),'runs':{}}
for side in ['mutant','after']:
 rows=[json.loads(l) for l in (s/(side+'-census.jsonl')).read_text().splitlines()];header=rows[0];assert not header['diagnostics'] and header['latent_mode']=='full';records=[r for r in rows if r.get('units')];assert len(records)==1;record=records[0];targets=[f for f in record['findings'] if 'assertEachNode' in f['reason']];assert targets and all(f['reason']=='overload 1 of assertEachNode parameter test cannot be served by implementation parameter test' for f in targets)
 summary['runs'][side]={'exit':json.loads((s/(side+'-census-result.json')).read_text())['exit'],'units':len(record['units']),'findings':len(record['findings']),'checkerDiagnostics':0,'assertEachNode':targets}
assert not summary['rule1SourceOnMain'];summary['limitation']='Both forms meet the existing overload-1 callback contract on this main. No removal/reintroduction delta or for-of-specific stop is claimed.';(s/'census-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
u=json.loads((s/'unused-import.json').read_text());removed,retained=u['cases'];delta=collections.Counter((d['code'],d['text']) for d in retained['diagnostics'])-collections.Counter((d['code'],d['text']) for d in removed['diagnostics']);assert delta=={(6133,"'every' is declared but its value is never read."):1};(s/'unused-import-summary.json').write_text(json.dumps({'configuredNoUnusedLocals':u['configuredNoUnusedLocals'],'probeEnabled':True,'existingDiagnostics':len(removed['diagnostics']),'additionalDiagnostic':"TS6133: 'every' is declared but its value is never read.",'deltaExactlyOne':True},indent=2)+'\n')
for folder in ['built-control-counts','after-control-counts']:
 rows=[json.loads(p.read_text()) for p in (s/folder).glob('*.json')];assert len(rows)==1 and (rows[0]['defined'],rows[0]['exposed'],rows[0]['throws'])==(1,1,0)
for name in ['source-hunks.patch','restoring-mutant.patch']:shutil.copy2(s/name,unit/name)
with tarfile.open(s/'after-javascript.tar.gz','w:gz') as tf:
 for p in sorted((s/'after/built/local').rglob('*')):
  if p.is_file() and (p.suffix=='.js' or p.name in ['typescript.d.ts','tsserverlibrary.d.ts']):tf.add(p,arcname=str(p.relative_to(s/'after')))
 tf.add(s/'after/src/compiler/debug.ts',arcname='src/compiler/debug.ts')
raw='https://github.com/system-inc/adamic/tree/typescript/each-node-evidence/each-node-evidence'
readme=f'''# assertEachNode's direct loop

Task 8n7eet3, site 2. Adamic Nov 10 22:01Z, Compiler accepts under the delegation.
Base origin/main: `{base}`.
Pinned TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.

JavaScript-changing. Inside the existing Normal assertion gate, the exact ruled
loop computes `passed`, then calls the existing assert. The message, lazy
callback and stack mark remain byte for byte. The callback receives only the
node, so every's index is no longer supplied. A throw from test has no every
frame. No upstream issue was filed.

The whole-lane admission count is **0**, over **1,086,412** assertEachNode calls
with test defined, with **0** throws from test. Both built two-parameter controls
count **1**, including after the index becomes undefined. All 398 units ran once;
106366 passes, 1 expected failure, 0 skips. The static census has 26 predicates,
all declaring `(node: Node)`, and six one-argument forwarding parameters.

## Pins and stacking

The before owner is exactly road (b)'s assertEachNode after pin, CRLF included.
The after owner is the ruled `for (const n of nodes)` form with both presence
checks. `every` has no other source use, so its namespace import is removed.
The owner and import must be in the same state; drift and half adaptation refuse
before writing. A completed tree is idempotent.

Road (b) is amended in place to recognize only this exact finished owner as an
after state. Its old pins, message edits and helper guards remain unchanged.
A changed loop and a mixed message state still refuse. 76's key name,
99-cli-debug-assertion-level's gate call, and 30's indexed-read sites stay.
The final full apply is the acceptance check for their composition.

With `noUnusedLocals` explicitly enabled, retaining every adds TS6133. The pinned
normal build configuration itself leaves that option unset; enabling it also
reports 60 existing diagnostics from earlier adaptations. This change removes
only the newly unused every import.

## Proof and compiler handoff

[PROOF.md](PROOF.md) lists every emitted hunk, the complete static inventory,
Node witness, guard checks and reproduction. [witness-result.json](witness-result.json)
retains 32 paired cases plus four index cases. Raw records, counts, logs and
census streams live on the [evidence branch]({raw}), outside adaptation keys.

The current main compiler still refuses overload 1's test contract in **both**
forms. It does not yet contain Optimizer's rule-1 implementation. Consequently
these main censuses do not establish a removal/reintroduction delta or reach a
for-of-specific stop. No successful Go lowering is claimed. The adaptation is
prepared for the paired Optimizer change; this limitation is explicit in PROOF.
'''
(unit/'README.md').write_text(readme)
rows='\n'.join('| `'+f['name']+'` | `'+', '.join(f['parameters'])+'` | `'+f['where']+'` |' for f in sorted(static['functions'],key=lambda f:f['name']))
hunks=(s/'emitted-hunks.patch').read_text();proof=f'''# assertEachNode proof and outstanding compiler handoff

Task 8n7eet3, site 2, Adamic Nov 10 22:01Z.
Final base: `{base}`. Initial preparation: `e24f2600a5d8a8521c5128827cad395d47c81c81`.
The audited fast-forward changes no stage3 input. Pinned TypeScript:
`050880ce59e30b356b686bd3144efe24f875ebc8`; Node v24.19.0; Go 1.27.1.

**Compiler limitation:** base main lacks `internal/lower/overload_callback_use.go`.
The original and loop forms both stop at the existing overload-1 callback
contract. The requested causal restoring delta and predicted for-of binding stop
cannot be demonstrated on this main. They are not claimed as passing. The source
adaptation, Node behavior checks and JavaScript scope checks are complete; the
backend change must be paired with Optimizer.

## 1. Whole-lane admission count

`stage3/lane/prepare_shards.py` completed apply, locked install, normal build and
harness preparation, child exit 0, in 1126.257 seconds. Then every generated
TestShard_ inventory unit ran through the ordinary shard.py path, as road (b)'s
section 1 does. Reconciliation reopens the raw process rows and ordinary guard
reports and proves a bijection with all generated shard functions.

| Observation | Result |
| --- | ---: |
| Accepted units, exactly once | 398 |
| Node process records | 797 |
| Pass / expected failure / skip | 106366 / 1 / 0 |
| assertEachNode calls with test defined | 1,086,412 |
| Calls whose test source can see a second argument | 0 |
| Throws from inside test | 0 |
| Distinct callback sources | 26 |
| `(n, i)` positive control, before / after | 1 / 1 |
| Throw control | 1, identical thrown object |

Private copies instrument the entry of each assertEachNode in _tsc.js,
typescript.js and run.js. Every JavaScript wrapper is copied, preventing Node's
realpath resolution from escaping to an uninstrumented bundle. The preload
propagates to child processes and VM contexts and writes explicit zero rows.
It records Function.prototype.toString source and parses parameters, defaults,
rest and arguments reads; it never uses Function.length. Every observed source
has one plain node parameter and no arguments read. The wrapper counts throws
and rethrows the same object. Normal build outputs are not instrumented.

All ordinary identity, baseline, API and test-count guards pass. The one expected
public-API failure remains. No candidate whole-lane run is claimed. The before
and after positive controls are separate from the lane's admission count.
Raw reconciliation, 20 raw shard archives and all 797 process rows are on the
[evidence branch]({raw}); their hashes are indexed in ARCHIVES.json and that
branch's SHA256SUMS.

## 2. Static census across all src

The checker-backed scan covers 291 relevant calls, 26 concrete predicates,
six forwarding parameters, and zero unresolved expressions. It follows
visitNodes, visitArray and visitArrayWorker and NodesVisitor calls, including the
three custom forwarding wrappers. Every concrete declaration is:

| Predicate | Declared parameters | Pinned source location |
| --- | --- | --- |
{rows}

The six forwarded test parameters are visitNodes.test, visitArray.test,
visitArrayWorker.test, checker.ts's deepCloneOrReuseNodes.test,
expressionToTypeNode.ts's visitNodesWithoutCopyingPositions.test, and
textChanges.ts's assignPositionsToNodeArray.test. Each declares
`((node: Node) => boolean) | undefined`; inbound NodesVisitor calls resolve to
the same inventory. Full call sites and declarations are in static-census.json.

## 3. Exact source, witness and guards

Only compiler/debug.ts differs among all 747 src files. The before pin is exactly
road (b)'s after pin. The gate call and key name are unchanged; the one assert
keeps its message, lazy callback and stack mark byte for byte. Source line endings
remain CRLF. Both declaration-emission modes match. every's only non-import
reference was the replaced call, so its import is removed.

The completed before/after debug.ts texts run at assertion levels 0, 1, 2 and 3.
All 32 paired cases pass: undefined nodes, empty array, a factory-created
NodeArray, all passing, first failure, middle failure, undefined test, and a throw
inside test. Recorded call sequences stop at the first failure. Error constructor,
message and actual throw point agree. A test throw loses its every stack frame.
Four additional two-parameter cases show index 0/1 before and undefined after.
Both built-artifact source counters still classify that test once.

The own adapter is idempotent, and road (b) recognizes both completed states.
Five mutants refuse: before body without import, after body with import, drifted
loop, road (b)'s drifted finished loop, and loop form with another message restored.
The explicit road (b) drift refusal is `unreviewed owner: assertEachNode`.
No rejection writes a file. Guard receipts are in guards-result.json.

Stock TypeScript with noUnusedLocals enabled adds exactly one diagnostic when
only every's import is restored: `TS6133: 'every' is declared but its value is
never read.` Removing it clears that added diagnostic. The actual pinned config
leaves noUnusedLocals unset; the explicit probe also exposes 60 pre-existing
unused diagnostics, retained raw. No claim is made that the normal build enables
that option. The normal candidate npm run build exits 0.

## 4. Main full census and exact restoring mutant

The unchanged full meter is rebuilt from final base main. Production Load/Lower
output guards remain active. Physical generated Go evidence names end in .go.txt.
Each run uses LATENT_FULL=1, LATENT_ONLY_FILE=<tree>/src/compiler/debug.ts,
LATENT_ASSERT_NO_OUTPUT=1 and a timings path, after clearing inherited LATENT flags.

The restoring source is an isolated candidate src copy with the entire original
owner and its required every import restored. Its debug.ts equals before byte for
byte; its patch is restoring-mutant.patch. The separately retained census-mutant
copy is byte-identical to that source.

| Run | Child exit | Units | Findings | Checker diagnostics | Observed stop |
| --- | ---: | ---: | ---: | ---: | --- |
| Restoring mutant | 0 | 81 | 22 | 0 | debug.ts:257:86, overload 1 test contract |
| Loop on main | 0 | 81 | 22 | 0 | debug.ts:256:86, overload 1 test contract |

Exact reason in both streams:

> overload 1 of assertEachNode parameter test cannot be served by implementation parameter test

It appears in two recording units in each run. Main short-circuits at overload 1;
there is no measured evidence for overloads 2 through 4, nor a for-of element-type
stop on this base. The old refusal is present in the restored source, but it never
left on main, so this is not presented as a successful removal/reintroduction
experiment. Optimizer's rule 1 and for-of extension remain the required handoff.

The legacy `measured on a checker-rejected program` label is retained verbatim;
both headers have an empty diagnostics array. No usable compiler output is emitted.
Raw streams, timings, command flags, binary hash and compiler base are external.

## 5. Every emitted JavaScript hunk and runtime exports

All 10 normal JavaScript files were compared. Exactly three differ, once each,
inside assertEachNode. After masking only that body, every remaining byte is
identical; identifier-token normalization also matches. There are **0** outside
hunks and **0** esbuild renames to excuse. The seven other JavaScript files are
byte-identical. Both public declaration files are byte-identical.
All **2256** runtime ts export keys match; sorted key SHA256:
`{build['exportKeys']['sha256']}`.

The initial symbol-binding comparison hit Node's default stack limit. Its larger
stack retry was stopped as unnecessary after inspecting the three body-only
hunks. The final comparison uses the stronger outside-byte check plus identifier
normalization and exits 0. All original child exits are retained as preliminary
measurement history, not passing results.

Every emitted hunk follows:

```diff
{hunks}```

## 6. Final composition acceptance and reproduction

The final acceptance runs `bash stage3/apply.sh <fresh-final-tree>` after this
directory's SHA256SUMS is frozen. The fresh tree is normally built and its entire
src inventory, JavaScript set, public declarations and export keys are compared
to this measured candidate. Its child-exit receipt and final verification are
retained externally as final-apply-result.json, final-build-result.json and
final-verify.json. Publishing is conditioned on those checks passing. Keeping
those logs external avoids changing the key of the product being verified.

Setup: source /workspace/adamic-tools/env.sh; bash cloud/setup.sh;
npm ci --prefix stage3/api; locked stage3/typescript dependencies; export
ADAMIC_TYPESCRIPT_SOURCE to the pinned 050880ce checkout. Evidence controllers
retain exact commands, paths and every child's own exit. prepare.py and lane.py
produce the baseline; reconcile.py independently reopens it. static.cjs records
the source census. guards.cjs and witness.cjs run against completed trees with
NODE_PATH=stage3/api/node_modules. proofs.py, resume-candidate-final.py,
compare-built.cjs and unused-import.cjs retain the normal candidate checks.
build-census.py and census-run.py produce the guarded main streams. The witness
and exact source/restore patches are retained here; raw counts, logs, census
streams, JavaScript archives and controllers are only on the evidence branch.

LEDGER.md and rulings.md record the JavaScript-changing adaptation and its
zero-count condition. No upstream issue was filed. No Go lowering success or
for-of-specific main stop is claimed.
'''
(unit/'PROOF.md').write_text(proof)
ledger=repo/'stage3/upstream/LEDGER.md';line="Adamic, task 8n7eet3, Nov 10 22:01Z, site 2: 99-correctness-core-each-node is JavaScript-changing. assertEachNode's test no longer receives every's index, and a throw from test has no every frame. The exact ruled direct const for-of loop stays inside the existing Normal gate; message, lazy callback and stack mark stay byte for byte. Whole-lane second-argument exposure count 0 over 1,086,412 calls with test defined, 398 units and 797 processes; throws from test 0; two-parameter control 1 before and 1 after. Road (b) recognizes only the exact finished owner and still refuses drift. Main still meets the overload-1 callback contract in both forms; the Optimizer pairing remains required. No upstream issue filed."
assert 'task 8n7eet3' not in ledger.read_text();ledger.write_text(ledger.read_text().rstrip()+'\n\n'+line+'\n')
ruling="- **tsc's assertEachNode replaces every(nodes, test) with the ruled direct const for-of loop, inside its existing Normal gate, only at a whole-lane count of 0 for defined tests that can observe a second argument. The test receives no index, and a throw from test has no every frame; the existing assert message, lazy callback and stack mark stay byte for byte. Drop the unused every import. Road (b) recognizes only the exact finished owner and still refuses drift.** Meaning (JavaScript-changing adaptation). Nov 10 2026 22:01Z, @system_adamic; source: #8n7eet3, site 2, Compiler accepts under Adamic's delegation. Measured: 0 exposures over 1,086,412 defined-test calls, 0 test throws; controls 1 before and after; 398 lane units. Node witnesses and complete source census are in stage3/adapt/99-correctness-core-each-node/PROOF.md. Main still refuses overload 1 in both forms until the Optimizer pairing; no for-of-specific stop or restoring delta is claimed. No upstream issue filed."
p=repo/'docs/rulings.md';lines=p.read_text().splitlines();at=next(i for i,l in enumerate(lines) if l.startswith("- **tsc's lazy assertion failure messages in debug.ts"));assert not any('#8n7eet3, site 2' in l for l in lines);lines.insert(at+1,ruling);p.write_text('\n'.join(lines)+'\n')
artifacts=[*sorted(s.glob('lane-raw-*.tar.gz')),s/'before-javascript.tar.gz',s/'after-javascript.tar.gz',*[s/n for n in ['reconciled.json','lane-summary.json','static-census.json','guards-result.json','census-summary.json','mutant-census.jsonl','after-census.jsonl','build-comparison.json','unused-import-summary.json','emitted-hunks.patch','main-advance-audit.json']]]
manifest={'branch':'typescript/each-node-evidence','directory':'each-node-evidence','artifacts':[{'file':p.name,'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in artifacts]};(unit/'ARCHIVES.json').write_text(json.dumps(manifest,indent=2)+'\n')
assert all('\u2014' not in p.read_text() and '\u2013' not in p.read_text() for p in unit.iterdir() if p.is_file())
(unit/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(unit.iterdir()) if p.is_file() and p.name!='SHA256SUMS'))
(s/'frozen-adaptation-manifest.json').write_text(json.dumps({str(p.relative_to(repo)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((repo/'stage3/adapt').rglob('*')) if p.is_file()},indent=2)+'\n')
print('frozen candidate, explicit census limitation, ledger and ruling; base',base)
