from pathlib import Path
import hashlib,json,shutil,sys
out=Path(__file__).resolve().parent
rows=[json.loads(s) for s in (out/'commands.jsonl').read_text().splitlines()];last={r['name']:r for r in rows}
required=['resumed-baseline-apply','amended-full-apply','red-census','green-census','census-pair-final','owner-contract','legacy-contract','node-witness-every-flush','current-main-compiler-build','house-types','normal-whole-lane','source-emission-comparison','real-bundle-comparison','final-generated-shards-check','final-shard-plan-regeneration','final-go-shard-regeneration','counting-oracle-verification','counting-totals','full-apply-matches-census','lane-matches-census']
for name in required:assert last[name]['exit']==0,(name,last.get(name))
assert last['counting-oracle']['exit']==1
counts=json.loads((out/'counts.json').read_text());lane=Path('/workspace/scratch/sourcemap-normal-lane');report=json.loads((lane/'report.json').read_text())
assert report['status']=='pass',report
saved=out/'normal-lane';saved.mkdir(exist_ok=True)
for name in ['execution.json','report.json','verdict.json','verdict.txt','failures.md','patch-set.md','apply.log']:
 p=lane/name
 if p.exists():shutil.copyfile(p,saved/(name+'.txt' if name.endswith('.log') else name))
(o:=saved/'oracle').mkdir(exist_ok=True)
for name in ['report.json','baseline.diff','install.log','build.log','tests.log','failures.md']:
 p=lane/'oracle'/name
 if p.exists():shutil.copyfile(p,o/(name+'.txt' if name.endswith('.log') else name))
(out/'results.json').write_text(json.dumps({'amendmentCommit':sys.argv[1],'base':'1a1c0dc18e6546b5861ee56352c542194471ed70','workers':'e8a1529e6f6addd4ae8dc28ee07c217bcb7b0f0e','requiredReceipts':{name:last[name] for name in required},'counts':counts,'lane':report},indent=2)+'\n')
(out/'REPORT.md').write_text(f'''# Completed sourcemap splice amendment

Amendment commit: {sys.argv[1]}, one commit on main 1a1c0dc18e6546b5861ee56352c542194471ed70. Evidence appends to the previously published evidence branch. STOP.md is historical. No Loom submission.

The amendment belongs to 99-correctness-axolotl-s. The whole nested flushMappingBuffer owner is pinned before/after with exact text and SHA-256, under createSourceMapGenerator. All owners validate before any write. The contract covers 46 no-write drift/mixed/missing/duplicate/moved/parse refusals, replay bytes/mtime, CRLF/LF and foreign edits. All three complete historical 99s migrations pass.

Workers e8a1529e's full single-file latent census is the restoring mutant observation, with LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1 and LATENT_ONLY_FILE naming the real source file. The scratch overlay sets CommonJSBundle on both fresh lowering constructors; production sources are unchanged. Both runs exit 0 with 59 units and zero checker diagnostics. The historical measurement banner says checker-rejected, but both actual headers have checker_rejected=false. Red reports Refused at adapted sourcemap.ts:319:13 (upstream clear line 316): assigning a non-tuple array's length. It is observed under generator and flush units. Green removes that distinct site: two Refused observations and their two rollback boundaries. Raw records go 137 -> 133. All other findings are identical after exact path/one-byte sourcemap-span relocation; dependency spans are unchanged. Exact real-file copies are named flushMappingBuffer-mutant-red/green-sourcemap.ts. The earlier standalone mutants still stop at apply NotYet and are not the refusal proof.

The complete oracle, including build/test child processes with observations, counts {counts['totals']['flush']:,} flush calls, {counts['totals']['nonempty']:,} nonempty calls/clears and {counts['totals']['chunk1024']:,} calls at length 1024. Reads: length {counts['reads']['length']:,}, push {counts['reads']['push']:,}, apply {counts['reads']['apply']:,}. counts.json records every observed read length with multiplicity. A shadow buffer receives the same pushes and uses length=0 at the actual splice point; every source read compares the two lengths and every apply compares resulting text. Every read after clear and before the next push sees zero: 6,409 push-receiver reads and 9,177 length reads. Source AST audit enumerates all six references; the array is private, captured by two nested functions, passed transiently to native apply, and never assigned, stored or returned as an external alias. Instrumentation is confined to a private counting copy; VM-loaded compiler copies receive the same observer.

Counting oracle: 106,366 passing, one sanctioned Public APIs baseline failure, zero pending; all runners, eight workers, no test filter. The real before/after Node witness prints identical mappings and zero buffer length after every flush, including automatic 1024-code flushes and multi-chunk input.

Both full applies pass; the full amended apply and lane trees' complete source manifests equal the census green tree. The entire source difference from main is the one clear statement. Whole-source transpilation and real bundle comparison permit only that statement replacement; no declaration change is permitted. House adamic types over all 79 compiler roots exits 0 with zero diagnostics. Normal uninstrumented whole lane: PASS, 106,366 passing / one sanctioned failure / zero pending. Shard plan, manifest and Go declarations were regenerated and are byte-identical to main. Every newly retained Go source under stage3 is named .go.txt; shard regeneration used a scratch .go.txt output. Raw evidence stays in this directory on the separate evidence branch; only code, contract, witness, mutant, README/PROOF and the ledger entry are in the amendment.

commands.jsonl contains actual proof child commands, cwd, selected environment and exits. PRELIMINARY.md and SETUP.md retain failed preliminary/setup outcomes, including tool-reported inline housekeeping. The chat tool transcript retains exploratory inspection commands. All required successful receipts are indexed in results.json. The publication commands are git commit and ordinary git push for each authorized branch; no force push or Loom command is used.
''')
# All retained artifacts except this manifest itself are hashed after collection.
manifest={str(p.relative_to(out)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(out.rglob('*')) if p.is_file() and '__pycache__' not in p.parts and p.name!='MANIFEST.json'}
(out/'MANIFEST.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(json.dumps({'status':'PASS','artifacts':len(manifest),'amendment':sys.argv[1]}))
