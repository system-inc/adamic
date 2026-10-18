from pathlib import Path
import subprocess,json,re,time,sys
out=Path('review/compiler/loom-r-fx3-refusals'); text=(out/'packages.log').read_text(); decoder=json.JSONDecoder(); packages=[]
while text.strip():
    text=text.lstrip(); p,n=decoder.raw_decode(text);packages.append(p);text=text[n:]
module='github.com/system-inc/adamic/'
by={p['ImportPath']:p for p in packages}; affected={module+'internal/lower'}
while True:
    more={p['ImportPath'] for p in packages if set(p.get('Imports',[]))&affected}; union=affected|more
    if union==affected:break
    affected=union
readers={p['ImportPath'] for p in packages if set(p.get('TestImports',[])+p.get('XTestImports',[]))&affected}
closure=sorted(affected|readers)
(out/'consumer-inventory.json').write_text(json.dumps([p.removeprefix(module) for p in closure],indent=2)+'\n')
previous={r['package']:r['tests'] for r in json.loads(Path('/tmp/fx3-consumer-selection.json').read_text())}
special={'./internal/lower':['TestArrayNarrowingEveryRefused'], './internal/oracle':['TestPredicateMiscompileRefusals'], './internal/flow':['TestFlowCorpusUnitsCoverEveryProgram','TestFlowCorpusRemainder','TestFlowCorpusSetupIsShared']}
rows=[]; results=[]
for pkg in closure:
    path='./'+pkg.removeprefix(module); p=by[pkg]
    names=[]
    for file in p.get('TestGoFiles',[])+p.get('XTestGoFiles',[]):
        names+=re.findall(r'^func (Test\w+)\(', (Path(p['Dir'])/file).read_text(),re.M)
    names=set(names)
    chosen=special.get(path,previous.get(path,[]))
    chosen=[name for name in chosen if name in names]
    if not chosen and names:
        candidates=sorted(name for name in names if not name.startswith('TestProduct_'))
        preferred=[n for n in candidates if re.search('Gap|Compiler|Lower|Predicate|Union',n)]
        chosen=(preferred or candidates)[:1]
    rows.append(dict(package=path,tests=chosen))
(out/'consumer-selection.json').write_text(json.dumps(rows,indent=2)+'\n')
completed={r['package'] for r in json.loads((out/'consumer-results.json').read_text()) if r['exit']==0}
for row in rows:
    if row['package'] in completed:continue
    # These packages were already covered by validate.py's focused checks.
    if row['package'] in ['./internal/lower','./internal/oracle','./internal/ir','./cmd/adamic']:continue
    pattern='^('+'|'.join(row['tests'])+')$' if row['tests'] else '^$'
    args=['go','test',row['package'],'-run',pattern,'-count=1','-v','-timeout','90s','-parallel','4']
    label='consumer-final-'+row['package'][2:].replace('/','-'); began=time.monotonic()
    with (out/(label+'.log')).open('w') as log:
        try: code=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=180).returncode
        except subprocess.TimeoutExpired: code=124
    result=dict(**row,command=args,exit=code,seconds=round(time.monotonic()-began,3));results.append(result)
    (out/'consumer-final-results.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(result),flush=True)
    if code:sys.exit(code)
    output=(out/(label+'.log')).read_text()
    for name,seconds in re.findall(r'--- PASS: (\S+) \(([\d.]+)s\)',output):
        if float(seconds)>=60:raise RuntimeError('over-budget leaf '+name+' must not be rerun')
