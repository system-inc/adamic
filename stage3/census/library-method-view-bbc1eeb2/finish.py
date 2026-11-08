"""Report the single paired entry pass, independently recount, and push one commit."""
from collections import Counter
import hashlib,json,os,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];tree=Path('/tmp/lmv-adapted')
base='f7390e5b75975418eb6cbd1a8ee24e1f792fd4ca';head='bbc1eeb2900538192d08ddb45b493d703e2b79ad'
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
stop=lambda f:(f['where'],f['kind'],f['reason']) if f else None
changed={u for u in bu if stop(bs.get(u))!=stop(hs.get(u))}
assert len(changed)==len(d['raw_findings']['unit_first_stop_delta'])
newly={u for u in changed if u not in bs and u in hs}
assert len(newly)==sum(r['type']=='previously_finding_free_unit' for r in d['raw_findings']['regressions'])
for side,pin,root in [('parent',base,repo),('head',head,Path('/tmp/lmv-head'))]:
 p=json.loads((out/'provenance.json').read_text())[side]
 assert p['sha']==pin
 for file,sha in p['source_hashes'].items():assert hashlib.sha256((root/file).read_bytes()).hexdigest()==sha,(side,file)
 assert hashlib.sha256(Path('/tmp/lmv-'+side+'-census').read_bytes()).hexdigest()==p['binary_sha256']
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
# Quote every head site with the guard's diagnostic template, including retained
# Error-property instances, so no new method-view occurrence is omitted.
quotes={}
for f in hf.values():
 if f['kind']!='NotYet' or not f['reason'].startswith('a view naming ') or ', which lib declares on ' not in f['reason']:continue
 key=(f['where'],f['reason'])
 if key in quotes:continue
 file,line,column=f['where'].rsplit(':',2);line,column=int(line),int(column)
 source=Path(file).read_bytes().decode('utf8').splitlines()
 lo=max(1,line-2);hi=min(len(source),line+2)
 old=any(g['where']==f['where'] and g['kind']==f['kind'] and g['reason']==f['reason'] for g in bf.values())
 quotes[key]=dict(where=f['where'],reason=f['reason'],parent_had_exact_finding=old,unit=f['unit'],phase=f['phase'],line=line,column=column,context_start_line=lo,context_end_line=hi,source='\n'.join(source[lo-1:hi]),exact_site_line=source[line-1])
(out/'method-view-sites.json').write_text(json.dumps(list(quotes.values()),indent=2)+'\n')
lines=['# Library method view: paired latent first-error census','', 'Task sn788gx. Report only. No compiler or adaptation change is committed.','',
 '| Compiler | Pin |','| --- | --- |',f'| Parent and apply | `{base}` |',f'| Head | `{head}` |','',
 'The head adds lib-declared method signatures to `libraryUnstoredMember`. The adapted input comes from the unmodified parent apply, using TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Both compiler binaries and their guarded overlays were built from their own pins; checker dependency `06af2e399c6c1fb93fd570b774027942747400b4` is shared unchanged.','',
 f"The entry is `src/tsc/tsc.ts`, with {len(a[0]['sources'])} reached source files and {len(a[0]['diagnostics'])} checker diagnostics on each compiler. Only latent **first-error** mode ran. `LATENT_ASSERT_NO_OUTPUT=1` checked both no-output guards. No full-core mode, replay, second pairing, native build or full gate ran.",'',
 '## Counts','', '| View | Cleared findings | New findings | Changed findings | New sites | First-stop unit changes | Regression entries |','| --- | ---: | ---: | ---: | ---: | ---: | ---: |']
for view,title in [('raw_findings','Raw'),('distinct_reads','Distinct outermost access chains')]:
 r=d[view];lines.append('| '+' | '.join(map(esc,[title,*[len(r[k]) for k in ['lowered','new_findings','moved','new_sites','unit_first_stop_delta','regressions']]]))+' |')
