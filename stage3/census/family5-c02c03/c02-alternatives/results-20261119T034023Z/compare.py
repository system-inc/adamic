import json,sys,re,collections,difflib,bisect
from pathlib import Path
base,head,out=map(Path,sys.argv[1:4]);groups=set(sys.argv[4:]);scout=json.loads(Path('/workspace/family5-analysis/sites.json').read_text());reasons={x['reason'] for x in scout if x['group'] in groups}
equivalent_reasons={'an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type'} if 'G02' in groups else set()
causes=reasons|equivalent_reasons
_maps={}
def map_position(file,line,col):
 if file not in _maps:
  rel=Path(file).relative_to('/workspace/family5-alt-c');aa=(Path('/tmp/family5-base-tree')/rel).read_bytes();bb=Path(file).read_bytes();lines=aa.splitlines(keepends=True);edits=json.loads(Path('/tmp/family5-c02-alternatives/candidates/c/edits.json').read_text());inserts=[]
  for e in edits:
   if e['file']==str(rel):
    offset=sum(map(len,lines[:e['line']-1]))+e['column']-1+len(e['before'].encode());inserts.append((offset,b' satisfies SourceFileLike'))
  if rel.name!='utilities.ts' and any(e['file']==str(rel) for e in edits):
   assert aa.startswith(b'import {'),file
   nl=b'\r\n' if b'\r\n' in aa else b'\n';inserts.append((len(b'import {'),nl+b'    type SourceFileLike,'))
  rebuilt=aa
  for offset,text in sorted(inserts,reverse=True):rebuilt=rebuilt[:offset]+text+rebuilt[offset:]
  assert rebuilt==bb,'unmapped scratch source drift: '+file
  al=[0]+[i+1 for i,v in enumerate(aa) if v==10];bl=[0]+[i+1 for i,v in enumerate(bb) if v==10];_maps[file]=(sorted(inserts),al,bl)
 inserts,al,bl=_maps[file];offset=bl[int(line)-1]+int(col)-1;delta=0;old=None
 for anchor,text in inserts:
  start=anchor+delta
  if offset<start:break
  if offset<start+len(text):old=anchor;break
  delta+=len(text)
 if old is None:old=offset-delta
 row=bisect.bisect_right(al,old)-1;return str(row+1),str(old-al[row]+1)
def normalized(s):
 def position(m):
  line,col=map_position(m[1],m[2],m[3]);return m[1]+':'+line+':'+col
 s=re.sub(r'(/workspace/family5-alt-c/src/[^:\s]+):(\d+):(\d+)',position,s)
 return re.sub(r'/tmp/family5-base-tree|/workspace/family5-c0[23]-tree|/workspace/family5-alt-[abc]','TREE',s)
def stripped(s):
 return re.sub(r':\d+:\d+','',normalized(s))
def key(f):return tuple(stripped(f[k]) for k in ['kind','phase','reason','text'])
def read(p):
 rows=[json.loads(l) for l in p.read_text().splitlines()];assert rows[-1].get('status')=='relation_report',p;assert not rows[0].get('checker_rejected'),p
 return [f for r in rows for f in r.get('findings',[])]
