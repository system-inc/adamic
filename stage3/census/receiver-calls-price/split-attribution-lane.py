"""Attribute shared-function metadata reads; improve creation-time observation."""
import pathlib,subprocess,sys,os,json,shutil,hashlib,collections
repo=pathlib.Path('/workspace/scratch/receiver-calls-price');scratch=repo.parent;out=scratch/'receiver-split-results';tree=scratch/'receiver-split-counting-attribution';source=scratch/'receiver-split';scripts=repo/'stage3/census/receiver-calls-price'
os.environ.update(GOTMPDIR=str(scratch/'receiver-tmp'),TMPDIR=str(scratch/'receiver-tmp'),GOCACHE=str(scratch/'receiver-gocache'),npm_config_cache=str(scratch/'receiver-npm-cache'))
def run(name,cmd,cwd=repo,env=None,allow=()):
 code=subprocess.run([sys.executable,str(scratch/'receiver-measure.py'),'--out',str(out/name),'--cwd',str(cwd),'--',*map(str,cmd)],env=env).returncode;print(json.dumps({'phase':name,'exit':code}),flush=True)
 if code and code not in allow:raise SystemExit(code)
if not tree.exists():shutil.copytree(source,tree,ignore=lambda directory,names:[n for n in names if pathlib.Path(directory)==source and n in ('node_modules','built','.parallelperf.json')])
run('instrument-attribution',['node',scripts/'split-instrument-attribution.cjs',tree,scratch/'receiver-heat.json',out/'instrument-attribution-inventory.json'])
run('stock-attribution',['node',repo/'stage3/api/node_modules/typescript/lib/tsc.js','--project','src/compiler/tsconfig.json','--noEmit','--strictFunctionTypes','true','--incremental','false','--composite','false','--pretty','false'],tree)
inventory={}
for folder in ['tests/cases','tests/baselines/reference']:
 a={str(p.relative_to(source)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (source/folder).rglob('*') if p.is_file()};b={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/folder).rglob('*') if p.is_file()};assert a==b
 inventory[folder]={'files':len(a),'identical':True,'inventorySha256':hashlib.sha256(json.dumps(a,sort_keys=True).encode()).hexdigest()}
(out/'test-inputs-attribution.json').write_text(json.dumps(inventory,indent=2)+'\n')
(out/'attribution-script-hashes.json').write_text(json.dumps({f:hashlib.sha256((scripts/f).read_bytes()).hexdigest() for f in ['split-instrument-attribution.cjs','split-hooks-attribution.cjs','split-aggregate.cjs']},indent=2)+'\n')
env=dict(os.environ,NODE_OPTIONS='--require='+str(scripts/'split-hooks-attribution.cjs'),RECEIVER_PRICE_DIR=str(out/'snapshots-attribution'))
run('counting-lane-attribution',[sys.executable,repo/'stage3/oracle/run.py',tree,out/'oracle-attribution'],env=env,allow=(1,))
r=json.loads((out/'oracle-attribution/report.json').read_text());assert r['counts']=={'passing':106366,'failing':1,'pending':0} and r['baseline_diffs']==['api/typescript.d.ts'],r
run('aggregate-attribution',['node',scripts/'split-aggregate.cjs',out/'snapshots-attribution',out/'counts-attribution-full.json'])
length=collections.Counter();samples={};late={}
for p in (out/'snapshots-attribution').glob('*.json'):
 d=json.loads(p.read_text());length.update({k:v for k,v in d['counts'].items() if k.startswith('lengthAttribution.')});samples.update(d.get('lengthSamples',{}));late.update(d.get('lateSamples',{}))
result={'lengthReadsByImplementation':dict(length),'lengthReadSamples':samples,'lateObservationSamples':late};(out/'metadata-attribution.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
old=json.loads((out/'counts-valid-full.json').read_text())['totals'];new=json.loads((out/'counts-attribution-full.json').read_text())['totals']
for k in ['parserPropertyReads','transformerArrowCalls','printerAliases','printerEntries']:assert old[k]==new[k],(k,old[k],new[k])
run('stock79-final-audit',['node',scripts/'split-final-audit.cjs'])
