import pathlib,subprocess,json
out=pathlib.Path('review/compiler/loom-r-cycle-targets-ac-m2')
def rows(s):
 result={}
 for l in s.splitlines():
  f=[x.strip() for x in l.strip('|').split('|')]
  if len(f)==7 and all(x.isdigit()for x in f[1:]):result[f[0]]=list(map(int,f[1:]))
 return result
get=lambda ref:rows(subprocess.check_output(['git','show',ref+':internal/oracle/counts.md'],text=True))
base=get('origin/main');now=rows(pathlib.Path('internal/oracle/counts.md').read_text());source=get('ca09c749');old=get('ca09c749~1');changed=[]
for name in sorted(base.keys()|now.keys()):
 if base.get(name)==now.get(name):continue
 delta=[b-a for a,b in zip(base[name],now[name])]if name in base and name in now else None
 olddelta=[b-a for a,b in zip(old[name],source[name])]if name in old and name in source else None
 changed.append(dict(path=name,main=base.get(name),branch=now.get(name),delta=delta,source_delta=olddelta,same_as_source_delta=delta==olddelta))
(out/'counts-rows.json').write_text(json.dumps(changed,indent=2)+'\n')
fmt=lambda v:'/'.join(map(str,v))if v is not None else'new'
text='Counter order: allocations/frees/retains/releases/peak/in regions.\n\n| Fixture | Main | Branch | Delta |\n|---|---|---|---|\n'
for r in changed:text+='| '+r['path']+' | '+fmt(r['main'])+' | '+fmt(r['branch'])+' | '+fmt(r['delta'])+' |\n'
(out/'counts-rows.md').write_text(text)
print(len(changed),'rows;',sum(r['main']is None for r in changed),'added')
print('Delta differs from source:',[r['path']for r in changed if not r['same_as_source_delta']])
