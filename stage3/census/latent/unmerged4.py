"""Recount the two untouched tips, each measured on its own adapted source bytes."""
import gzip, hashlib, json, pathlib, shutil, sys
from rerun2 import LABEL, normalize, summarize
ROOT=pathlib.Path(__file__).resolve().parent;DATA=ROOT/'data/unmerged4'

def top_ten(reasons):
 return [dict(measurement=LABEL,kind=k.split(': ',1)[0],reason=k.split(': ',1)[1],count=v) for k,v in sorted(((k,v) for k,v in reasons.items() if k.startswith(('NotYet: ','Refused: '))),key=lambda kv:(-kv[1],kv[0]))[:10]]

def build(scratch):
 runs=[]
 for spec in json.loads((scratch/'runs.json').read_text()):
  assert spec['status']=='complete' and spec['head']==spec['main'] and not spec['features'] and spec['unmerged'],spec
  source=pathlib.Path(spec['adapted']);name=spec['name']
  manifest=[dict(file=str(p.relative_to(source)),sha256=hashlib.sha256(p.read_bytes()).hexdigest(),generated='.generated.' in p.name) for p in sorted((source/'src/compiler').rglob('*.ts'))]
  raw=[normalize(json.loads(l),str(source)) for l in (scratch/(name+'.jsonl')).read_text().splitlines()]
  row=dict(spec,**summarize(raw,manifest,len(manifest)));row['source']=dict(files=manifest,typescript='6.0.3',commit='050880ce59e30b356b686bd3144efe24f875ebc8');row['top_10']=top_ten(row['per_reason'])
  row['overlay_sha256']={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((scratch/(name+'-overlay')).glob('*.go'))}
  row['binary_sha256']=hashlib.sha256((scratch/(name+'-census')).read_bytes()).hexdigest()
  row['source']['root_count']=len(manifest)
  row['source']['extra_adapter_roots']=['src/compiler/hostErrors.ts'] if name=='area-unmerged' else []
  with gzip.open(DATA/(name+'.jsonl.gz'),'wt') as f:
   for item in raw:f.write(json.dumps(item)+'\n')
  for phase in ['overlay','build','run']:shutil.copyfile(scratch/(name+'-'+phase+'.log'),DATA/(name+'-'+phase+'.log'))
  runs.append(row)
 assert [r['name'] for r in runs]==['main-unmerged','area-unmerged']
 result=dict(complete=True,meter_version=4,measurement=LABEL,count_definition='Unique (kind, where, reason, text) findings across independent top-level attempts. NotYet and Refused are separate; SkippedDependency, ordinary errors and panics are excluded. Top ten combines NotYet and Refused exact reason families, ranked by count descending then kind/reason lexical order.',method='Unchanged checker-body eligibility and independent-unit lowering; each untouched commit uses its own apply.sh and adapted source. No feature merges or compiler source edits.',runs=runs)
 (ROOT/'REPORT.json').write_text(json.dumps(result,indent=2)+'\n');return result

