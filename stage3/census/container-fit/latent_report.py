"""Validate and join completed pieces, run main's pairing tool, and render observations."""
import collections,gzip,hashlib,heapq,json,math,pathlib,re,subprocess,sys
O=pathlib.Path(__file__).resolve().parent;R=O.parents[2];T=pathlib.Path('/workspace/container-fit-adapted');E=O/'evidence';M=E/'merged';M.mkdir(exist_ok=True)
load=lambda p:[json.loads(x) for x in gzip.open(p,'rt') if x.strip()]
sha=lambda b:hashlib.sha256(b).hexdigest()
def dump(p,v):p.write_text(json.dumps(v,indent=2)+'\n')
def norm(v):
 if isinstance(v,str):return v.replace(str(T)+'/', '')
 if isinstance(v,list):return [norm(x) for x in v]
 if isinstance(v,dict):return {k:norm(x) for k,x in v.items()}
 return v
def location(w):
 p,l,c=w.rsplit(':',2);return p,int(l),int(c)
plan=json.loads((E/'plan/INDEX.json').read_text())
stock=json.loads((E/'stock-top-level.json').read_text());wanted={u for r in stock for u in r['units']}
manifest=json.loads((E/'source-manifest.json').read_text());assert manifest=={str(p.relative_to(T)):sha(p.read_bytes()) for p in sorted((T/'src').rglob('*')) if p.is_file()}
streams={};timers={};metrics={}
for label in ['base','log']:
 files={};seen=set();times=[];measures=[];header=None
 for pid in plan['pieces']:
  D=E/'latent'/label/pid;m=json.loads((D/'metrics.json').read_text());assert m['returncode']==0 and not m.get('validation_error') and m['wall_seconds']<1520
  for key,value in m.items():
   if key.endswith('_sha256'):
    with gzip.open(D/(key[:-7]+'.gz'),'rb') as f:assert sha(f.read())==value
  rows=norm(load(D/'records.jsonl.gz'));ts=norm(load(D/'times.jsonl.gz'));selection=json.loads((D/'selection.json').read_text())
  assert {u['where'] for r in rows[1:] for u in r['units']}==set(selection['expected_units'])
  assert set(t['where'] for t in ts)=={u['where'] for r in rows[1:] for u in r['units'] if u['status'] in ['attempted','panic']}
  if header is None:header=rows[0]
  assert header==rows[0]
  for row in rows[1:]:
   meta={k:v for k,v in row.items() if k not in ['units','findings']}
   if row['file'] not in files:files[row['file']]=dict(meta,units=[],findings=[])
   dest=files[row['file']];assert meta=={k:v for k,v in dest.items() if k not in ['units','findings']}
   for u in row['units']:assert u['where'] not in seen;seen.add(u['where']);dest['units'].append(u)
   existing={json.dumps(f,sort_keys=True) for f in dest['findings']}
   for f in row['findings']:
    if f['phase']=='lowering':assert f['unit'] in set(selection['expected_units'])
    identity=json.dumps(f,sort_keys=True)
    if identity not in existing:dest['findings'].append(f);existing.add(identity)
  times+=ts;measures.append(m)
 assert seen==wanted and len(files)==79 and len(times)==len({t['where'] for t in times})
 for row in files.values():row['units'].sort(key=lambda u:location(u['where']))
 streams[label]=[header]+[files[k] for k in sorted(files)];timers[label]={t['where']:t for t in times};metrics[label]=measures
 with gzip.open(M/(label+'.jsonl.gz'),'wt') as f:
  for row in streams[label]:f.write(json.dumps(row,separators=(',',':'))+'\n')
byte_checks={}
for pid in plan.get('public_head_check_pieces',['p00']):
 with gzip.open(E/'latent/log'/pid/'records.jsonl.gz','rb') as f:a=f.read()
 with gzip.open(E/'latent/head'/pid/'records.jsonl.gz','rb') as f:b=f.read()
 assert a==b,'utilitiesPublic head/log records differ in '+pid
 byte_checks[pid]=dict(equal=True,raw_records_sha256=sha(a),bytes=len(a))
