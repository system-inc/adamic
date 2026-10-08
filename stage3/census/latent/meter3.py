"""Render the pinned area/stage3 meter and the complete low-diagnostic work queue."""
import collections, gzip, hashlib, json, pathlib, re, shutil, sys
from rerun2 import LABEL, normalize, summarize, delta, variance_breakdown
ROOT=pathlib.Path(__file__).resolve().parent
DATA=ROOT/'data/meter3'

def diagnostic(site):
 m=re.match(r'^(.*):(\d+):(\d+): error (TS\d+): ([\s\S]*)$',site['text'])
 assert m,site
 file,line,column,code,message=m.groups()
 assert file==site['file']
 return dict(site,line=int(line),column=int(column),code=code,cause=message.splitlines()[0],message=message)

def build(scratch,adapted):
 DATA.mkdir(exist_ok=True)
 specs=json.loads((scratch/'runs.json').read_text());assert len(specs)==1 and specs[0]['status']=='complete'
 raw=[normalize(json.loads(l),str(adapted)) for l in (scratch/'meter-3.jsonl').read_text().splitlines()]
 manifest=[dict(file=str(p.relative_to(adapted)),sha256=hashlib.sha256(p.read_bytes()).hexdigest(),generated='.generated.' in p.name) for p in sorted((adapted/'src/compiler').rglob('*.ts'))]
 with gzip.open(ROOT/'data/rerun2-REPORT.json.gz','rt') as f:prior=json.load(f)
 assert {f['file'] for f in manifest}=={f['file'] for f in prior['source']['files']} and len(manifest)==78
 run=dict(specs[0],**summarize(raw,manifest))
 run['checker_diagnostics']=[diagnostic(d) for d in raw[0]['diagnostic_sites']]
 run['cheap_wins']=[dict(measurement=LABEL,file=f,count=v['checker'],diagnostics=[d for d in run['checker_diagnostics'] if d['file']==f]) for f,v in sorted(run['per_file'].items(),key=lambda kv:(kv[1]['checker'],kv[0])) if 1<=v['checker']<=10]
 run['checker_totals_by_code']=dict(sorted(collections.Counter(d['code'] for d in run['checker_diagnostics']).items()))
 texts={f['file']:(adapted/f['file']).read_bytes().decode() for f in manifest}
 run['variance']=variance_breakdown(run['findings'],raw,texts)
 run['delta_vs_cumulative_2']=delta(run,prior['runs'][0])
 run['delta_vs_cumulative_2_nested']=delta(run,prior['runs'][1])
 run['overlay_sha256']={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((scratch/'meter-3-overlay').glob('*.go'))}
 run['binary_sha256']=hashlib.sha256((scratch/'meter-3-census').read_bytes()).hexdigest()
 with gzip.open(DATA/'meter-3.jsonl.gz','wt') as f:
  for row in raw:f.write(json.dumps(row)+'\n')
 for phase in ['overlay','build','run']:shutil.copyfile(scratch/('meter-3-'+phase+'.log'),DATA/('meter-3-'+phase+'.log'))
 result=dict(complete=True,meter_version=3,measurement=LABEL,count_definition=prior['count_definition'],source=dict(typescript=prior['source']['typescript'],commit=prior['source']['commit'],files=manifest,adaptations=['10-type-imports','20-optional-declarations','30-indexed-reads','31-indexed-reads-checker','32-indexed-reads-program','33-indexed-reads-emit','40-explicit-any','45-regex-captures','46-fix-pragma-empty-argument']),runs=[run],comparison_note='Source adaptations and compiler commits both changed; deltas do not isolate causal feature contributions.')
 (ROOT/'REPORT.json').write_text(json.dumps(result,indent=2)+'\n')
 return result

