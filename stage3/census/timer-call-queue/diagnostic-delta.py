from pathlib import Path
import json,difflib,re,collections
p=Path(__file__).parent
for delta in p.glob('delta.*.json'):
 d=json.loads(delta.read_text());old=d['old'];new=d['new'];file=d['file'];name=('server-'if file.startswith('server/')else'')+Path(file).stem
 a=json.loads((p/f'census.{old}.{name}.jsonl').read_text().splitlines()[0]);b=json.loads((p/f'census.{new}.{name}.jsonl').read_text().splitlines()[0]);maps={}
 for f in (p/f'sources.{old}').rglob('*.ts'):
  rel=f.relative_to(p/f'sources.{old}');g=p/f'sources.{new}'/rel
  if not g.exists():continue
  x=f.read_text().splitlines();y=g.read_text().splitlines();maps[str(rel)]={j+1:i+1 for tag,i1,i2,j1,j2 in difflib.SequenceMatcher(None,x,y,autojunk=False).get_opcodes()if tag=='equal'for i,j in zip(range(i1,i2),range(j1,j2))}
 def key(text,shift):
  def sub(m):
   f,line,column=m.groups();n=int(line);n=maps[f].get(n,'edited'+line)if shift and f in maps else n;return f+':'+str(n)+':'+column
  return re.sub(r'[^\s]*?/src/([^:\s]+):(\d+):(\d+)',sub,text)
 ca=collections.Counter(key(x,False)for x in a['diagnostics']);cb=collections.Counter(key(x,True)for x in b['diagnostics'])
 r={'old_count':len(a['diagnostics']),'new_count':len(b['diagnostics']),'gone':list((ca-cb).elements()),'new':list((cb-ca).elements()),'normalization':'Unchanged source lines map to old coordinates; diagnostics compare as a multiset of full text.'};d['checker_diagnostics']=r;delta.write_text(json.dumps(d,indent=2)+'\n');print(delta.name,r['old_count'],r['new_count'],'gone',len(r['gone']),'new',len(r['new']))