lines+=['','Raw means pair.py exact raw-position keys, not event multiplicity. Distinct uses the stock TypeScript 6.0.3 outermost-access-chain normalization. New findings include replacement reasons, so new and changed are not disjoint. A newly stopped unit was finding-free during latent lowering before and now has a lowering stop. A different first stop in an already-blocked unit is listed separately. This is a lowering observation, not proof of previous native correctness.','',
 f'## Every newly stopped unit ({len(new_stops)})','', '| Unit | Unit location | Head stop site | Exact head reason |','| --- | --- | --- | --- |']
for r in new_stops:lines.append('| '+' | '.join(map(esc,[uname(r['unit']),loc(r['unit']['where']),loc(r['head']['where']),r['head']['kind']+': '+r['head']['reason']]))+' |')
if not new_stops:lines+=['| None | | | |']
lines+=['',f'## Every changed first stop in an already-blocked unit ({len(moves)})','', '| Unit | Unit location | Parent site and exact reason | Head site and exact reason |','| --- | --- | --- | --- |']
for r in moves:lines.append('| '+' | '.join(map(esc,[uname(r['unit']),loc(r['unit']['where']),why(r['base']),why(r['head'])]))+' |')
if not moves:lines+=['| None | | | |']
lines+=['',f'## Cleared unit first stops ({len(cleared_units)})','']
for r in cleared_units:lines.append('- `'+uname(r['unit'])+'` at `'+loc(r['unit']['where'])+'`: '+why(r['base'])+' -> finding-free.')
if not cleared_units:lines+=['None.']
lines+=['','## Head method-view NotYet sites and source text','', 'Each unique location and exact reason matching the guard template is quoted below. The JSON retains the recording unit, phase, exact site line and source context. Retained parent findings are labeled rather than attributed to the new method arm.','']
for q in sorted(quotes.values(),key=lambda q:(q['where'],q['reason'])):
 lines+=['### '+loc(q['where']),'', '**'+('Already present in parent' if q['parent_had_exact_finding'] else 'New exact finding')+'**. Unit: `'+uname(hu[q['unit']])+'` at `'+loc(q['unit'])+'`; phase `'+q['phase']+'`.','',q['reason'],'',f"Source lines {q['context_start_line']} through {q['context_end_line']}; diagnostic column {q['column']}:",'','```typescript',q['source'],'```','']
if not quotes:lines+=['No head finding matches the library-unstored diagnostic template.','']
lines+=['## Panics and regression details','']
for side in ['before','after']:lines.append(f"- {side}: {len(d['raw_findings']['panics'][side])} raw-position panic findings.")
for r in d['distinct_reads']['regressions']:
 f=r.get('finding') or r.get('head');lines.append('- '+r['type']+': '+why(f))
if not d['distinct_reads']['regressions']:lines+=['- No normalized regression entries.']
lines+=['','## Time and peak memory','', '| Run | Child exit | Wall seconds | Peak RSS MiB |','| --- | ---: | ---: | ---: |']
for side in ['parent','head']:
 s=json.loads((out/(side+'-run.status.json')).read_text());m=json.loads((out/(side+'-run.time.json')).read_text())
 lines.append(f"| {side} | {s['exit']} | {s['wall_seconds']:.3f} | {m['peak_rss_kib']/1024:.1f} |")