dump(E/'utilitiesPublic-byte-identity.json',dict(equal=True,pieces=byte_checks,head='62ddd47969e5168e5a0e74a23498b85d144b2361',log='29de73556749a0374da691bb9599e0e024d1214a'))
subprocess.run([sys.executable,str(R/'stage3/census/pair/pair.py'),str(M/'base.jsonl.gz'),str(M/'log.jsonl.gz'),str(T),'--typescript','/workspace/container-fit-base/stage3/api/node_modules/typescript','--json',str(O/'latent-delta.json'),'--report',str(O/'latent-pair-report.md')],check=True,timeout=120)
delta=json.loads((O/'latent-delta.json').read_text());view=delta['distinct_reads']
def first(rows):
 result={}
 for row in rows[1:]:
  for f in row['findings']:
   if f['phase']=='lowering':result.setdefault(f['unit'],f)
 return result
bf,hf=first(streams['base']),first(streams['log'])
stop=lambda f:None if f is None else (f['kind'],f['where'],f['reason'])
def transition(u):
 a,b=bf.get(u),hf.get(u)
 return dict(unit=u,base=a,head=b,classification='unchanged' if stop(a)==stop(b) else 'gone' if b is None else 'moved')
reason="a Map or Set interface implemented by a value that isn't a Map or Set"
public=collections.defaultdict(list)
for u,f in bf.items():
 if f['where'].startswith('src/compiler/utilitiesPublic.ts:') and f['reason']==reason:public[f['where']].append(transition(u))
public_rows=[]
for site,rs in sorted(public.items(),key=lambda kv:location(kv[0])):
 kind='unchanged' if all(r['classification']=='unchanged' for r in rs) else 'gone' if all(r['classification']=='gone' for r in rs) else 'moved'
 public_rows.append(dict(site=site,classification=kind,units=rs))
binder=[transition(u) for u,f in sorted(bf.items(),key=lambda kv:location(kv[0])) if f['where']=='src/compiler/binder.ts:427:31']
caps=[]
for label,rows in streams.items():
 for row in rows[1:]:
  for f in row['findings']:
   if re.search(r'a fit walk past .* distinct pairs',f['reason']):caps.append(dict(label=label,file=row['file'],finding=f))
dump(O/'first-stops.json',dict(utilitiesPublic=public_rows,binder=binder))
dump(O/'cap-rows.json',caps)
blocked=[r for r in view['unit_first_stop_delta'] if r['base']]
dump(O/'blocked-first-stops.json',blocked)
common=sorted(timers['base'].keys()&timers['log'].keys(),key=location)
changes=[dict(unit=u,base=timers['base'][u],head=timers['log'][u],wall_delta=timers['log'][u]['wall_seconds']-timers['base'][u]['wall_seconds'],cpu_delta=timers['log'][u]['cpu_seconds']-timers['base'][u]['cpu_seconds']) for u in common]
time_summary=dict(common_units=len(common),only_base=sorted(timers['base'].keys()-timers['log'].keys()),only_head=sorted(timers['log'].keys()-timers['base'].keys()),sums={label:{field:sum(timers[label][u][field] for u in common) for field in ['wall_seconds','cpu_seconds']} for label in ['base','log']},biggest_increases=sorted([r for r in changes if r['wall_delta']>0],key=lambda r:-r['wall_delta'])[:20],biggest_decreases=sorted([r for r in changes if r['wall_delta']<0],key=lambda r:r['wall_delta'])[:20],units=changes)
dump(O/'latent-times.json',time_summary)
judged=[];past=[];tops=[];kept={};walk_count=0;unattributed=0;cap_walks=0
pattern=re.compile(r'^walk judged=(\d+) past=(\d+) kept=(\d+) stop=(.*)$')
for pid in plan['pieces']:
 owner=None
 with gzip.open(E/'latent/log'/pid/'pairs.log.gz','rt') as f:
  for line_number,line in enumerate(f,1):
   line=line.rstrip('\n')
   if line.startswith('scan '):owner=None;continue
   if line.startswith('unit '):owner=line[5:];continue
   m=pattern.fullmatch(line);assert m,line
   j,p,k=map(int,m.groups()[:3]);walk_count+=1;judged.append(j);past.append(p)
   if owner is None:unattributed+=1
   else:assert owner in timers['log'];kept[owner]=max(kept.get(owner,0),k)
   if 'a fit walk past ' in m[4]:cap_walks+=1
   entry=dict(piece=pid,line=line_number,unit=owner,judged=j,past=p,kept=k,stop=m[4])
   rank=(j,p,k,walk_count,entry)
   if len(tops)<5:heapq.heappush(tops,rank)
   elif rank[:4]>tops[0][:4]:heapq.heapreplace(tops,rank)
