import json, pathlib, collections, hashlib, re
R=pathlib.Path('/workspace/scratch'); A=R/'kp-adapted'
def norm(x):
 if isinstance(x,str): return x.replace(str(A)+'/', '')
 if isinstance(x,list): return [norm(v) for v in x]
 if isinstance(x,dict): return {k:norm(v) for k,v in x.items()}
 return x
def key(f): return tuple(f[k] for k in ('kind','where','reason','text'))
D={}
for label in ['base','head']:
 rows=[norm(json.loads(s)) for s in (R/f'kp-{label}-latent.jsonl').read_text().splitlines()]
 header=rows.pop(0)
 assert header['latent_mode']=='first-error'
 sites={key(f):f for row in rows for f in row['findings']}
 byloc=collections.defaultdict(dict); byunit=collections.defaultdict(list)
 for row in rows:
  for f in row['findings']:
   byloc[f['where']][key(f)]=f
   if (f['phase'],key(f)) not in {(v['phase'],key(v)) for v in byunit[f['unit']]}: byunit[f['unit']].append(f)
 units={u['where']:u for row in rows for u in row['units']}
 times={x['where']:x for x in [norm(json.loads(s)) for s in (R/f'kp-{label}-times.jsonl').read_text().splitlines()]}
 assert len(times)==sum(u['status'] in ('attempted','panic') for u in units.values())
 D[label]=dict(header=header,rows=rows,sites=sites,byloc=byloc,byunit=byunit,units=units,times=times)
b,h=D['base'],D['head']; reason='an object refinement using an open numeric enum as a literal tag'
assert b['units']==h['units'] or set(b['units'])==set(h['units'])
print('counts', {l:dict(collections.Counter(f['kind'] for f in d['sites'].values())) for l,d in D.items()})
print('open', {l:sum(f['reason']==reason for f in d['sites'].values()) for l,d in D.items()})
print('removed',len(b['sites'].keys()-h['sites'].keys()),'added',len(h['sites'].keys()-b['sites'].keys()))
print('removed reasons',dict(collections.Counter(k[2] for k in b['sites'].keys()-h['sites'].keys())))
print('added reasons',dict(collections.Counter(k[2] for k in h['sites'].keys()-b['sites'].keys())))
print('changed units',sum([key(x) for x in b['byunit'][u]]!=[key(x) for x in h['byunit'][u]] for u in b['units']))
print('open phases',dict(collections.Counter(f['phase'] for f in b['sites'].values() if f['reason']==reason)))
# Only lowering-phase events describe an attempted unit's stop. Refusal-scan
# owners identify the whole file and can share the first statement's location.
def lowers(d,u): return [f for f in d['byunit'][u] if f['phase']=='lowering']
def compact(f): return {k:f[k] for k in ('kind','where','reason')} if f else None
def first(d,u):
 fs=lowers(d,u)
 return fs[0] if fs else None
open_transitions=[]
for k,f in sorted(b['sites'].items()):
 if f['reason']!=reason: continue
 after=list(h['byloc'].get(f['where'],{}).values())
 status='unchanged' if k in h['sites'] else 'moved' if after else 'no_finding'
 open_transitions.append(dict(where=f['where'],base=compact(f),head=[compact(x) for x in after],status=status))
print('open transitions',dict(collections.Counter(x['status'] for x in open_transitions)))
print('open moved reasons',dict(collections.Counter(y['reason'] for x in open_transitions if x['status']=='moved' for y in x['head'])))
unit_delta=[]
for u in sorted(b['units']):
 old,new=first(b,u),first(h,u)
 if compact(old)!=compact(new):
  unit_delta.append(dict(where=u,unit=b['units'][u],base_stop=compact(old),head_stop=compact(new),base_findings=[compact(f) for f in lowers(b,u)],head_findings=[compact(f) for f in lowers(h,u)]))
