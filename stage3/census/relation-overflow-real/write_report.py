"""Assemble the final report from the recorded measurements."""
import json,pathlib
root=pathlib.Path(__file__).resolve().parent
read=lambda name:json.loads((root/name).read_text())
phases=['real-checked','real-lowered','real-units','census-checked','census-units']
measurements={phase:{k:v for k,v in read(phase+'.json').items() if k not in ['pairs','checker','phase']} for phase in phases}
lower=read('real-lower.json')
real=read('real-pairs.json')
census=read('census-pairs.json')
text='''# Relation overflow on the real tsc entry

Task ny0c3gb. Measurement only. No compiler, checker policy, adaptation, or production source changes are committed.

## Provenance and method

Main is pinned at `c3b7baf0b` (origin/main when fetched). TypeScript source is `050880ce59e30b356b686bd3144efe24f875ebc8`, version 6.0.3. Exact submodule pins and tools are in [provenance.json](provenance.json). Setup, npm install, adaptation, builds, and measurements have file logs and child exit codes in this directory.

The entry is `/tmp/ny0-adapted/src/tsc/tsc.ts`. The real run calls `load.Load([]string{entry})`, with internal/load's compiler options, prelude, Node library handling, and `SingleThreaded: true`. There is no tsconfig parse or substitution of compiler options. The real loader retains its normal diagnostic rejection. A copy of all production lower package files is compiled under a separate scratch package path to run the unchanged `Lower`, independently of the census overlays.

After production Lower stops, the same process and same checker run the latent first-error top-level unit walker. Each attempted unit has fresh lowering state, while source identity, type text, numeric enum, program facts, and checker caches are shared exactly as in `stage3/census/latent/lower.go.txt`. Generic declarations are deferred until a concrete call reaches them, as in that census. The separate census process loads the same single entry with `LatentLoad`, expands only its resolved source graph using `LatentReplayEntryReach`, and runs the first-error census once without a preceding production Lower. The checker identity assertion succeeds for every graph file. `compiler/checkerpool.go` sets checkerCount to one for SingleThreaded; the pool persists across files and units, not across processes.

Counts walk the same five `Relation.results` maps and `enumRelation` as `RelationOverflowMarked`, testing flag `1 << 5`. TS2859 comes directly from the checker's whole `DiagnosticsCollection`, including globals and library files, without triggering another check. Missing fields panic rather than being silently counted as zero.

The prior evidence at `48ac0f5f` was read, including `evidence/trace-overflow/{base,head}-probe-summary.txt` and the head probe stream. It shows overflow already present at later transformer queries, while early ones were clean. Those probes do not sample directly after Load. The reader in `internal/load/relation_overflow.go` is the basis for these cache and diagnostic reads.

For pair recovery, keys are 128-bit hashes, so type texts cannot be recovered by decoding a key. A scratch overlay adds one passive hook at relater.go's only ComplexityOverflow cache write. It saves original source and target pointers, key, relation, current/error node, and a Go caller stack for at most 50 distinct pairs. It makes no checker queries and does not change a budget or result. All workload counts are captured before `TypeToString` renders these saved pointers. The rendered snapshots separately record any observer effect. Go helper sources are retained as `.go.txt`; `prepare.py` reconstructs the overlays without editing production files on disk.

## Results

'''
text+='Production Lower result:\n\n```json\n'+json.dumps(lower,indent=2)+'\n```\n\n'
text+='Counts immediately after each workload:\n\n```json\n'+json.dumps(measurements,indent=2)+'\n```\n\n'
text+='Unit workload sizes:\n\n```json\n'+json.dumps({mode:read(mode+'-work.json') for mode in ['real','census']},indent=2)+'\n```\n\n'
text+='## Flagged relations\n\n'
for mode,rows in [('real',real),('census',census)]:
 text+=f'### {mode}\n\n'
 if not rows:
  text+='No captured live pair carries the overflow flag.\n\n'
 else:
  text+=f'{len(rows)} captured live flagged pairs are listed below. Full hashes, type IDs, texts, and original caller stacks are in [{mode}-pairs.json]({mode}-pairs.json). Site is the checker error/current node, when held; the caller stack identifies the querying code even when no source node is held.\n\n'
  text+='| Cache | Source type (ID) | Target type (ID) | Checker site |\n|---|---|---|---|\n'
  esc=lambda v:str(v).replace('|','&#124;').replace('\n','<br>')
  for row in rows:
   text+='| '+ ' | '.join(esc(v) for v in [row['cache'],f"{row['source']} ({row['source_id']})",f"{row['target']} ({row['target_id']})",row['site'] or 'No node held'])+' |\n'
  text+='\n'
text+='## Rendering observer check\n\n```json\n'+json.dumps({mode:read(mode+'-rendered.json')['caches'] for mode in ['real','census']},indent=2)+'\n```\n\n'
text+='''## Reproduction

From the repository root, source `/workspace/adamic-tools/env.sh`, run `bash cloud/setup.sh`, `npm ci --prefix stage3/api`, and export `ADAMIC_TYPESCRIPT_SOURCE` to the pinned source checkout. Run `bash stage3/apply.sh /tmp/ny0-adapted` on a new path. Use `run.py LOG COMMAND...` around each child to preserve its actual exit code without a pipeline.

```sh
python3 stage3/census/relation-overflow-real/prepare.py /tmp/ny0-overlay
go build -o /tmp/ny0-adamic ./cmd/adamic
go build -overlay=/tmp/ny0-overlay/real.json -o /tmp/ny0-real ./stage3/census/relation-overflow-real/tool
go build -overlay=/tmp/ny0-overlay/census.json -o /tmp/ny0-census ./stage3/census/relation-overflow-real/tool
/tmp/ny0-real real /tmp/ny0-adapted/src/tsc/tsc.ts OUTPUT_DIRECTORY
/tmp/ny0-census census /tmp/ny0-adapted/src/tsc/tsc.ts OUTPUT_DIRECTORY
```

All helpers terminate normally. No native tsc binary is claimed: production lowering stopped at the recorded diagnostic. These measurements establish the state at that stop and under the census's continuation, not the relation demand of a future successful whole-program lowering.
'''
assert '\u2013' not in text and '\u2014' not in text
(root/'REPORT.md').write_text(text)
