"""Retry only the corrected counting experiment after all census jobs finish."""
import pathlib,subprocess,sys,os,json,shutil,hashlib
repo=pathlib.Path('/workspace/scratch/receiver-calls-price');scratch=repo.parent;out=scratch/'receiver-split-results';split=scratch/'receiver-split';counting=scratch/'receiver-split-counting-measured';scripts=repo/'stage3/census/receiver-calls-price'
os.environ.update(GOTMPDIR=str(scratch/'receiver-tmp'),TMPDIR=str(scratch/'receiver-tmp'),GOCACHE=str(scratch/'receiver-gocache'),npm_config_cache=str(scratch/'receiver-npm-cache'))
def run(name,cmd,cwd=repo,env=None,allow=()):
 code=subprocess.run([sys.executable,str(scratch/'receiver-measure.py'),'--out',str(out/name),'--cwd',str(cwd),'--',*map(str,cmd)],env=env).returncode
 print(json.dumps({'phase':name,'exit':code}),flush=True)
 if code and code not in allow:raise SystemExit(code)
run('controls-final',['node',scripts/'split-fixture-controls.cjs'])
control=out/'control-final-snapshots';control.mkdir(exist_ok=True);shutil.copyfile(scratch/'split-fixture-controls.json',control/'control.json')
run('this-controls-final',['node',scripts/'split-aggregate.cjs',control,out/'this-controls-final-result.json'])
assert json.loads((out/'this-controls-final-result.json').read_text())['implementationThis']['reads']==1
if not counting.exists():shutil.copytree(split,counting,ignore=lambda directory,names:[n for n in names if pathlib.Path(directory)==split and n in ('node_modules','built','.parallelperf.json')])
run('instrument-final',['node',scripts/'split-instrument.cjs',counting,scratch/'receiver-heat.json',out/'instrument-final-inventory.json'])
inventory={}
for folder in ['tests/cases','tests/baselines/reference']:
 a={str(p.relative_to(split)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (split/folder).rglob('*') if p.is_file()};b={str(p.relative_to(counting)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (counting/folder).rglob('*') if p.is_file()};assert a==b
 inventory[folder]={'files':len(a),'identical':True,'inventorySha256':hashlib.sha256(json.dumps(a,sort_keys=True).encode()).hexdigest()}
(out/'test-inputs-final.json').write_text(json.dumps(inventory,indent=2)+'\n')
env=dict(os.environ,NODE_OPTIONS='--require='+str(scripts/'split-hooks.cjs'),RECEIVER_PRICE_DIR=str(out/'snapshots-measured'))
run('counting-lane-measured',[sys.executable,repo/'stage3/oracle/run.py',counting,out/'oracle-measured'],env=env,allow=(1,))
report=json.loads((out/'oracle-final/report.json').read_text());assert report['phases']['install']['exit']==0 and report['phases']['build']['exit']==0 and 'tests' in report['phases'],report
run('aggregate-final',['node',scripts/'split-aggregate.cjs',out/'snapshots-measured',out/'counts-final-full.json'])