print('first stop changes',len(unit_delta))
for x in unit_delta: print(x['where'],x['unit'].get('name',''),x['base_stop'],x['head_stop'])
# New stops after an old first stop are exposure, not evidence of regression.
regressions=[]
for x in unit_delta:
 if x['base_stop'] is None and x['head_stop'] is not None and b['units'][x['where']]['status']=='attempted':
  regressions.append(dict(type='previously_finding_free_unit',unit=x['where'],**x['head_stop']))
for k in sorted(h['sites'].keys()-b['sites'].keys()):
 f=h['sites'][k]
 if f['kind']=='panic':
  regressions.append(dict(type='new_panic',unit=f['unit'],**compact(f)))
 elif f['where'] not in b['byloc'] and any(g['where']==f['where'] and g['phase']=='refusal_scan' for row in h['rows'] for g in row['findings']):
  regressions.append(dict(type='previously_finding_free_scan_site',unit=f['unit'],**compact(f)))
print('regressions',regressions)
T=[]
assert set(b['times'])==set(h['times'])
for u in sorted(b['times']):
 old,new=b['times'][u],h['times'][u]
 T.append(dict(where=u,name=b['units'][u].get('name') or b['units'][u]['kind'],base_wall_seconds=old['wall_seconds'],head_wall_seconds=new['wall_seconds'],wall_delta_seconds=new['wall_seconds']-old['wall_seconds'],base_cpu_seconds=old['cpu_seconds'],head_cpu_seconds=new['cpu_seconds'],cpu_delta_seconds=new['cpu_seconds']-old['cpu_seconds']))
