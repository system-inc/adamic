"""Report the single paired entry pass, independently recount, and push one commit."""
from collections import Counter
import hashlib,json,os,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];tree=Path('/workspace/aow-adapted')
base='e3aa0bb7642163ebd004a71ba40a4657d27eab71';head='e1d221ce929ec37a4677f7ae98e076206f522634'
while not (out/'pair.status.json').exists():
 if (out/'control-failed.json').exists():raise SystemExit('pairing failed')
 time.sleep(2)
assert json.loads((out/'pair.status.json').read_text())['exit']==0
# Let the controller finish its final print and exit.
time.sleep(2)
def load(name):return [json.loads(s) for s in (out/(name+'-run.jsonl')).read_text().splitlines()]
a,b=load('parent'),load('head');d=json.loads((out/'paired.json').read_text())
assert a[0]==b[0] and a[0]['status']=='measurement' and a[0]['latent_mode']=='first-error'
expected=json.loads((out/'source-hashes.json').read_text())
for file,sha in expected.items():assert hashlib.sha256((tree/file).read_bytes()).hexdigest()==sha,file
for rows in [a,b]:
 assert len(rows)==1+len(rows[0]['sources'])
 assert {str(Path(p).relative_to(tree)) for p in rows[0]['sources']}==set(expected)
 assert {r['file'] for r in rows[1:]}==set(rows[0]['sources'])

def raw(rows):
 fs={};us={};first={}
 for r in rows[1:]:
  for u in r.get('units') or []:us[u['where']]=u
  for f in r.get('findings') or []:
   fs.setdefault((f['where'],f['kind'],f['reason'],f['text']),f)
   if f['phase']=='lowering':first.setdefault(f['unit'],f)
 return fs,us,first
bf,bu,bs=raw(a);hf,hu,hs=raw(b);assert bu.keys()==hu.keys()
assert dict(Counter(f['kind'] for f in bf.values()))==d['raw_findings']['totals']['before']
assert dict(Counter(f['kind'] for f in hf.values()))==d['raw_findings']['totals']['after']
# Independent exact-position delta recount, separate from pair.py's implementation.
base_sites={k[0] for k in bf};head_sites={k[0] for k in hf}
removed=bf.keys()-hf.keys()
assert len(hf.keys()-bf.keys())==len(d['raw_findings']['new_findings'])
assert sum(k[0] not in head_sites for k in removed)==len(d['raw_findings']['lowered'])
assert sum(k[0] in head_sites for k in removed)==len(d['raw_findings']['moved'])
assert sum(k[0] not in base_sites for k in hf.keys()-bf.keys())==len(d['raw_findings']['new_sites'])
stop=lambda f:(f['where'],f['kind'],f['reason']) if f else None
changed={u for u in bu if stop(bs.get(u))!=stop(hs.get(u))}
assert len(changed)==len(d['raw_findings']['unit_first_stop_delta'])
newly={u for u in changed if u not in bs and u in hs}
assert len(newly)==sum(r['type']=='previously_finding_free_unit' for r in d['raw_findings']['regressions'])
for side,pin,root in [('parent',base,repo),('head',head,Path('/workspace/aow-head'))]:
 p=json.loads((out/'provenance.json').read_text())[side]
 assert p['sha']==pin
 for file,sha in p['source_hashes'].items():assert hashlib.sha256((root/file).read_bytes()).hexdigest()==sha,(side,file)
 assert hashlib.sha256(Path('/workspace/aow-'+side+'-census').read_bytes()).hexdigest()==p['binary_sha256']
assert not list(out.rglob('*.go'))

def esc(s):return str(s).replace('|','&#124;').replace('\n',' ')
def loc(s):return s.split('/src/',1)[-1] if s else ''
def why(f):return f['kind']+': '+f['reason']+' at '+loc(f['where']) if f else 'finding-free'
def uname(u):return u.get('name') or u['kind']
unit_delta=d['raw_findings']['unit_first_stop_delta']
new_stops=[r for r in unit_delta if not r['base'] and r['head']]
moves=[r for r in unit_delta if r['base'] and r['head']]
cleared_units=[r for r in unit_delta if r['base'] and not r['head']]
(out/'newly-stopped.json').write_text(json.dumps(new_stops,indent=2)+'\n')
(out/'changed-first-stops.json').write_text(json.dumps(moves,indent=2)+'\n')
# Retain the independent identity, source, compiler and first-stop audit above.
new_findings = d['distinct_reads']['new_findings']
raw_new = d['raw_findings']['new_findings']
def context(f):
 file,line,column=f['where'].rsplit(':',2)
 source=Path(file).read_text().splitlines();line=int(line)
 return dict(start_line=max(1,line-2),end_line=min(len(source),line+2),
             text='\n'.join(source[max(0,line-3):line+2]),site_text=source[line-1])