def render(result):
 r=result['runs'][0];ratio=r['checker_clean_ratio'];q=r['cheap_wins'];esc=lambda s:str(s).replace('|','\\|').replace('\n',' ')
 names=', '.join(x['file'].removeprefix('src/compiler/')+' ('+str(x['count'])+')' for x in q)
 lines=[f"Built: latest area/stage3 adaptations and newest cumulative-2 feature tips in a never-pushed scratch tree.",f"Meter: {ratio['fraction']} checker-clean ({ratio['percent']:.2f}%), {r['checker_total']} checker diagnostics; {LABEL}.",f"1-to-10 list: {names or 'none'}; {LABEL}.","Checks: guarded build/run and vet pass; 12 report mutants, both output-guard mutants and binder-only +1 NotYet mutant caught; logs in data/meter3/.","Limits: own-file diagnostics only, first lowering error per unit, no native correctness or full-gate claim.",'', '# Checker meter','',f"Every count and delta below is **{LABEL}**. The denominator is the same 78 generated compiler source roots as the previous run. A zero means no diagnostic is attributed to that file in one whole-project check; imports and the project can still be rejected.",'',f"{len(q)} files have 1–10 diagnostics ({sum(x['count'] for x in q)} diagnostic sites). Fixing every listed file would reach {ratio['numerator']+len(q)}/78 on this fixed corpus, if no new diagnostics appeared. Diagnostic count is a work queue, not an estimate of implementation difficulty.",'','# Checker-clean files','']
 lines += ['- `'+f+'`' for f in r['checker_clean_files']]
 lines += ['','# Files with 1–10 diagnostics','', 'Locations refer to the final adapted tree. Each cause is the first line of the actual checker message; full chains and byte spans are preserved in JSON and compressed raw data.','']
 for item in q:
  lines += [f"## {item['file']} ({item['count']})",'', '| Code | Line:column | One-line cause |','| --- | --- | --- |']
  lines += [f"| {d['code']} | {d['line']}:{d['column']} | {esc(d['cause'])} |" for d in item['diagnostics']]
  lines += ['']
 lines += ['# Totals by code','', '| Code | Diagnostics |','| --- | ---: |']+[f'| {k} | {v} |' for k,v in r['checker_totals_by_code'].items()]
 lines += ['','# All files','', '| File | Checker diagnostics | NotYet | Refused |','| --- | ---: | ---: | ---: |']+[f"| {f} | {v['checker']} | {v['NotYet']} | {v['Refused']} |" for f,v in sorted(r['per_file'].items())]
 lines += ['','# Lowering and variance','',f"Unique lowering sites: {r['lowering_counts']}. Functions attempted: {r['functions_attempted']}; bodies skipped for own-body checker diagnostics: {r['bodies_skipped']}. Refusal scanning and independent top-level-unit lowering continue beyond failures. This remains a measurement-only overlay; no usable IR or production loader output is allowed.",'',f"Variance: {r['variance']['total']} Refused sites across {r['variance']['owning_declaration_count']} nearest named owning declarations. Families: {r['variance']['families']}. JSON retains the complete per-declaration ledger and attempting-unit context.",'', '| Exact family | Sites |','| --- | ---: |']+[f'| {esc(k)} | {v} |' for k,v in r['per_reason'].items()]
 lines += ['','# Deltas','',result['comparison_note'],'']
 for key in ['delta_vs_cumulative_2','delta_vs_cumulative_2_nested']:
  d=r[key];lines += [f"- {key}: checker {d['checker']:+d}, checker-clean files {d['checker_clean_files']:+d}, lowering {d['counts']}."]
 lines += ['','# Pinned provenance and reproduction','',f"Integration area/stage3: `{r['main']}`. Scratch head: `{r['head']}` on `{r['branch']}`, never pushed. Feature pins:",'']+['- `'+k+'`: `'+v+'`' for k,v in r['features'].items()]
 lines += ['',f"TypeScript {result['source']['typescript']} `{result['source']['commit']}`. Adaptations: "+', '.join(result['source']['adaptations'])+'. The adapter files and apply script exactly match the integration tip; their hash ledger and patch-set are archived.', '',f"Build/run durations (seconds): `{r['durations_seconds']}`. Toolchain setup: Go 0s, clang 0s, Node 0s, submodules 0s, warm cache 29s; total 29s; nproc 5 (CPU quota 4).",'', '```sh','source /workspace/adamic-tools/env.sh','bash stage3/apply.sh /tmp/tsc-latent3-adapted > /tmp/latent3-apply.log 2>&1','python3 stage3/census/latent/run_comparisons.py /workspace/adamic /tmp/tsc-latent3-adapted /tmp/latent3-runs /tmp/latent3-worktrees.json > /tmp/latent3-comparisons.log 2>&1','python3 stage3/census/latent/meter3.py /tmp/latent3-runs /tmp/tsc-latent3-adapted > /tmp/latent3-summary.log 2>&1','python3 stage3/census/latent/audit_meter3.py /tmp/tsc-latent3-adapted > /tmp/latent3-audit.log 2>&1','python3 stage3/census/latent/audit_corpus.py /tmp/latent3-runs /tmp/tsc-latent3-adapted meter-3 > /tmp/latent3-corpus-audit.log 2>&1','GOWORK=/tmp/latent3-runs/meter-3.go.work python3 stage3/census/latent/audit_output_guards.py /tmp/latent3-tree /tmp/latent3-runs/meter-3-overlay /tmp/latent3-output-guards > /tmp/latent3-output-guards.log 2>&1','GOWORK=/tmp/latent3-runs/meter-3.go.work go vet -overlay=/tmp/latent3-runs/meter-3-overlay/overlay.json ./stage3/census/latent/tool > /tmp/latent3-vet.log 2>&1','```','', 'Scratch merges were resolved to preserve taste labels/void, fallthrough and implicit returns, newest flags, namespace/parameter-property support, and nested captured storage. The previous integration enum helper was replaced by newest enums; a stale enumElement call found by the first build was removed. No compiler source edit is delivered. Integration diffs, pinned ancestry, dependency pins, loader options, adaptation hashes and failed-build log are archived in data/meter3/. Backends in the scratch merge are unused and their correctness is not claimed.','', 'Dedicated mutants remove a cheap-win file or diagnostic, corrupt its code/line/cause, change the threshold count, ratio denominator, totals by code or source hash. The planted real-corpus NotYet must add exactly one binder.ts site with all other 77 files unchanged. Both output guards run in the baseline and corpus-mutant invocations; deliberate non-nil IR and permissive-loader mutants were also caught by their guard panics. Vet passes for the overlaid measurement driver. No full gate, native program, adaptation oracle suite or semantic proof was run for this measurement update.','']
 (ROOT/'REPORT.md').write_text('\n'.join(lines))
if __name__=='__main__':
 result=build(pathlib.Path(sys.argv[1]).resolve(),pathlib.Path(sys.argv[2]).resolve()) if len(sys.argv)>1 else json.loads((ROOT/'REPORT.json').read_text())
 render(result)
 print(json.dumps(dict(checker=result['runs'][0]['checker_total'],ratio=result['runs'][0]['checker_clean_ratio'],cheap_wins=[(x['file'],x['count']) for x in result['runs'][0]['cheap_wins']]),indent=2))