lines+=['','Peak RSS is Linux child `getrusage` in KiB converted to MiB. Wall observations include up to one second of polling resolution. Parent and head overlap after a memory-aware stagger ([launch record](head-start.json)); these are shared-host observations, not benchmarks. Every wrapper records its child exit, wall time and peak RSS, and stops a child after 30 minutes without output/progress. Unit timing streams provide progress even while the census buffers file records.','',
 '## Evidence and validation','',
 '- [Parent stream](parent-run.jsonl), [head stream](head-run.jsonl), [full paired JSON](paired.json), and [paired ledger](paired.md).',
 '- [Newly stopped units](newly-stopped.json), [changed first stops](changed-first-stops.json), and [method-view source quotations](method-view-sites.json).',
 '- [Applied patch set](patch-set.md), [all measured source hashes](source-hashes.json), [compiler/source/binary provenance](provenance.json), and [guarded overlays](overlays/). Go evidence uses `.go.txt` filenames.',
 '- [Parent timings](parent-run-units.jsonl), [head timings](head-run-units.jsonl), run logs and `*.status.json` / `*.time.json` files retain commands, own child exits and resources.',
 '- The independent audit below recounts raw keys, unit first stops and newly stopped units, verifies every input and compiler source hash, and checks the 81-file scope against the measurement header. No measured source was removed or changed. Only unused upstream tests were removed after apply to reserve RAM.',
 '', 'The environment and pinned checker were reused from the preceding task after verifying the dependency pin. The API packages were installed with `npm ci`. The parent apply and both own-pin builds exited 0. A sparse-checkout exclusion initially prevented staging the checkpoint directory; it was corrected before any census run. Logs are retained and never piped into head or tail.', '',
 'The apply checkpoint uses a separate suffixed ref ([record](checkpoint-apply.json)), allowing this final branch to contain one commit directly on f7390e5b without any force push. No main branch is pushed.','']
report='\n'.join(lines);assert '\u2013' not in report and '\u2014' not in report
(out/'REPORT.md').write_text(report)
(out/'audit.json').write_text(json.dumps(dict(source_files=len(expected),source_hashes_unchanged=True,compiler_hashes_verified=True,headers_equal=True,unit_identities_equal=True,raw_totals_recounted=True,unit_first_stop_changes=len(changed),newly_stopped_units=len(newly),head_template_sites=len(quotes),entry_only=True),indent=2)+'\n')
# Verify no compiler child or wrapper remains active before sealing artifacts.
active=[]
for proc in Path('/proc').iterdir():
 if not proc.name.isdigit():continue
 try:
  args=proc.joinpath('cmdline').read_bytes().decode().split('\0')
  if args[0] in ['/tmp/lmv-parent-census','/tmp/lmv-head-census']:active.append(int(proc.name))
 except (FileNotFoundError,ProcessLookupError,PermissionError,UnicodeError):pass
assert not active,active
(out/'processes-final.json').write_text(json.dumps(dict(live_compiler_processes=active),indent=2)+'\n')
import shutil
shutil.rmtree(repo/'stage3/census/pair/__pycache__',ignore_errors=True)
(out/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.relative_to(out).as_posix()+'\n' for p in sorted(out.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip()==base
subprocess.run(['git','add','-f','--',str(out)],cwd=repo,check=True)
changed_files=subprocess.check_output(['git','diff','--cached','--name-only',base],cwd=repo,text=True).splitlines()
assert all(f.startswith('stage3/census/library-method-view-bbc1eeb2/') for f in changed_files)
with Path('/tmp/lmv-final-commit.log').open('w') as log:rc=subprocess.run(['git','commit','-m','stage3/census: pair library-method-view bbc1eeb2 against f7390e5b'],cwd=repo,stdout=log,stderr=subprocess.STDOUT).returncode
assert rc==0
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip()
assert subprocess.check_output(['git','rev-parse','HEAD^'],cwd=repo,text=True).strip()==base
with Path('/tmp/lmv-final-push.log').open('w') as log:rc=subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/library-method-view-census'],cwd=repo,stdout=log,stderr=subprocess.STDOUT).returncode
assert rc==0
remote=subprocess.check_output(['git','ls-remote','--heads','origin','refs/heads/typescript/library-method-view-census'],cwd=repo,text=True).split()[0];assert remote==sha
print(json.dumps(dict(sha=sha,newly_stopped=[dict(unit=uname(r['unit']),unit_where=r['unit']['where'],site=r['head']['where'],reason=r['head']['reason']) for r in new_stops],counts={v:{k:len(d[v][k]) for k in ['lowered','new_findings','moved','new_sites','regressions']} for v in ['raw_findings','distinct_reads']}),indent=2),flush=True)