def render(result):
 a,b=result['runs'];esc=lambda s:str(s).replace('|','\\|').replace('\n',' ')
 lines=[f"Built: latent census of untouched main {a['head']} and area/stage3 {b['head']}, each using its own apply.sh.",f"Main: NotYet {a['lowering_counts']['NotYet']}, Refused {a['lowering_counts']['Refused']}; {LABEL}.",f"Area/stage3: NotYet {b['lowering_counts']['NotYet']}, Refused {b['lowering_counts']['Refused']}; {LABEL}.","Checks: both guarded builds/runs, independent unique-site/ranking/body-span recount, 12 dedicated evidence mutants, and output-guard mutants for both binaries; logs in data/unmerged4/.","Limits: first lowering failure per unit; checker-diagnosed bodies skipped; no usable IR, native run, adaptation oracle suite or full gate.",'', '# Summary','',f"All counts in this report and JSON are **{LABEL}**. These are lowering observations after bypassing only the checker rejection in the measurement loader; they do not establish accepted programs or native correctness. Both top-10 tables combine NotYet and Refused exact reason families; lower kinds, dependency skips, ordinary errors and panics are excluded from the ranking.",'']
 for r in result['runs']:
  lines += [f"## {r['name']}",'',f"NotYet **{r['lowering_counts']['NotYet']}**; Refused **{r['lowering_counts']['Refused']}**. **{LABEL}**.",'','| Kind | Exact reason | Count |','| --- | --- | ---: |']+[f"| {v['kind']} | {esc(v['reason'])} | {v['count']} |" for v in r['top_10']]
  lines += ['',f"Checker diagnostics: {r['checker_total']}; own-file zero diagnostics: {r['checker_clean_ratio']['fraction']}. Functions attempted: {r['functions_attempted']}; own-body skips: {r['bodies_skipped']}. All lowering outcomes: `{r['lowering_counts']}`.",'']
 lines += ['# Provenance and reproduction','',result['method'],'','Both scratch heads equal their fetched origin pins, not merge commits made for this measurement. Compiler files, loader options and all adaptation files match the pinned trees byte for byte. The existing scratch overlay supplies measurement hooks and disables ordinary output APIs. Each config uses the same cohere pin recorded by both original trees; dependency and adaptation hashes, patch sets and command logs are archived. Main has 78 source roots; area/stage3 has 79, including its adapter-created hostErrors.ts helper. The two source manifests differ, so this is a two-tree observation, not an isolated compiler-feature delta.','']
 for r in result['runs']:
  lines += [f"- {r['name']}: `{r['head']}` on never-pushed `{r['branch']}`; adapted input `{r['adapted']}`; elapsed seconds `{r['durations_seconds']}`."]
 lines += ['', '```sh','source /workspace/adamic-tools/env.sh','bash /tmp/latent4-main/stage3/apply.sh /tmp/tsc-latent4-main > /tmp/latent4-main-apply.log 2>&1','bash /tmp/latent4-area/stage3/apply.sh /tmp/tsc-latent4-area > /tmp/latent4-area-apply.log 2>&1','python3 stage3/census/latent/run_comparisons.py /workspace/adamic /tmp/tsc-latent4-main /tmp/latent4-runs /tmp/latent4-worktrees.json > /tmp/latent4-comparisons.log 2>&1','python3 stage3/census/latent/unmerged4.py /tmp/latent4-runs > /tmp/latent4-summary.log 2>&1','python3 stage3/census/latent/audit_unmerged4.py > /tmp/latent4-audit.log 2>&1','python3 stage3/census/latent/audit_corpus.py /tmp/latent4-runs /tmp/tsc-latent4-main main-unmerged /tmp/latent4-main-unmerged-mutant > /tmp/latent4-main-unmerged-corpus-audit.log 2>&1','GOWORK=/tmp/latent4-runs/main-unmerged.go.work python3 stage3/census/latent/audit_output_guards.py /tmp/latent4-main /tmp/latent4-runs/main-unmerged-overlay /tmp/latent4-main-unmerged-guards > /tmp/latent4-main-unmerged-guards.log 2>&1','python3 stage3/census/latent/audit_corpus.py /tmp/latent4-runs /tmp/tsc-latent4-area area-unmerged /tmp/latent4-area-unmerged-mutant > /tmp/latent4-area-unmerged-corpus-audit.log 2>&1','GOWORK=/tmp/latent4-runs/area-unmerged.go.work python3 stage3/census/latent/audit_output_guards.py /tmp/latent4-area /tmp/latent4-runs/area-unmerged-overlay /tmp/latent4-area-unmerged-guards > /tmp/latent4-area-unmerged-guards.log 2>&1','```','', 'The audit independently deduplicates all raw events, attributes findings by actual location, recounts each reason and file, verifies top-10 order/counts and body skip spans, and checks the unmerged head/dependency/source manifests. Dedicated artifact mutants corrupt totals, reason counts, rank counts/order, pins, source hashes, the actual area source denominator and body skip evidence. The real-corpus extra-NotYet mutant runs on each binary and must change only binder.ts by exactly one unique site. An initial attempt to run the two corpus mutants with a shared output filename was stopped; only the isolated reruns are evidence. No internal/ edits are committed on the delivery branch.','', 'Toolchain setup: Go ready 0s, clang 1s, Node 1s, submodules 1s, cache warm 33s; total 33s, nproc 5, CPU quota 4. Exact timings are in the archived setup log.','']
 for r in result['runs']:
  lines += ['# '+r['name']+' full per-file counts','', '| File | Checker | NotYet | Refused |','| --- | ---: | ---: | ---: |']+[f"| {f} | {v['checker']} | {v['NotYet']} | {v['Refused']} |" for f,v in sorted(r['per_file'].items())]
  lines += ['', '# '+r['name']+' all exact reasons','', '| Kind and reason | Count |','| --- | ---: |']+[f"| {esc(k)} | {v} |" for k,v in sorted(r['per_reason'].items(),key=lambda kv:(-kv[1],kv[0]))]+['']
 (ROOT/'REPORT.md').write_text('\n'.join(lines))
if __name__=='__main__':
 result=build(pathlib.Path(sys.argv[1])) if len(sys.argv)>1 else json.loads((ROOT/'REPORT.json').read_text());render(result)
 print(json.dumps([dict(name=r['name'],counts=r['lowering_counts'],top_10=r['top_10']) for r in result['runs']],indent=2))
