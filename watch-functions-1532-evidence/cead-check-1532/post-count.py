from pathlib import Path
import json,os,subprocess,time,shutil,collections
s=Path('/tmp/watch-file-check-cead');tree=Path('/tmp/watch-functions-1532/main-tree')
while not (s/'count-exit.txt').exists() or (tree/'src/compiler/sys.ts').read_bytes()!=(s/'sys-main.ts').read_bytes():time.sleep(1)
r=json.loads((s/'count-oracle/report.json').read_text());assert r['counts']=={'passing':106366,'failing':1,'pending':0};assert r['runners']=='all' and r['tests'] is None and r['workers']==8
assert [f['title'] for f in r['failures']]==['unittests:: Public APIs for typescript.d.ts should be acknowledged when they change']
assert r['baseline_diffs']==['api/typescript.d.ts']
p=s/'dynamic-lane.jsonl';rows=[json.loads(l) for l in p.read_text().splitlines()] if p.exists() else [];c=collections.Counter(x['interval'] for x in rows)
h={'250':c.pop('250',0),'500':c.pop('500',0),'2000':c.pop('2000',0),'undefined':c.pop('undefined',0),'others':dict(c),'total':len(rows)};(s/'histogram.json').write_text(json.dumps(h,indent=2)+'\n');assert not h['others'] and h['undefined']==0
if not p.exists():p.write_text('')
env=dict(os.environ,ADAMIC_DYNAMIC_COUNT=str(s/'dynamic-control.jsonl'),TMPDIR=str(s/'tmp'),npm_config_cache='/tmp/watch-functions-1532/npm-cache',PYTHONDONTWRITEBYTECODE='1')
with (s/'control.json').open('w') as log:subprocess.run(['node',str(s/'dynamic-control.cjs')],env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
with (s/'baseline-build-dts.log').open('w') as log:
 for cmd in [['npm','run','build:compiler'],['node_modules/.bin/hereby','dts']]:subprocess.run(cmd,cwd=tree,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
shutil.copyfile(tree/'built/local/typescript.js',s/'baseline-typescript.cjs');dest=s/'baseline-declarations';dest.mkdir()
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:shutil.copyfile(tree/'built/local'/name,dest/name)
(s/'COUNT_OK').write_text('count oracle, control and pristine baseline build/dts complete\n')
