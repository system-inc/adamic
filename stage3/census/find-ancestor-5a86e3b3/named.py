"""Attribute each requested call to its real focused-unit boundary and quote source."""
import json
from pathlib import Path
out=Path(__file__).resolve().parent
for src,dest in [('/tmp/fa-control.log','original-control.log'),('/tmp/fa-new-control.log','control.log'),('/tmp/fa-prepare.log','prepare.log'),('/tmp/fa-fetch.log','fetch.log'),('/tmp/fa-typescript-clone.log','source-clone.log')]:
 if Path(src).exists():(out/dest).write_bytes(Path(src).read_bytes())
sites=json.loads((out/'sites.json').read_text());indexes=json.loads((out/'focus-index.json').read_text())
def loc(s):return s.split('/src/',1)[-1]
def offset(where):
 file,line,col=where.rsplit(':',2);line,col=int(line),int(col);ls=Path(file).read_bytes().decode('utf8').splitlines(keepends=True)
 prefix='';units=0
 for ch in ls[line-1]:
  if units>=col-1:break
  prefix+=ch;units+=len(ch.encode('utf-16-le'))//2
 assert units==col-1
 return sum(len(s.encode()) for s in ls[:line-1])+len(prefix.encode())
def obs(side,u,site):
 name=side+'-focus-'+str(indexes[u]);assert json.loads((out/(name+'.status.json')).read_text())['exit']==0
 rows=[json.loads(s) for s in (out/(name+'.jsonl')).read_text().splitlines()];r=rows[1];unit=r['units'][0];assert unit['where']==u
 pos=offset(site);fs=r.get('findings') or []
 bs=[f for f in fs if f['kind']=='Boundary' and f['where'].rsplit(':',2)[0]==site.rsplit(':',2)[0] and (f.get('start',0)<=pos<f.get('end',0) or (f['reason'].startswith('function signature or prologue') and f['where']==u))]
 first=next((f for f in fs if f['kind']!='Boundary' and f.get('phase')=='lowering'),None)
 if bs:
  b=min(bs,key=lambda f:f['end']-f['start']);match=[f for f in fs if f['kind']!='Boundary' and f['text']==b['text']];assert match
  return dict(run=name,unit=unit,state='unit prologue' if b['reason'].startswith('function signature or prologue') else 'enclosing statement',finding=match[0],boundary=b,first_lowering_stop=first)
 row=next((r for r in sites['checker'] if r['site']==site),None)
 if row:
  helper=next(h for h in sites['helpers'] if h['name']==row['predicate_helper'])
  candidates=[f for f in fs if f['kind']!='Boundary' and f.get('phase')=='lowering' and f['where'].rsplit(':',2)[0]==helper['file'] and helper['start']<=offset(f['where'])<helper['end']]
  if candidates:
   return dict(run=name,unit=unit,state='callee helper '+helper['name']+' stop recorded in this observing unit; no caller-local failed boundary',finding=candidates[0],first_lowering_stop=first,helper=helper,attribution='Shared helper finding; a single recorded finding can serve more than one caller of that helper in this unit. This is not independent per-call execution evidence.')
 return dict(run=name,unit=unit,state='deferred generic' if unit['status']=='deferred_generic_instantiation' else 'no boundary contains this site',finding=None,first_lowering_stop=first)
def esc(s):return str(s).replace('|','&#124;').replace('\n',' ')
def why(o):
 f=o.get('finding');return o['state']+': '+(f['kind']+': '+f['reason']+' at '+loc(f['where']) if f else 'no finding recorded; status '+o['unit']['status'])
def quote(where):
 file,l,c=where.rsplit(':',2);l=int(l);src=Path(file).read_text().splitlines();lo=max(1,l-2);hi=min(len(src),l+2)
 return '\n'.join([f'Source `{loc(file)}:{lo}` through line {hi}:','','```typescript','\n'.join(src[lo-1:hi]),'```',''])
rows=[];lines=['## Six requested checker callers, both A and B','','Focused full-mode inventories use the same replay selector as pairing A. They retain the complete entry graph and admit no usable IR. A failed statement may contain multiple calls; its boundary is not evidence of independent execution of every call. These observations are separate from the entry first-error counts.','','| Upstream line | Adapted call | Enclosing declaration | Parent boundary stop | Head boundary stop |','| ---: | --- | --- | --- | --- |']
for r in sites['checker']:
 a=obs('parent',r['unit'],r['site']);b=obs('head',r['unit'],r['site']);rows.append(dict(**r,parent=a,head=b))
 lines.append('| '+' | '.join(map(esc,[r['requested_line'],loc(r['site']),r['owners'][-1]+' at '+loc(r['unit']),why(a),why(b)]))+' |')
lines+=['','A head helper stop is a real callee finding within the selected observing unit. Where the full-mode continuation records its boundary inside core.ts rather than at the caller, the table says so. Repeated calls can share that single helper finding; no independent per-call execution or whole-unit success is claimed.','', '### Source and first stops for each caller','']
for r in rows:
 lines+=['#### Upstream '+str(r['requested_line'])+', adapted '+loc(r['site']),'','Enclosing unit `'+r['owners'][-1]+'` at `'+loc(r['unit'])+'`.','', 'Call source:','','```typescript',r['text'],'```','','Parent: '+why(r['parent'])+'.','','Head: '+why(r['head'])+'.','']
 for side in ['parent','head']:
  f=r[side]['first_lowering_stop'];lines+=['Focused unit first lowering stop ('+side+'): '+(f['kind']+': '+f['reason']+' at '+loc(f['where']) if f else 'none; status '+r[side]['unit']['status'])+'.','']
 if r['head'].get('finding'):lines+=[quote(r['head']['finding']['where'])]
 if r['head'].get('attribution'):lines+=[r['head']['attribution'],'']
 lines+=[quote(r['site'])]
lines+=['## utilitiesPublic.ts findAncestor itself, both A and B','']
impl=next(r for r in sites['findAncestor'] if r['has_body']);a=obs('parent',impl['site'],impl['site']);b=obs('head',impl['site'],impl['site'])
lines+=['Implementation at `'+loc(impl['site'])+'`.','','```typescript',impl['text'],'```','','Parent focused observation: '+why(a)+'.','','Head focused observation: '+why(b)+'.','']
# A generic declaration can be deferred even when concrete calls get past its signature.
if a['unit']['status']=='deferred_generic_instantiation' or b['unit']['status']=='deferred_generic_instantiation':lines+=['A deferred generic declaration is unexamined in this isolated observation; it is not a lowering success. Concrete caller observations above test the reached instantiations.','']
for side,o in [('parent',a),('head',b)]:
 if o.get('finding'):lines+=[side+':',quote(o['finding']['where'])]
(out/'named-sites.json').write_text(json.dumps(dict(checker=rows,findAncestor=dict(**impl,parent=a,head=b)),indent=2)+'\n')
(out/'named-sites.md').write_text('\n'.join(lines)+'\n')
