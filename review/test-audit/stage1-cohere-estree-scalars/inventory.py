import subprocess,os,json,re
from pathlib import Path
p=Path(__file__).resolve().parent
scratch=Path('/tmp/u088');scratch.mkdir(exist_ok=True)
coverage=scratch/'coverage';coverage.mkdir(exist_ok=True)
cases=[('eof-type.ts','type X = {'),('eof-interface.ts','interface I {'),('eof-method.ts','type X = { m(a: string): void;')]+[(f.stem,f.read_text()) for f in Path('stage1/cohere/estree/validation/followup/stalls').glob('*.input')]
env=os.environ.copy();env['NODE_V8_COVERAGE']=str(coverage)
for i,(name,text) in enumerate(cases):
 f=scratch/name;f.write_text(text)
 with (p/f'inventory-{i:02d}.log').open('w') as out:subprocess.run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs','stage1/cohere/estree/main.ts',str(f)],env=env,stdout=out,stderr=subprocess.STDOUT,timeout=10)
functions=set()
for f in coverage.glob('*.json'):
 for item in json.loads(f.read_text())['result']:
  if '/stage1/' not in item['url']:continue
  for fn in item['functions']:
   if fn['functionName'] and any(r['count']>0 for r in fn['ranges']):functions.add((item['url'],fn['functionName'],fn['ranges'][0]['startOffset']))
(p/'reached-functions.json').write_text(json.dumps(sorted(functions),indent=2))
print(len(functions),'named V8 functions reached')
