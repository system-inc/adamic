"""Inventory actual sites and unit statuses without counting continuation boundaries as refusals."""
import os,json,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];tree=Path('/workspace/scratch/fa-adapted')
p=subprocess.run(['node',str(out/'function-spans.cjs'),str(tree)],env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules')),capture_output=True,text=True)
(out/'function-spans.log').write_text(p.stderr);print('function-spans child_exit='+str(p.returncode),flush=True);assert p.returncode==0
spans=json.loads(p.stdout);(out/'function-spans.json').write_text(p.stdout)
def loc(s):return s.split('/src/',1)[-1]
def esc(s):return str(s).replace('|','&#124;').replace('\n',' ')
def offset(where):
 file,line,col=where.rsplit(':',2);ls=Path(file).read_bytes().decode('utf8').splitlines(keepends=True);line,col=int(line),int(col)
 prefix=ls[line-1].encode('utf-16-le')[:(col-1)*2].decode('utf-16-le')
 return len((''.join(ls[:line-1])+prefix).encode())
def why(f):return f['kind']+': '+f['reason']+' at '+loc(f['where']) if f else 'none'
lines=['## Full-file inventories','','These are separate full-mode measurements on the compiler directory, selecting exactly one adapted file with LATENT_ONLY_FILE. Continuation Boundary records are retained in raw streams but are not counted as refusal sites here. A deferred generic body with no finding is unexamined, not proven to lower.','']
ledger={}
for file in ['utilitiesPublic','core']:
 data={side:[json.loads(s) for s in (out/(side+'-full-'+file+'.jsonl')).read_text().splitlines()] for side in ['parent','head']}
 assert data['parent'][0]==data['head'][0]
 ledger[file]={}
 names=['findAncestor'] if file=='utilitiesPublic' else ['orPredicate2','orPredicate3','or']
 for name in names:
  funcs=[r for r in spans if r['name']==name and r['file'].endswith('/'+file+'.ts')]
  lines+=['### '+name,'']
  for f in funcs:lines+=['```typescript',f['text'],'```','']
  lines+=['| Side | Unit | Status | Instantiated |','| --- | --- | --- | --- |']
  ledger[file][name]={}
  for side,rows in data.items():
   us=[u for row in rows[1:] for u in row.get('units') or [] if any(u['where']==f['where'] for f in funcs)]
   for u in us:lines.append('| '+' | '.join(map(esc,[side,loc(u['where']),u['status'],u.get('instantiated','')]))+' |')
   fs={}
   for row in rows[1:]:
    for finding in row.get('findings') or []:
     if finding['kind']=='Boundary':continue
     if any(finding['where'].rsplit(':',2)[0]==f['file'] and f['start']<=offset(finding['where'])<f['end'] for f in funcs):
      key=(finding['where'],finding['kind'],finding['reason'],finding['phase'])
      fs.setdefault(key,dict(finding=finding,recorded_units=[]))['recorded_units'].append(finding['unit'])
   ledger[file][name][side]=dict(units=us,sites=list(fs.values()))
  lines+=['','| Side | Every recorded site | Phase | Recording units |','| --- | --- | --- | --- |']
  for side in ['parent','head']:
   fs=ledger[file][name][side]['sites']
   for item in fs:
    f=item['finding'];lines.append('| '+' | '.join(map(esc,[side,why(f),f['phase'],', '.join(sorted(set(loc(u) for u in item['recorded_units'])))]))+' |')
   if not fs:lines.append('| '+side+' | No site finding recorded | | |')
  lines+=['']
 # Every moved actual site in this file, plus introduced or cleared ones.
 delta=json.loads((out/(file+'-paired.json')).read_text())
 changes=[]
 for category in ['lowered','moved','new_findings']:
  for r in delta['raw_findings'][category]:
   f=r.get('base') or r.get('finding')
   if f['kind']!='Boundary' and f['where'].rsplit(':',2)[0]==str(tree/'src/compiler'/(file+'.ts')):
    changes.append(dict(category=category,**r))
 ledger[file]['all_changes']=changes
 lines+=['### Every changed site in '+file+'.ts','','| Category | Old | New |','| --- | --- | --- |']
 for r in changes:
  lines.append('| '+' | '.join(map(esc,[r['category'],why(r.get('base')), ('; '.join(why(f) for f in r.get('head',[])) or 'none') if r.get('base') else why(r.get('finding'))]))+' |')
 if not changes:lines+=['| None | | |']
 lines+=['']
# Include every helper finding reached from the complete entry and focused callers.
lines+=['### Helper sites across entry and focused contexts','','| Context | Side | Helper | Site and exact reason | Phase | Recording unit |','| --- | --- | --- | --- | --- | --- |']
cross=[]
for side in ['parent','head']:
 for context,prefix in [('entry','entry-run')]+[('focus-'+str(i),'focus-'+str(i)) for i in range(1,5)]:
  rs=[json.loads(s) for s in (out/(side+'-'+prefix+'.jsonl')).read_text().splitlines()]
  seen=set()
  for row in rs[1:]:
   for f in row.get('findings') or []:
    if f['kind']=='Boundary':continue
    owners=[r for r in spans if r['name'] in ['orPredicate2','orPredicate3','or'] and f['where'].rsplit(':',2)[0]==r['file'] and r['start']<=offset(f['where'])<r['end']]
    if not owners:continue
    key=(f['kind'],f['where'],f['reason'],f['phase'],f['unit'])
    if key in seen:continue
    seen.add(key);owner=owners[0]
    cross.append(dict(context=context,side=side,helper=owner['name'],finding=f))
    lines.append('| '+' | '.join(map(esc,[context,side,owner['name'],why(f),f['phase'],loc(f['unit'])]))+' |')
if not cross:lines+=['| All | Both | orPredicate2, orPredicate3, or | No site finding recorded | | |']
lines+=[''];ledger['cross_context_helpers']=cross
(out/'full-sites.json').write_text(json.dumps(ledger,indent=2)+'\n')
(out/'full-sites.md').write_text('\n'.join(lines)+'\n')
