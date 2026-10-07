import subprocess,json,re
from pathlib import Path
repo=Path('/workspace/adamic')
def git(*args):return subprocess.check_output(['git',*args],cwd=repo).decode()
counts={}
for part in ['compiler','repository']:
 for line in (repo/'stage1/cohere/typeaware/validation-volume'/f'{part}-all.counts').read_text().splitlines():
  if line.startswith('findings '):continue
  name,value=line.replace('\\t','\t').split('\t');counts.setdefault(name,{})[part]=int(value)
ranking=sorted([dict(name=name,total=sum(v.values()),**v) for name,v in counts.items()],key=lambda x:(-x['total'],x['name']))
ported=set();port_sources={}
for ref in ['origin/main','origin/codex/tsgo-c-library']:
 for file in git('ls-tree','-r','--name-only',ref,'stage1/cohere').splitlines():
  if not file.endswith(('.a','.ts')) or '/claims/' in file:continue
  text=git('show',ref+':'+file)
  for name in counts:
   candidates=[name]
   if name.startswith('@typescript-eslint/'):candidates.append(name.split('/')[-1])
   if any(re.search(r'''['"]'''+re.escape(n)+r'''['"]''',text) for n in candidates):
    ported.add(name);port_sources.setdefault(name,[]).append(ref+':'+file)
refs=git('for-each-ref','--format=%(refname:short)','refs/remotes/origin').splitlines();claims={}
for ref in refs:
 for file in git('ls-tree','-r','--name-only',ref,'stage1/cohere/typeaware/claims').splitlines():
  if not file.endswith('.md'):continue
  text=git('show',ref+':'+file)
  for name in counts:
   if re.search(r'(?<![A-Za-z0-9_/@-])'+re.escape(name)+r'(?![A-Za-z0-9_/@-])',text):claims.setdefault(name,[]).append(ref+':'+file)
remaining=[v for v in ranking if v['name'] not in ported and v['name'] not in claims]
result=dict(refs=len(refs),rules=len(counts),ported=sorted(ported),port_sources=port_sources,claimed=claims,remaining=remaining)
Path('/workspace/wave13-next-selection.json').write_text(json.dumps(result,indent=2)+'\n')
print('refs',len(refs),'rules',len(counts),'ported',len(ported),'claimed',len(claims),'remaining',len(remaining));print(json.dumps(remaining[:12],indent=2))
