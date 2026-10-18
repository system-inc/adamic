from pathlib import Path
import json,re,subprocess,time,sys
out=Path('review/compiler/loom-r-fx3-refusals');decoder=json.JSONDecoder(); text=(out/'packages.log').read_text();packages={}
while text.strip():
    text=text.lstrip();p,n=decoder.raw_decode(text);text=text[n:];packages[p['ImportPath'].removeprefix('github.com/system-inc/adamic/')]=p
covered=set(json.loads((out/'consumer-inventory.json').read_text()));manifest=json.loads(Path('cloud/fast-gate/compiler-dependencies.json').read_text())['packages']; results=[]
rows=[]
for pkg,deps in manifest.items():
    if pkg in covered or not set(deps)&{'internal/lower','cmd/adamic'}:continue
    p=packages[pkg];names=[]
    for f in p.get('TestGoFiles',[])+p.get('XTestGoFiles',[]):names+=re.findall(r'^func (Test\w+)\(', (Path(p['Dir'])/f).read_text(),re.M)
    choices=sorted(n for n in names if not n.startswith('TestProduct_'))
    preferred=[n for n in choices if re.search('Compiler|Dependenc|Key|Cache|Build',n)]
    picked=(preferred or choices)[:1]
    rows.append(dict(package='./'+pkg,tests=picked))
(out/'declared-reader-selection.json').write_text(json.dumps(rows,indent=2)+'\n')
for row in rows:
    pattern='^('+'|'.join(row['tests'])+')$' if row['tests'] else '^$'
    args=['go','test',row['package'],'-run',pattern,'-count=1','-v','-timeout','90s','-parallel','4']
    began=time.monotonic();label='declared-'+row['package'][2:].replace('/','-')
    with (out/(label+'.log')).open('w') as log:r=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=180)
    results.append(dict(**row,command=args,exit=r.returncode,seconds=round(time.monotonic()-began,3)));(out/'declared-reader-results.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(results[-1]),flush=True)
    if r.returncode:sys.exit(r.returncode)
