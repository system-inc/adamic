import subprocess,collections,re,json
from pathlib import Path
run=lambda *args:subprocess.check_output(['git',*args],text=True)
counts=collections.Counter()
for name in ['compiler-all.counts','repository-all.counts']:
 for line in (Path('stage1/cohere/typeaware/validation-volume')/name).read_text().splitlines():
  parts=line.split('\t')
  if len(parts)==2:counts[parts[0]]+=int(parts[1])
refs=run('for-each-ref','--format=%(refname:short)','refs/remotes/origin').splitlines()
claims={};seen=set();claimfiles=[]
for ref in refs:
 for line in run('ls-tree','-r',ref,'stage1/cohere/typeaware/claims/').splitlines():
  _,kind,objpath=line.split(' ',2);oid,path=objpath.split('\t',1)
  if kind!='blob' or oid in seen or not path.endswith('.md'):continue
  seen.add(oid);txt=run('cat-file','blob',oid);claimfiles.append(dict(ref=ref,path=path,oid=oid,text=txt))
  for name in counts:
   if re.search(r'(?<![\w@/.-])'+re.escape(name)+r'(?![\w/.-])',txt):claims.setdefault(name,[]).append(ref+':'+path)
ported={}
for ref in ['origin/codex/tsgo-c-library','origin/main']:
 for path in run('ls-tree','-r','--name-only',ref,'stage1/cohere').splitlines():
  if not path.endswith(('.a','.ts')) or '/testdata/' in path or '/gaps/' in path:continue
  txt=run('show',ref+':'+path)
  for name in counts:
   variants=[name]
   if '/typeaware/' in path and name.startswith('@typescript-eslint/'):variants.append(name.split('/',1)[1])
   if any(re.search(r'[\'"`]'+re.escape(v)+r'[\'"`]',txt) for v in variants):ported.setdefault(name,[]).append(ref+':'+path)
rank=sorted(counts,key=lambda n:(-counts[n],n))
remaining=[n for n in rank if n not in claims and n not in ported]
result=dict(total=len(counts),refs=len(refs),claim_blobs=len(claimfiles),claimed=len(claims),ported=len(ported),remaining=remaining,selected=remaining[:3],bases={ref:run('rev-parse',ref).strip() for ref in ['origin/main','origin/codex/tsgo-c-library']},rank=[dict(rule=n,total=counts[n],ported=ported.get(n,[]),claimed=claims.get(n,[])) for n in rank])
Path('/tmp/wave-18-next3-selection.json').write_text(json.dumps(result,indent=2))
Path('/tmp/wave-18-next3-claims.json').write_text(json.dumps(claimfiles,indent=2))
print(json.dumps({k:v for k,v in result.items() if k not in ['rank','remaining']},indent=2))