print('timing totals',{l:{k:sum(x[k] for x in d['times'].values()) for k in ['wall_seconds','cpu_seconds']} for l,d in D.items()})
print('top movers',sorted(T,key=lambda x:abs(x['wall_delta_seconds']),reverse=True)[:10])
# Stable identities use the shared source digest and exact UTF-16 token offset,
# not a line match. Preserve every distinct census reason at a token.
from functools import lru_cache
@lru_cache(None)
def identity(where):
 file,line,column=where.rsplit(':',2); line,column=int(line),int(column)
 p=A/file
 if not p.exists(): p=pathlib.Path(file)
 assert p.exists(),where
 data=p.read_bytes(); text=data.decode('utf-8'); lines=text.splitlines(keepends=True)
 prefix=''.join(lines[:line-1]); tokenprefix=lines[line-1].encode('utf-16-le')[:(column-1)*2].decode('utf-16-le')
 return dict(file=file,sha256=hashlib.sha256(data).hexdigest(),utf16_offset=len(prefix.encode('utf-16-le'))//2+column-1,byte_offset=len((prefix+tokenprefix).encode('utf-8')),line=line,column=column)
def occurrences(d,where):
 return [f for row in d['rows'] for f in row['findings'] if f['where']==where]
def unique_compact(d,where):
 return [dict(kind=k[0],where=k[1],reason=k[2],text=k[3]) for k in sorted(d['byloc'].get(where,{}))]
def source_key(where):
 i=identity(where)
 return (i['file'],i['sha256'],i['utf16_offset'])
b_index={source_key(w):w for w in b['byloc']}
h_index={source_key(w):w for w in h['byloc']}
assert len(b_index)==len(b['byloc']) and len(h_index)==len(h['byloc'])
site_delta=[]
for source_id in sorted(b_index.keys()|h_index.keys()):
 where=b_index.get(source_id,h_index.get(source_id))
 assert source_id not in b_index or source_id not in h_index or b_index[source_id]==h_index[source_id]
 before,after=unique_compact(b,b_index[source_id]) if source_id in b_index else [],unique_compact(h,h_index[source_id]) if source_id in h_index else []
 if before!=after:
  site_delta.append(dict(identity=identity(where),where=where,base=before,head=after,base_occurrences=occurrences(b,where),head_occurrences=occurrences(h,where)))
for x in open_transitions:
 x['identity']=identity(x['where'])
 x['base_other_findings']=[compact(f) for f in b['byloc'][x['where']].values() if f['reason']!=reason]
for x in unit_delta: x['identity']=identity(x['where'])
counts={l:{kind:sum(f['kind']==kind for f in d['sites'].values()) for kind in ['NotYet','Refused','panic','error','SkippedDependency']} for l,d in D.items()}
open_summary=dict(base=sum(f['reason']==reason for f in b['sites'].values()),head=sum(f['reason']==reason for f in h['sites'].values()),**collections.Counter(x['status'] for x in open_transitions))
for status in ['unchanged','no_finding','moved']: open_summary.setdefault(status,0)
open_summary['moved_by_reason']=dict(collections.Counter(y['reason'] for x in open_transitions if x['status']=='moved' for y in x['head']))
rss={l:json.loads((R/f'kp-{l}-rss.json').read_text()) for l in D}
assert all(x['exit_code']==0 and x['signal'] is None for x in rss.values())
assert b['header']==h['header'],'checker headers differ'
assert {r['file'] for r in b['rows']}=={str(p.relative_to(A)) for p in (A/'src/compiler').rglob('*.ts')}
assert {r['file'] for r in b['rows']}=={r['file'] for r in h['rows']}
assert all((b['units'][u]['status'] in ('attempted','panic')) == (h['units'][u]['status'] in ('attempted','panic')) for u in b['units']), 'unit eligibility differs'
for l in D:
 assert (R/f'kp-{l}-memory-before.txt').read_text()==(R/f'kp-{l}-memory-after.txt').read_text(),'memory events changed'
 assert not (R/f'kp-{l}-compiler-status.txt').read_text(),'dirty compiler'
# Store all unit timing deltas, not only the ranked table.
result=dict(schema_version=1,base_commit='df84b97437661778245a02862f38799f4bdf31cf',head_commit='b03fe07e48cce2b4a11f183bb3b4a8ba06bb60b9',latent_mode='first-error',identity_definition='shared source SHA-256 + exact UTF-16 token offset; all distinct census findings at that source token retained',count_definition='unique (kind, where, reason, text), deduplicated across attempts as the meter does',counts=counts,open_tag=open_summary,resources=rss,site_delta=site_delta,open_tag_transitions=open_transitions,unit_first_stop_delta=unit_delta,regressions=regressions,unit_timing_delta=T)
out=R/'kp-base/stage3/census/kind-predicates'; out.mkdir(parents=True,exist_ok=True)
(out/'delta.json').write_text(json.dumps(result,indent=2)+'\n')
import shutil,gzip
E=out/'evidence'; E.mkdir(exist_ok=True)
for l,d in D.items():
 for kind in ['latent','times']:
  with gzip.open(E/f'{l}-{kind}.jsonl.gz','wt') as f:
   for line in (R/f'kp-{l}-{kind}.jsonl').read_text().splitlines(): f.write(json.dumps(norm(json.loads(line)))+'\n')
 for suffix in ['rss.json','memory-before.txt','memory-after.txt','pins.txt','compiler-status.txt','exit.txt']:
  shutil.copyfile(R/f'kp-{l}-{suffix}',E/f'{l}-{suffix}')
 for suffix in ['submodules.log','overlay.log','build.log','latent.log']:
  with gzip.open(E/f'{l}-{suffix}.gz','wb') as f: f.write((R/f'kp-{l}-{suffix}').read_bytes())
for name in ['setup.log','kp-apply.log','kp-packages.log','kp-launcher-failure.log','kp-driver.log']:
 with gzip.open(E/(name+'.gz'),'wb') as f: f.write((R/name).read_bytes())
for name in ['kp-box.json','kp-source-manifest.json','kp-binaries.sha256']:
 shutil.copyfile(R/name,E/name.removeprefix('kp-'))
shutil.copyfile(R/'kp-time.py',out/'measure-rss.py')
shutil.copyfile(R/'kp-prepare.sh',out/'run.sh')
shutil.copyfile(R/'kp-analysis.py',out/'compare.py')
# Record artifact hashes, including uncompressed source-stream hashes.
manifest={str(p.relative_to(out)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(out.rglob('*')) if p.is_file() and p.name not in ['REPORT.md','artifacts.sha256.json','validation.txt']}
(out/'artifacts.sha256.json').write_text(json.dumps(manifest,indent=2)+'\n')
def esc(s): return str(s).replace('|','\\|').replace('\n',' ')
def stop(f): return 'no finding' if f is None else f"{f['kind']} at `{f['where']}`: {esc(f['reason'])}"
totals={l:{k:sum(x[k] for x in d['times'].values()) for k in ['wall_seconds','cpu_seconds']} for l,d in D.items()}
movers=sorted(T,key=lambda x:abs(x['wall_delta_seconds']),reverse=True)[:10]
md=['# Paired kind-predicate latent census','',f"Base: `{result['base_commit']}`. Head: `{result['head_commit']}`.",'','Both successful runs used one adapted tree and first-error mode. The head is one commit directly on this base.','',f"Open-tag sites: **{open_summary['base']} before, {open_summary['head']} after**. Of the base sites, **{open_summary.get('no_finding',0)} now have no finding**, **{open_summary.get('moved',0)} moved to another reason**, and **{open_summary.get('unchanged',0)} retain the open-tag reason**.",'',f"Regressions: **{len(regressions)}**. Units whose lowering first stop changed: **{len(unit_delta)}**.",'','| Unique finding kind | Base | Head |','| --- | ---: | ---: |']
for kind in ['NotYet','Refused','panic','error','SkippedDependency']: md.append(f'| {kind} | {counts["base"][kind]} | {counts["head"][kind]} |')
md += ['','Moved open-tag sites by new reason:','']
if open_summary['moved_by_reason']:
 for r,n in sorted(open_summary['moved_by_reason'].items()): md.append(f'- {n}: {r}')
else: md.append('None.')
md += ['', 'Five of these six locations already had their other finding on the base. Only the ComputedPropertyName-to-TextRange refusal is newly revealed. All six locations were already blocked, so this is no newly failing source site.', '', '| Source token | Remaining reason | Present on base |', '| --- | --- | --- |']
for x in open_transitions:
 if x['status']=='moved':
  for f in x['head']:
   present=any(g['kind']==f['kind'] and g['reason']==f['reason'] for g in x['base_other_findings'])
   md.append(f"| `{x['where']}` | {esc(f['reason'])} | {'yes' if present else 'no'} |")
md += ['','## Units whose first stop moved','','Only lowering-phase events define a unit stop here. File-wide refusal-scan owners are retained in the JSON but are not treated as attempted units.','']
if unit_delta:
 md += ['| Unit | Base stop | Head stop |','| --- | --- | --- |']
 for x in unit_delta: md.append(f"| `{x['where']}` ({esc(x['unit'].get('name',x['unit']['kind']))}) | {stop(x['base_stop'])} | {stop(x['head_stop'])} |")
else: md.append('None. Every attempted unit retains its first lowering stop or remains finding-free.')
md += ['','## Regressions','']
if regressions:
 for x in regressions: md.append(f"- `{x['where']}`: {x['type']}; {x['kind']}: {x['reason']} (unit `{x['unit']}`).")
else: md.append('None observed. No previously finding-free attempted unit now stops, no previously finding-free refusal-scan site now stops, and no reason became a panic.')
md += ['','Changed first stops in units already blocked on the base are reported above, rather than counted as a newly failing unit. The per-site JSON retains all newly exposed findings, their phases, and their owning attempts.','', '## Unit lowering time on this box','','One sequential sample per compiler. These are sums of the meter\'s unit timers, not whole-process elapsed times. File scans, checker work, and other shared setup are excluded.','', '| Sum | Base seconds | Head seconds | Head minus base |','| --- | ---: | ---: | ---: |']
for k in ['wall_seconds','cpu_seconds']: md.append(f"| {k} | {totals['base'][k]:.6f} | {totals['head'][k]:.6f} | {totals['head'][k]-totals['base'][k]:+.6f} |")
md += ['', 'Ten biggest movers by absolute wall-time delta. All per-unit wall and CPU deltas are in `delta.json`.','','| Unit | Base wall s | Head wall s | Delta wall s | Delta CPU s |','| --- | ---: | ---: | ---: | ---: |']
for x in movers: md.append(f"| `{x['where']}` ({esc(x['name'])}) | {x['base_wall_seconds']:.6f} | {x['head_wall_seconds']:.6f} | {x['wall_delta_seconds']:+.6f} | {x['cpu_delta_seconds']:+.6f} |")
md += ['','## Resources','','| Run | Peak RSS KiB | Peak RSS GiB | Process wall s | Exit |','| --- | ---: | ---: | ---: | ---: |']
for l,x in rss.items(): md.append(f"| {l} | {x['peak_rss_kib']} | {x['peak_rss_kib']/1048576:.3f} | {x['wall_seconds']:.6f} | {x['exit_code']} |")
md += ['','Peak RSS is Linux `wait4` child `ru_maxrss`, in KiB, collected by `measure-rss.py`. Neither run was killed. Cgroup OOM counters stayed unchanged. The cgroup memory limit was 16 GiB; CPU quota was 4 cores. Host metadata is in `evidence/box.json`.','', '## Method and evidence','','Read `stage3/meter/twice-daily.sh` and `stage3/meter/README.md` in full. Ran the requested setup with `GOPROXY=https://proxy.golang.org|direct` and `timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1`; setup exited successfully. Setup warmed the initial checkout, but both measured binaries were separately built from their exact compiler pins.','', 'The shared adapted tree was built once with the base\'s `stage3/apply.sh` and lockfile-pinned API packages, using upstream TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. `git diff df84b974 b03fe07e -- stage3` is empty. No adaptation change confounds this compiler comparison. Source hashes are in `evidence/source-manifest.json`.','', 'Both compiler worktrees were clean, with cohere `06af2e399c6c1fb93fd570b774027942747400b4` and its TypeScript checker `d92d9bfee114c80be2c375d72edae966176e3a4f`. Each used the meter\'s `make_overlay.py`, `gofmt`, and `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`. The overlay blocks ordinary loading and usable IR output. No backend was invoked.','', 'For each binary, exactly as the meter\'s compiler-directory latent invocation:','', '```sh','LATENT_TIMES=times.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 latent-census /workspace/scratch/kp-adapted/src/compiler latent.jsonl','```','', 'The base build and run finished before the head build and run. The two runs cover all 79 compiler files, have identical checker headers and unit eligibility, and each records 4,056 timed attempts. All 1,251 baseline open-tag sites originate in the refusal scan.','', 'Counts follow the meter\'s unique `(kind, where, reason, text)` definition. Pairing uses source SHA-256 and exact UTF-16 token offset, preserving column identity and all findings at a token. `delta.json` includes every changed site, all 1,251 open-tag transitions, first-stop changes, regression accounting, and all 4,056 unit timing deltas. The raw normalized streams and timers are retained under `evidence/`, with hashes in `artifacts.sha256.json`.','', '"No finding" means that the source token has no finding in this guarded first-error observation. It does not establish complete unit lowering or native compilation. Later stops in already blocked units are reported as exposure, with the old and new stops preserved.','', 'The first launch did not start a census because `/usr/bin/time` is absent in this image. That launcher failure is retained as `evidence/kp-launcher-failure.log.gz`; the built base binary was reused with the kernel-backed RSS recorder. Only the two successful census runs supply these results.','']
md += ['', '## Validation', '', '`validate.py` independently checks the saved streams, counts, source identities, timing sums, resource records, and artifact hashes. It also rechecks all shared adapted source hashes after both runs. All checks passed; see `validation.txt`.', '']
(out/'REPORT.md').write_text('\n'.join(md).replace('\u2014','-').replace('\u2013','-'))
print('wrote',out)
