from pathlib import Path
import subprocess,os,json,time
root=Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-lint-factory_hooks_shards';m=json.loads((out/'plan.json').read_text())[1];p=root/m['file'];s=p.read_text();p.write_text(s.replace(m['old'],m['new']));env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u108/cache/M02';regex='^TestRulesAgree(Union|_[0-9]{3})$';metrics=[]
try:
 for attempt in [1,2]:
  start=time.monotonic()
  with (out/'logs'/f'M02-RulesAgree-{attempt}.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',regex],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  metrics.append(dict(attempt=attempt,wall=time.monotonic()-start,exit=r.returncode));(out/'M02-extension-times.json').write_text(json.dumps(metrics,indent=2))
  es=[]
  for line in (out/'logs'/f'M02-RulesAgree-{attempt}.log').read_text().splitlines():
   try:es.append(json.loads(line))
   except:pass
  completed={e.get('Test') for e in es if e['Action'] in ['pass','fail'] and e.get('Test','').startswith('TestRulesAgree_')}
  if len(completed)==16:break
finally:p.write_text(s)