def distribution(values):
 values=sorted(values);n=len(values)
 return dict(total=sum(values),p50=values[max(0,math.ceil(n*.5)-1)],p90=values[max(0,math.ceil(n*.9)-1)],p99=values[max(0,math.ceil(n*.99)-1)],max=values[-1])
pairs=dict(walks=walk_count,judged=distribution(judged),past=distribution(past),unattributed_walks=unattributed,cap_walks=cap_walks,largest_kept=max(kept.values()),largest_kept_per_unit=dict(sorted(kept.items(),key=lambda kv:-kv[1])),top_five=[r[-1] for r in sorted(tops,key=lambda r:r[:4],reverse=True)],percentiles='nearest rank across all logged walks, including zero judged/past')
dump(O/'pair-counts.json',pairs)
cell=lambda x:str(x).replace('|','\\|').replace('\n',' ')
def desc(f):return 'none' if f is None else f"{f['kind']} {f['where']}: {f['reason']}"
lines=['# Container fit census (zrnkpdz)','','Base d4c6b07a versus A plus statecopy 62ddd47969e5168e5a0e74a23498b85d144b2361. Logging measurements use 29de73556749a0374da691bb9599e0e024d1214a, the identical statecopy patch on f913c480. Report only; no optimizer feature is landed.','','## Method and coverage','','One adapted tree was produced by base stage3/apply.sh from TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. All 746-style source inventory entries are checked by source-manifest.json before each piece and again here; no adaptation was rerun. The actual inventory count is '+str(len(manifest))+'.','','The fixed-array rule copies a fixed array by value only when its element copier returns the unchanged value. Struct/map/slice/IR elements retain the existing refusal. containerFits stays an ordinary owned snapshot map, not a programFact. The named struct-array mutant turns the guard red; the restored full latent Go test subtree passes; all three meter overlays compile. Evidence and the one-commit implementation are linked through evidence/latent-build and the statecopy branch.','','The frozen plan has '+str(len(plan['pieces']))+' pieces, 79 files and '+str(len(wanted))+' units. Each unit occurs exactly once. utilities.ts and utilitiesPublic.ts use disjoint contiguous unit chunks; all other files belong to one disjoint file group each. Every child has a 1,500 second ceiling with at most 10 seconds to terminate. Every piece was committed and ordinarily pushed on exit. Each piece loads the whole checked 79-root program. First-error mode, LATENT_ASSERT_NO_OUTPUT=1, and LATENT_TIMES were enabled. No backend was run.','','The reused selector admission audit verifies mode-off identity for base, head and log, fixture split first stops, and dropped/duplicate/misassigned-work controls. Generic instantiated witnesses remain scoped to the callers in each piece, exactly as observed; they are not normalized to claim a whole-run census. Repeated file scans are deduplicated only on identical finding records; lowering owners and all unit identities are checked. No new whole latent run was started. Main stage3/census/pair pairs the joined same-context streams.','','Every completed utilitiesPublic chunk has byte-identical head and log record streams; individual raw SHA-256 values are in evidence/utilitiesPublic-byte-identity.json. The original whole-file p00 hit its 25 minute ceiling after partial timers, before any file finding record was emitted. That sizing attempt is saved and excluded from all paired findings and time sums. The single whole-file byte check could not be completed within the piece budget; every public unit is covered by the replacement chunk checks. This check excludes the separate timing streams.','','## First stops','','| utilitiesPublic Map/Set first-stop site | Classification | Head first stops in its base attempting units |','| --- | --- | --- |']
for r in public_rows:
 outcomes=collections.Counter(desc(u['head']) for u in r['units']);text='; '.join(f'{n} unit(s): {d}' for d,n in outcomes.items())
 lines.append(f'| {r["site"]} | {r["classification"]} | {cell(text)} |')