def category(f):
 if f['unit'] in bs:return 'already-blocked unit'
 if f['unit'] in newly:return 'newly stopped unit'
 return 'unit without a lowering first stop'
ledger=[dict(**r,category=category(r['finding']),source=context(r['finding'])) for r in new_findings]
(out/'new-findings.json').write_text(json.dumps(ledger,indent=2)+'\n')
lines=['# Absent optional write: paired latent first-error census','',
'Task g2aj1pb. Report only. No compiler or adaptation change is committed.','',
f'Parent and apply: `{base}`. Head: `{head}`, exactly one commit on the parent. TypeScript 6.0.3 source: `050880ce59e30b356b686bd3144efe24f875ebc8`. Checker dependency: `06af2e399c6c1fb93fd570b774027942747400b4`.','',
'The head keys `absentOptionalWrite` by field name and includes `decodeJson` as an omitted-field producer. Its wording at observed new sites is quoted verbatim below.','',
'Method: [library-method-view reference](https://github.com/system-inc/adamic/blob/5f83ad9f1ea1aa994c11de2cd9a7e6e0184e742b/stage3/census/library-method-view-bbc1eeb2/REPORT.md). Each guarded entry driver and overlay is built from its own compiler SHA. Both run over the same unchanged parent apply, with `LATENT_FULL`, `LATENT_SPECULATIVE`, and `LATENT_ONLY_FILE` unset and `LATENT_ASSERT_NO_OUTPUT=1`. The entry is `src/tsc/tsc.ts`. Pairing uses the parent `stage3/census/pair` and stock TypeScript 6.0.3 outermost-access-chain normalization.','',
f"The streams have {len(a[0]['sources'])} reached source files and {len(a[0]['diagnostics'])} checker diagnostics on each side. Headers and attempted eligibility match. This is a lowering observation ledger, not proof of previous native correctness or exhaustive expression coverage. Generic deferrals, isolated unit state and first-error stopping retain the reference method limits. No full census, replay, native build or full gate ran.",'',
'## Counts','',
'| View | Cleared | New | Changed | New sites | First-stop changes | Regression entries |',
'| --- | ---: | ---: | ---: | ---: | ---: | ---: |']
for v,title in [('raw_findings','Raw'),('distinct_reads','Distinct outermost access chains')]:
 r=d[v];lines.append('| '+' | '.join(map(esc,[title,*[len(r[k]) for k in ['lowered','new_findings','moved','new_sites','unit_first_stop_delta','regressions']]]))+' |')
lines+=['','Cleared means no finding remains at that identity. Changed means a removed finding has another reason remaining there. New includes replacement reasons, so new and changed need not be disjoint. Raw counts unique exact-position keys, not event multiplicity. Distinct keys include source hash, outermost access-chain start, and exact reason. Unit first stops use lowering events only; refusal-scan events do not make a unit newly stopped.','',
'| View | Before | After |','| --- | ---: | ---: |']
for v,title in [('raw_findings','Raw'),('distinct_reads','Distinct')]:lines.append(f"| {title} | {sum(d[v]['totals']['before'].values())} | {sum(d[v]['totals']['after'].values())} |")
lines+=['',f'## Every newly stopped unit ({len(new_stops)})','']
for r in new_stops:
 f=r['head'];c=context(f)
 lines += ['### '+uname(r['unit']),'',f"Unit: `{loc(r['unit']['where'])}`. Site: `{loc(f['where'])}`.",'',f"{f['kind']}: {f['reason']}",'',f"Source lines {c['start_line']} through {c['end_line']}:",'','```typescript',c['text'],'```','']
if not new_stops:lines+=['None.','']
lines += [f'## Every changed first stop in an already-blocked unit ({len(moves)})','',
'| Unit | Unit location | Parent stop | Head stop |','| --- | --- | --- | --- |']
for r in moves:lines.append('| '+' | '.join(map(esc,[uname(r['unit']),loc(r['unit']['where']),why(r['base']),why(r['head'])]))+' |')
if not moves:lines+=['| None | | | |']
lines+=['',f'Cleared unit first stops: {len(cleared_units)}.','']
for r in cleared_units:lines.append('- `'+uname(r['unit'])+'` at `'+loc(r['unit']['where'])+'`: '+why(r['base']))
lines += ['', '## New findings grouped by reason and file','',
'Every distinct new finding is listed below, including findings in already-blocked units. The category uses the recording unit parent lowering first stop. Raw positions and all attempt contexts remain in the paired JSON and streams.','']
groups={}
for r in ledger:
 f=r['finding'];groups.setdefault(f['reason'],{}).setdefault(loc(f['where']).rsplit(':',2)[0],[]).append(r)
