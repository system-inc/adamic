import json,pathlib,shutil,gzip
root=pathlib.Path(__file__).resolve().parent
phases=['real-checked','real-lowered','real-units','census-checked','census-units']
measurements={}
for phase in phases:
 p=root/(phase+'.json')
 if p.exists():
  x=json.loads(p.read_text())
  measurements[phase]={k:v for k,v in x.items() if k not in ['pairs','checker']}
 else:measurements[phase]={'status':'not_collected_before_environment_deadline','caches':None,'RelationOverflow':None,'RelationOverflowMarked':None,'TS2859':None}
processes={}
for mode in ['real','census']:
 for suffix in ['.log','.log.json']:
  source=pathlib.Path('/tmp/ny0-'+mode+'-run-corrected'+suffix)
  if source.exists():shutil.copyfile(source,root/(mode+'-run'+suffix))
 p=root/(mode+'-run.log.json')
 if p.exists():processes[mode]=json.loads(p.read_text())
 raw=root/(mode+'-units.jsonl')
 if raw.exists():
  status='' if (root/(mode+'-units.json')).exists() else '-incomplete'
  with raw.open('rb') as inp,(root/(mode+'-units'+status+'.jsonl.gz')).open('wb') as out:
   with gzip.GzipFile(filename='',mode='wb',fileobj=out,mtime=0) as compressed:shutil.copyfileobj(inp,compressed)
  raw.unlink()
text='''# Relation overflow on the real tsc entry

Task ny0c3gb. Partial measurement, stopped for the environment time limit. No compiler, adaptation, or checker policy changes are committed.

## Answers

A real compile's checker does not overflow after checking or after today's production lowering stop. Both RelationOverflow values are empty, both RelationOverflowMarked values are false, every relation cache has zero flagged pairs, and no TS2859 is held at those points. These counts were repeated and agreed exactly.

No relation is flagged in either measured real-compile state. The canonical entry-first unit continuation and independent census did not both produce their end-state evidence before the environment deadline. Missing end-state counts and pair texts are unknown, not zero. The pooled-checker end-state comparison is therefore not established by this run.

Production lowering returns no IR and stops at src/compiler/core.ts:2599:15, the ambient process host binding, pending runtime/platform clearance. It stops in the refusal scan before declareModule and moduleStatements. It adds 1,156 unflagged cache entries. A future successful whole-program lowering is not measured.

## Recorded counts

Only completed snapshots are reported as measurements. Missing fields are explicitly null.

```json
'''+json.dumps(measurements,indent=2)+'\n```\n\n'
text+='Production lowering result:\n\n```json\n'+(root/'real-lower.json').read_text()+'\n```\n\n'
text+='## Pair recovery from completed continuations\n\n'
for mode in ['real','census']:
 snap=root/(mode+'-units.json')
 pairs=root/(mode+'-pairs.json')
 if snap.exists() and pairs.exists():
  keys={(x['cache'],x['key']) for x in json.loads(snap.read_text())['pairs']}
  rows=[x for x in json.loads(pairs.read_text()) if (x['cache'],x['key']) in keys]
  shown=[{k:v for k,v in x.items() if k!='stack'} for x in rows[:50]]
  text+=mode+':\n\n```json\n'+json.dumps(shown,indent=2)+'\n```\n\nOriginal caller stacks remain in '+mode+'-pairs.json.\n\n'
 else:text+=mode+': no completed pair recovery was written before stopping.\n\n'
text+='''## Method and stopping reason

Main is c3b7baf0b3438bf5ac7aa9fcf6eb2ba385e2f9bd, origin/main when fetched. TypeScript is pinned at 050880ce59e30b356b686bd3144efe24f875ebc8, version 6.0.3. Required setup, npm ci, stage3/apply.sh, main adamic, and measurement helper builds passed. Exact pins, source checksums, binary checksums, and overlay hashes are in provenance.json.

The real run calls internal/load.Load on the single src/tsc/tsc.ts entry with the loader's own options and SingleThreaded true, without a tsconfig parse. It uses a byte-for-byte copy of production lowering in a separate scratch package, followed by the census's fresh lowering per unit on that same checker. The independent census uses LatentLoad and the canonical LatentEntryReach from stage3/meter/entry_overlay.py, keeping the entry first and then following compiler dependency order. Each program's graph files were asserted to share one checker. The census pool has one checker for SingleThreaded and retains it across files and units. File-wide refusal scans and registration/program-fact probes are part of the census work too.

Counts read the same five Relation.results maps and enumRelation as RelationOverflowMarked, testing bit 1 << 5. TS2859 is read directly from the checker's DiagnosticsCollection, without another semantic check. A passive scratch hook at the sole ComplexityOverflow cache write retains original type pointers and caller stacks for later printing; it does not ask relations or change budgets. Type text rendering is deferred until the unit end-state snapshot.

The prior evidence at 48ac0f5f was read, including its trace-overflow probe summaries and head probe stream. It establishes late pooled-checker marks, not a post-Load measurement. The current real compile measurements separate that question and are clear.

An initial continuation attempt used the replay reach helper, which incorrectly put dependencies before the entry for this census. It was terminated, archived under invalid-entry-order-*, and excluded from the unit comparison. Its checking and ordinary lowering samples were valid and match the repeated corrected samples. The corrected helpers use the canonical entry order. They remained active for the recorded durations without producing all requested end-state evidence. They were terminated before the user's approximately two-hour environment window ended, so no started measurement process remains running. The census buffers records until the whole walk completes; incomplete buffered output is not presented as a completed census.

Two redundant preparation builds were also terminated intentionally before successful retries. All wrappers preserve the child's own exit code. Full logs and numeric exit records are retained.

## Final child results

```json
'''+json.dumps(processes,indent=2)+'\n```\n\n'
text+='''## Reproduction

The retained main.go.txt, snapshot.go.txt, checker-hook.go.txt, prepare.py, and run.py reconstruct the measurement without editing production files. Source the tool environment, run the setup specified by the task, export ADAMIC_TYPESCRIPT_SOURCE to the pinned checkout, and build a fresh adapted tree. Then prepare fresh overlays and build the two tools:

```sh
python3 stage3/census/relation-overflow-real/prepare.py /tmp/ny0-overlay-new
go build -overlay=/tmp/ny0-overlay-new/real.json -o /tmp/ny0-real-new ./stage3/census/relation-overflow-real/tool
go build -overlay=/tmp/ny0-overlay-new/census.json -o /tmp/ny0-census-new ./stage3/census/relation-overflow-real/tool
```

Run each tool as `TOOL real|census ENTRY OUTPUT_DIRECTORY`, using a fresh output directory and run.py to preserve its log and exit code. These longer unit walks need a runtime window beyond the remaining time available in this session. No end-state result is inferred from the interrupted attempts.
'''
assert '\u2013' not in text and '\u2014' not in text
(root/'REPORT.md').write_text(text)
(root/'completion.json').write_text(json.dumps({'status':'partial','reason':'environment runtime window; canonical unit walks interrupted before all end-state evidence was written','completed_snapshots':[x for x,y in measurements.items() if 'status' not in y]},indent=2)+'\n')
print('Partial report written with unknown end-states explicitly null')