lines+=['',f'Measured {len(public_rows)} matching utilitiesPublic sites (brief reference: 20) and {len(binder)} binder roots (brief reference: 50). The reference 20 find sites and 50 ordinary binder roots were not reproduced. This table enumerates all measured public Map/Set first-stop sites, without selecting an arbitrary 20. Ordinary roots are distinct from latent attempting units: the saved 79-root ordinary runs have 73 core.ts:736:36 cast refusals, four readiness/generic stops and two accepted roots, with no binder.ts:427:31 first stops. The cause of the reference-count discrepancy is unresolved; these measurements do not establish outcomes for an unidentified reference cohort. Full unit mappings are in first-stops.json.','','| Base binder.ts:427:31 attempting unit | Classification | Head first stop and reason |','| --- | --- | --- |']
for r in binder:lines.append(f'| {r["unit"]} | {r["classification"]} | {cell(desc(r["head"]))} |')
lines+=['','## Regressions and already-blocked units','',f'Main pair tool: {len(view["regressions"])} distinct-read regressions; {len(delta["raw_findings"]["regressions"])} raw regressions. There are {len(blocked)} changed first stops in already-blocked units, kept separately in blocked-first-stops.json. Complete pairing: latent-delta.json and latent-pair-report.md.','',f'Panics: {len(view["panics"]["before"])} base; {len(view["panics"]["after"])} head.']
for r in view['regressions']:
 f=r.get('finding') or r.get('head');lines+=['',f'- {r["type"]}: {cell(desc(f))}; unit {f.get("unit","")}']
if not view['regressions']:lines+=['','No regressions under the pair tool definition. New reasons exposed in already-blocked units are not counted as finding-free-unit regressions.']
lines+=['','## Cap rows','',('**N=4096 IS TOO LOW IN THE ROWS BELOW.**' if caps else 'No cap-reason finding row was observed in either completed stream.'),f'All {len(caps)} cap rows are preserved in cap-rows.json; logged cap-stop walks: {cap_walks}.']
for r in caps:lines.append(f'- {r["label"]}: {r["finding"]["unit"]}: {cell(desc(r["finding"]))}')
lines+=['','## Pair counts','','Percentiles use nearest rank across all '+str(walk_count)+' logged walks, including zero counts. Totals count work across independent walks; they are not distinct program-wide pair counts.','','| Count | Total | p50 | p90 | p99 | Max per walk |','| --- | ---: | ---: | ---: | ---: | ---: |']
for key in ['judged','past']:
 v=pairs[key];lines.append(f'| {key} | {v["total"]} | {v["p50"]} | {v["p90"]} | {v["p99"]} | {v["max"]} |')
lines+=['',f'Largest kept per unit: {pairs["largest_kept"]}. Unattributed setup walks: {unattributed}. Every unit\'s largest kept is preserved in pair-counts.json.','','| Top judged walk unit | Piece:line | Judged | Past | Kept |','| --- | --- | ---: | ---: | ---: |']
for r in pairs['top_five']:lines.append(f'| {r["unit"]} | {r["piece"]}:{r["line"]} | {r["judged"]} | {r["past"]} | {r["kept"]} |')
lines+=['','## Time: one sample on this box','','Cloud host 6feab438e03f, CPU quota 4 cores, 16 GiB cgroup memory, Go 1.27.1 linux/amd64, Node v24.19.0. All measured children ran sequentially. Head LATENT_TIMES includes the requested log file append overhead. Timers include unit registration, synchronous dependency lowering and recovery; checker loading, scans and other shared setup are excluded. Repeated setup caused by partitioning is not a whole-run time claim.','','| Common-unit LATENT_TIMES sum | Base | Logging head |','| --- | ---: | ---: |']
for key in ['wall_seconds','cpu_seconds']:lines.append(f'| {key} | {time_summary["sums"]["base"][key]:.6f} | {time_summary["sums"]["log"][key]:.6f} |')
lines+=['',f'Common timed units: {len(common)}; base-only: {len(time_summary["only_base"])}; head-only: {len(time_summary["only_head"])}. Complete per-unit sums and differences are in latent-times.json.']
for label,key in [('20 largest wall increases','biggest_increases'),('20 largest wall decreases','biggest_decreases')]:
 lines+=['',label+':','','| Unit | Base wall s | Head wall s | Delta wall s | Delta CPU s |','| --- | ---: | ---: | ---: | ---: |']
 for r in time_summary[key]:lines.append(f'| {r["unit"]} | {r["base"]["wall_seconds"]:.6f} | {r["head"]["wall_seconds"]:.6f} | {r["wall_delta"]:+.6f} | {r["cpu_delta"]:+.6f} |')