for reason,files in sorted(groups.items()):
 lines += ['### '+reason,'', 'Exact head wording:','', '> NotYet: '+reason,'']
 for file,rows in sorted(files.items()):
  lines += ['#### '+file,'',f'{len(rows)} distinct new findings.','',
  '| Site | Recording unit | Unit location | Phase | Category |','| --- | --- | --- | --- | --- |']
  for r in sorted(rows,key=lambda r:tuple(map(int,r['finding']['where'].rsplit(':',2)[1:]))):
   f=r['finding'];u=hu[f['unit']]
   lines.append('| '+' | '.join(map(esc,[loc(f['where']),uname(u),loc(f['unit']),f['phase'],r['category']]))+' |')
  lines+=['']
if not ledger:lines+=['None.','']
blocked=[r for r in ledger if r['category']=='already-blocked unit']
lines += [f'## New findings inside already-blocked units ({len(blocked)})','',
'| Site | Recording unit | Exact head reason |','| --- | --- | --- |']
for r in blocked:
 f=r['finding'];lines.append('| '+' | '.join(map(esc,[loc(f['where']),uname(hu[f['unit']]),f['kind']+': '+f['reason']]))+' |')
if not blocked:lines+=['| None | | |']
lines += ['', '## Panics, resources and validation','',
f"Panics: {len(d['raw_findings']['panics']['before'])} before, {len(d['raw_findings']['panics']['after'])} after.",'',
'| Run | Child exit | Wall seconds | Peak RSS MiB |','| --- | ---: | ---: | ---: |']
for side in ['parent','head']:
 st=json.loads((out/(side+'-run.status.json')).read_text());m=json.loads((out/(side+'-run.time.json')).read_text())
 lines.append(f"| {side} | {st['exit']} | {st['wall_seconds']:.3f} | {m['peak_rss_kib']/1024:.1f} |")
lines += ['', 'Peak RSS uses Linux child getrusage. Wall times include polling and overlap after a memory-aware stagger, so they are shared-host observations, not benchmarks. Runners retain each child own exit and stop it after 30 minutes without progress. All logs go to files and are never piped into head or tail. Setup, npm ci, parent apply and both builds passed. The first overlay attempt failed because the scratch workspace redundantly replaced its listed checker module. Removing that replacement fixed the workspace; the failed attempt log and status are retained. No production source changed.','',
'- [Parent raw stream](parent-run.jsonl), [head raw stream](head-run.jsonl), [paired JSON](paired.json), [paired ledger](paired.md).',
'- [Newly stopped units](newly-stopped.json), [changed first stops](changed-first-stops.json), [all distinct new findings and source context](new-findings.json).',
'- [Applied patches](patch-set.md), [frozen source hashes](source-hashes.json), [compiler and binary provenance](provenance.json), [guarded overlays](overlays/), [independent audit](audit.json).',
'- [Apply checkpoint](checkpoint-apply.json) is on a separate suffixed ref, preserving one final commit directly on e3aa0bb7 without force-pushing. Main is never pushed.',
'- The audit independently recounts raw totals and first-stop changes and verifies all adapted and compiler source hashes. Every reached source matches the frozen manifest. Go evidence filenames end in `.go.txt`.','']
report='\n'.join(lines)
assert '\u2013' not in report and '\u2014' not in report
(out/'REPORT.md').write_text(report)
(out/'audit.json').write_text(json.dumps(dict(source_files=len(expected),source_hashes_unchanged=True,compiler_hashes_verified=True,headers_equal=True,unit_identities_equal=True,raw_totals_recounted=True,unit_first_stop_changes=len(changed),newly_stopped_units=len(newly),new_findings_in_already_blocked_units=len(blocked),entry_only=True),indent=2)+'\n')
print(json.dumps(dict(newly_stopped=[dict(unit=uname(r['unit']),site=loc(r['head']['where']),reason=r['head']['reason']) for r in new_stops],counts={v:{k:len(d[v][k]) for k in ['lowered','new_findings','moved','new_sites','regressions']} for v in ['raw_findings','distinct_reads']}),indent=2),flush=True)
