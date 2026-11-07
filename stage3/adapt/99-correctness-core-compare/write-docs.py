import json,sys
from pathlib import Path
s=Path('/workspace/scratch/core-compare');repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-core-compare';count=json.loads((s/'RECONCILED.json').read_text());control=json.loads((s/'CONTROL.json').read_text());assert count['complete'] and control['complete'];final=len(sys.argv)>1 and sys.argv[1]=='final'
comparison=json.loads((s/('final-comparison.log' if final else 'comparison.log')).read_text());census=json.loads((s/'CENSUS.json').read_text());rows=[]
for pair in ['number/number','string/string','undefined/undefined','string/undefined','undefined/string','number/undefined','undefined/number']:
 rows.append('| '+pair.replace('/',' / ')+' | '+format(count['counts'].get(pair,0),',')+' |')
rows.append('| Mixed string and number | 0 |')
changed=[row['name'] for row in comparison['emitted'] if not row['identical']]
text=f'''# compareComparableValues compares like with like

JavaScript-changing, count-authorized established-shape adaptation for task b5w3ycg, wave row L1, site F18 at adapted core.ts:2003:9. Upstream is TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8. Base is origin/main f0271c6a7b2e92900eb7046d8246cfb89f52cfb6.

Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts.

The two overloads and the first three conditional arms are unchanged. The final comparison splits on both operands' typeof: numbers use their own `<`, and strings use their own `<`. Source line count is preserved. The third arm returns `Comparison.GreaterThan` for a mixed string/number pair. The overloads forbid that pair, so a well-typed tsc caller cannot reach it. The complete lane count found 0 mixed calls. Deliberate mixed Node witnesses show the old coercing comparison can return LessThan while the new arm returns GreaterThan; this off-count class has count 0 and is recorded in the ledger.

Equal operands, undefined on either side, ordinary numbers, strings with different case, and NaN have equal before/after witnesses. NaN remains GreaterThan, including NaN compared with itself. Every observed operand-type pair is witnessed. No new cast, import, lib change, compiler overlay for ordinary checks, or production lowering change is introduced.

## Count before construction

All 398 unchanged lane units pass their normal identity, baseline, API and count guards: {control['counts']['pass']:,} passing tests, {control['counts']['fail']} sanctioned failure, 0 skipped. The separate unchanged `stage3/lane/run.sh` control passes, with the identical test-identity multiset, counts and raw baseline diff. Counted calls total **{count['calls']:,}**, with **0 mixed calls and 0 result differences**. No adapter was constructed until the count, independent reconciliation and whole-lane control completed.

| typeof pair | Calls |
| --- | ---: |
'''+ '\n'.join(rows)+f'''

Each call computes the original and candidate bodies on the same operands, records both results, and returns the original result to tsc. The scratch observer is absent from adapt.cjs. The frozen source/built manifest and recorder hash bind the measurement. A byte records four base-3 tags: typeof a, typeof b, original Comparison, adapted Comparison. BINARY-SCHEMA.json describes the lossless encoding of these required fields. Every call is written synchronously, including calls before an intentional process termination. Mixed pairs or differences also retain full operand values in a sidecar and stop the count.

The compact recorder matches the original JSON observer on all 676 excluded seed calls. Earlier JSON attempts hit unchanged timing guards and are excluded and archived. An additional real-call comparison verifies all 16,586,589 ordered JSON records from the first passing JSON piece against the original observer. The accepted census uses the single frozen compact recorder throughout. Piece 1 used ordinary files; piece 2 retains 19 passing ordinary-file units and the isolated memory-backed retry of reallyLargeFile.ts, which passed in 18.82 seconds. Later pieces use two concurrent units with synchronous memory-backed files copied into the archive after each unit. The runner's 90-second unit bound and upstream 40-second test bound are unchanged. Seeds and failed attempts are excluded from the canonical 398-unit union.

RECONCILED.json reopens every archive, verifies every raw member hash, reconstructs the type/result matrix and validates every report against the unchanged lane manifest. COUNT.json retains piece provenance and the isolated retry. CONTROL.json records the whole-lane match.

## Proof

| Check | Result |
| --- | --- |
| Main full apply | PASS |
| Main plus compare full apply | {'PASS' if final else 'Pending final fresh apply'} |
| Paired replay of every adapter | PASS; all other outcomes identical, compare applies once before and performs zero edits after |
| Owner contract | PASS; exact full overload owner pins, no-write drift refusals, LF/CRLF and foreign edits preserved, replay bytes and mtime preserved |
| Adamic types | PASS; all 79 roots, 0 diagnostics before and after |
| Source scope | All 746 source files checked; only the pinned owner differs |
| Emitted JavaScript | All 9 files checked; only {', '.join(changed)} differ, only at the owner |
| Public declarations | typescript.d.ts and tsserverlibrary.d.ts byte-identical |
| Full core.ts census | F18 removed; {census['beforeDistinctSites']} to {census['afterDistinctSites']} distinct sites; no new site anywhere; all 3 overload probes attempted |
| Original-owner mutant | Restores the entire baseline site set, including F18 |

JAVASCRIPT-DIFF.md shows every JavaScript hunk. The full census uses the same LATENT_FULL, LATENT_ONLY_FILE, LATENT_ASSERT_NO_OUTPUT and timing procedure as core-sites-4186581e. Raw findings, timings, manifests, both type logs, all replay outcomes, exact diffs and tool provenance are archived. Every retained Go source is named .go.txt. Evidence-only file updates can change the product cache key without changing the frozen transformation.

The initial ownership scan found no earlier owner pin. Matches in axolotl-j are checker.ts call-site edits. The adapter pins the entire three-declaration owner, refuses missing, duplicate, nested, drifted or half-adapted owners before a write, and edits nothing outside that owner. Replay performs zero writes.
'''
(out/'README.md').write_text(text)
(out/'PROOF.md').write_text('''# Reproduction and evidence

Source /workspace/adamic-tools/env.sh, run bash cloud/setup.sh and npm ci --prefix stage3/api, and export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source at 050880ce59e30b356b686bd3144efe24f875ebc8. The API parser is stock TypeScript 6.0.3.

The retained scripts use /workspace/adamic and /workspace/scratch/core-compare. prepare.sh records the baseline full apply; instrument.cjs changes only the scratch owner to compute both results while returning the original. count.py runs every lane unit through the unchanged shard.py. reconcile.py independently decodes the canonical archives. control.py runs the unchanged whole lane and compares exact identities, counts and baseline bytes. construct.cjs refuses to construct the adapter until all three gates pass.

post-measure.py runs contracts, actual-source Node witnesses, both ordinary compiler builds, owner-scoped JavaScript/declaration comparison, the after and mutant full censuses, and paired replay. close-proof.py then performs a fresh full apply of the complete main-plus-compare stack, checks all 746 source hashes against the independently proved tree, runs all 79 types roots and the full core census on that exact output, builds it with the pinned packages, and verifies its emitted bytes. Ordinary checks use the unmodified main Adamic binary. Only the full census binary has the reference measurement-only overlay; production Lower is disabled in that binary and both no-output assertions remain enabled.

The setup and early preparation corrections are retained: a slow full-history source clone was replaced by a shallow exact-commit fetch; the scratch types.py name was corrected to check-types.py; the observer was corrected to preserve CRLF. No call from a failed attempt or an excluded seed contributes to COUNT.json.

Canonical piece archives contain raw one-byte calls, decoded histograms, every unit report and observation, baseline diffs and SHA256SUMS. The evidence archive retains complete proof logs and scripts, the excluded seeds, raw censuses, exact emitted hunks, source manifests, owner witnesses, tool hashes and command results. Excluded JSON and packed timing attempts are separate archives. Root SHA256SUMS covers every retained file except itself.
''')
line=f'compareComparableValues, 99-correctness-core-compare: Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts. JavaScript-changing typeof split; whole lane {count["calls"]} calls, 0 mixed string/number calls, 0 result differences. The overload-forbidden mixed arm returns Comparison.GreaterThan; mixed witnesses can differ and their real-call count is 0.'
ledger=repo/'stage3/upstream/LEDGER.md';old=ledger.read_text()
if line not in old:
 assert '99-correctness-core-compare' not in old
 ledger.write_text(old.rstrip('\n')+'\n\n'+line+'\n')