lines+=['','The prior complete ordinary 79-root censuses use base d4c6b07a and unmodified A 1197ee88. The statecopy commit affects measurement generation only, so those observations were retained. All ordinary first stops matched, with zero ordinary regressions, fatal panics or cap rows. Ordinary process walls: 780.291357 seconds base and 726.362797 seconds head. Per-root sums: 777.552194 and 724.051171 seconds. These are also one sample on this box and are not latent unit times. See ordinary-summary.json.','','## Evidence and limits','','Raw piece records, LATENT_TIMES, append logs, process logs, binary hashes and per-piece publication checkpoints are under evidence/. Merged streams normalize only the common adapted prefix; original archives are lossless. Source hashes, the planned unit partition, exact timer coverage, all raw hashes and the utilitiesPublic byte check are independently asserted by this script before reporting. Header checker_rejected is '+str(streams['base'][0]['checker_rejected']).lower()+'; the fixed measurement label is retained verbatim. These observations are first stops in isolated latent units, not successful compilation or an exhaustive blocker list.','']
text='\n'.join(lines).replace('All 746-style source inventory entries','All source inventory entries')
assert '\u2013' not in text and '\u2014' not in text
proof=E/'whole-public-byte-identity.json'
if proof.exists():
 whole=json.loads(proof.read_text());assert whole['equal']
 text=text.replace('The single whole-file byte check could not be completed within the piece budget; every public unit is covered by the replacement chunk checks.', 'A separate supplemental whole-file head/log comparison also completed below the 30 minute process limit with byte-identical records. Its raw hashes, timings and coverage are in evidence/whole-public-byte-identity.json. Supplemental duplicate work is excluded from the primary timing sums and pair distributions. The original p00 sizing attempt remains excluded.')
else:
 text=text.replace('The single whole-file byte check could not be completed within the piece budget; every public unit is covered by the replacement chunk checks.', 'Every public unit is covered by the replacement chunk checks. The supplemental whole-file head attempt r00 also hit its hard 29 minute ceiling (1740.113 seconds); its matching logging run was not started. Whole-file byte comparison is incomplete. This intentional budget stop is archived and excluded from compiler panics, paired findings and timing sums. See evidence/whole-public-status.json.')
(O/'REPORT.md').write_text(text)
report=O/'latent-pair-report.md';report.write_text(report.read_text().replace(chr(92)*2+'|',chr(92)+'|'))
dump(O/'latent-outcome.json',dict(status='complete',statecopy='62ddd47969e5168e5a0e74a23498b85d144b2361',utilities_sites=len(public_rows),binder_roots=len(binder),regressions=len(view['regressions']),changed_blocked=len(blocked),cap_rows=len(caps),pair_counts=pairs,time_sums=time_summary['sums'],common_timed_units=len(common)))
print(json.dumps({k:v for k,v in json.loads((O/'latent-outcome.json').read_text()).items() if k!='pair_counts'},indent=2))

v=json.loads((O/'latent-outcome.json').read_text());v['whole_file_byte_comparison']='incomplete: head budget stop';v['chunk_byte_comparisons']='all seven identical';dump(O/'latent-outcome.json',v);dump(O/'outcome.json',v)