allchanges=[];counts={}
for bfile in sorted(base.glob('*/latent.jsonl')):
 name=bfile.parent.name;b=read(bfile);h=read(head/name/'latent.jsonl');bg=collections.defaultdict(list);hg=collections.defaultdict(list)
 for f in b:bg[key(f)].append(f)
 for f in h:hg[key(f)].append(f)
 ub=[];uh=[]
 for k in sorted(set(bg)|set(hg)):
  bs=bg[k][:];hs=hg[k][:]
  for specificity in [2,1]:
   rem=[]
   for f in bs:
    i=next((i for i,x in enumerate(hs) if normalized(x.get('unit',''))==normalized(f.get('unit','')) and (specificity==1 or normalized(x['where'])==normalized(f['where']))),None)
    if i is None:rem.append(f)
    else:hs.pop(i)
   bs=rem
  n=min(len(bs),len(hs));ub+=bs[n:];uh+=hs[n:]
 changes=[]
 for f in ub:
  i=next((i for i,x in enumerate(uh) if normalized(x['where'])==normalized(f['where']) and x['phase']==f['phase'] and normalized(x.get('unit',''))==normalized(f.get('unit',''))),None)
  if i is None and f['phase']!='refusal_scan' and f['kind']!='Boundary':
   i=next((i for i,x in enumerate(uh) if x['phase']==f['phase'] and x['kind']!='Boundary' and normalized(x.get('unit',''))==normalized(f.get('unit',''))),None)
  if i is None:changes.append({'grade':'gone','before':f})
  else:
   cleared=(f['reason'] in reasons or (f['kind']=='Boundary' and any(r in f['text'] for r in reasons))) and not (uh[i]['reason'] in equivalent_reasons or any(r in uh[i].get('text','') for r in equivalent_reasons))
   changes.append({'grade':'reached' if cleared else 'changed','before':f,'after':uh.pop(i),'previous_stop_cleared':cleared})
 for f in uh:
  blocked=[x for x in b if x.get('unit') and normalized(x['unit'])==normalized(f.get('unit','')) and x['kind']!='Boundary']
  proof=None
  try:
   file,line,col=f['where'].rsplit(':',2);lines=Path(file).read_bytes().splitlines(keepends=True);offset=sum(map(len,lines[:int(line)-1]))+len(lines[int(line)-1].decode()[:int(col)-1].encode())
   proof=next(({'boundary':x} for x in b if f['phase'] in ['lowering','deferred','registration','prepass','tail'] and x['kind']=='Boundary' and any(r in x.get('text','') for r in reasons) and normalized(x.get('unit',''))==normalized(f.get('unit','')) and normalized(x['where'].rsplit(':',2)[0])==normalized(file) and x.get('start',-1)<=offset<x.get('end',-1)),None)
  except (ValueError,IndexError,FileNotFoundError):pass
  if proof is None and f['phase']=='deferred':
   earlier=next((x for x in blocked if x['phase'] in ['refusal_scan','contract_scan'] and x['reason'] in reasons),None)
   if earlier and not any(x['reason'] in causes and normalized(x.get('unit',''))==normalized(f.get('unit','')) for x in h):proof={'prior_unit_refusal':earlier}
  changes.append({'grade':'reached' if proof else 'new','after':f,'reach_proof':proof,'previously_blocked_unit':blocked[0] if blocked else None})
 targetb=[f for f in b if f['reason'] in reasons];targeth=[f for f in h if f['reason'] in reasons];mentions=[f for f in h if any(r in f.get('reason','') or r in f.get('text','') for r in causes)]
 counts[name]={'before_events':len(b),'after_events':len(h),'target_before_distinct_sites':len({(f['where'],f['reason']) for f in targetb}),'target_after_distinct_sites':len({(f['where'],f['reason']) for f in targeth}),'target_after_equivalent_fact_distinct_sites':len({f['where'] for f in h if f['reason'] in causes}),'equivalent_relation_proof_distinct_sites':len({f['where'] for f in h if f['reason'] in equivalent_reasons}),'after_target_mentions_all_kinds':len(mentions),'grades':dict(collections.Counter(x['grade'] for x in changes))}
 targets={normalized(f['where']) for f in targetb}
 def in_targets(f):
  if normalized(f.get('where','')) in targets:return True
  m=re.match(r'(.+:\d+:\d+): (?:Adamic|stage 0)',f.get('text',''))
  return bool(m and normalized(m[1]) in targets)
 for x in changes:x.update(file=name,within_original_target_sites=any(in_targets(x[k]) for k in ['before','after'] if k in x))
 allchanges+=changes
result={'files':counts,'changes':allchanges,'groups':sorted(groups),'target_reasons':sorted(reasons),'equivalent_relation_proof_reasons':sorted(equivalent_reasons)}
out.write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(counts,indent=2))
new=[x for x in allchanges if x['grade'] in ['new','changed']]
remaining=sum(x['after_target_mentions_all_kinds'] for x in counts.values())
print(('FAIL' if new or remaining else 'PASS')+f' comparison: new_or_changed={len(new)}, target_remaining={remaining}')
sys.exit(bool(new or remaining))
