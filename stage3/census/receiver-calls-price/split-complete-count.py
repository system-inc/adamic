"""Close derived-PrintHandlers creation coverage, then recount the whole lane."""
import pathlib,subprocess,sys,os,json,shutil,hashlib,collections
repo=pathlib.Path('/workspace/scratch/receiver-calls-price');scratch=repo.parent;out=scratch/'receiver-split-results';source=scratch/'receiver-split';tree=scratch/'receiver-split-counting-complete';scripts=repo/'stage3/census/receiver-calls-price'
os.environ.update(GOTMPDIR=str(scratch/'receiver-tmp'),TMPDIR=str(scratch/'receiver-tmp'),GOCACHE=str(scratch/'receiver-gocache'),npm_config_cache=str(scratch/'receiver-npm-cache'))
def run(name,cmd,cwd=repo,env=None,allow=()):
 code=subprocess.run([sys.executable,str(scratch/'receiver-measure.py'),'--out',str(out/name),'--cwd',str(cwd),'--',*map(str,cmd)],env=env).returncode;print(json.dumps({'phase':name,'exit':code}),flush=True)
 if code and code not in allow:raise SystemExit(code)
shutil.copytree(source,tree,ignore=lambda directory,names:[n for n in names if pathlib.Path(directory)==source and n in ('node_modules','built','.parallelperf.json')])
run('instrument-complete',['node',scripts/'split-instrument-complete.cjs',tree,scratch/'receiver-heat.json',out/'instrument-complete-inventory.json'])
run('install-complete',['npm','ci','--no-audit','--no-fund'],tree)
run('stock-complete',['node',repo/'stage3/api/node_modules/typescript/lib/tsc.js','--project','src/compiler/tsconfig.json','--noEmit','--strictFunctionTypes','true','--incremental','false','--composite','false','--pretty','false'],tree)
inventory={}
for folder in ['tests/cases','tests/baselines/reference']:
 a={str(p.relative_to(source)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (source/folder).rglob('*') if p.is_file()};b={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/folder).rglob('*') if p.is_file()};assert a==b;inventory[folder]={'files':len(a),'identical':True,'inventorySha256':hashlib.sha256(json.dumps(a,sort_keys=True).encode()).hexdigest()}
(out/'test-inputs-complete.json').write_text(json.dumps(inventory,indent=2)+'\n')
env=dict(os.environ,NODE_OPTIONS='--require='+str(scripts/'split-hooks-attribution.cjs'),RECEIVER_PRICE_DIR=str(out/'snapshots-complete'))
run('counting-lane-complete',[sys.executable,repo/'stage3/oracle/run.py',tree,out/'oracle-complete'],env=env,allow=(1,))
r=json.loads((out/'oracle-complete/report.json').read_text());assert r['counts']=={'passing':106366,'failing':1,'pending':0} and r['baseline_diffs']==['api/typescript.d.ts'],r
run('aggregate-complete',['node',scripts/'split-aggregate.cjs',out/'snapshots-complete',out/'counts-complete-full.json'])
d=json.loads((out/'counts-complete-full.json').read_text());assert not any(v for k,v in d['counts'].items() if k.startswith('lateReceiver.')),d['counts'];assert not d['failures'],d['failures']
length=collections.Counter();samples={};late={}
for p in (out/'snapshots-complete').glob('*.json'):
 x=json.loads(p.read_text());length.update({k:v for k,v in x['counts'].items() if k.startswith('lengthAttribution.')});samples.update(x.get('lengthSamples',{}));late.update(x.get('lateSamples',{}))
(out/'metadata-complete.json').write_text(json.dumps({'lengthReadsByImplementation':dict(length),'lengthReadSamples':samples,'lateObservationSamples':late},indent=2)+'\n')
old=json.loads((out/'counts-valid-full.json').read_text())['totals']
for k in ['parserPropertyReads','transformerArrowCalls','printerAliases','printerEntries']:assert d['totals'][k]==old[k],(k,d['totals'][k],old[k])
run('codemod-reproduction',['node',scripts/'split-codemod.cjs',scratch/'receiver-baseline',scratch/'receiver-heat.json',out/'codemod-reproduction-result.json','--verify',source])
(out/'complete-script-hashes.json').write_text(json.dumps({f:hashlib.sha256((scripts/f).read_bytes()).hexdigest() for f in ['split-instrument-complete.cjs','split-hooks-attribution.cjs','split-aggregate.cjs','split-codemod.cjs']},indent=2)+'\n')
