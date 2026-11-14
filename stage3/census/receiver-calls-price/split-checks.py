"""Sequential declarations, real counter controls, counting lane and full censuses."""
import os,sys,json,pathlib,subprocess,shutil,hashlib
repo=pathlib.Path('/workspace/scratch/receiver-calls-price');scratch=pathlib.Path('/workspace/scratch');out=scratch/'receiver-split-results';base=scratch/'receiver-baseline';split=scratch/'receiver-split';counting=scratch/'receiver-split-counting';scripts=repo/'stage3/census/receiver-calls-price'
os.environ.update(GOTMPDIR=str(scratch/'receiver-tmp'),TMPDIR=str(scratch/'receiver-tmp'),GOCACHE=str(scratch/'receiver-gocache'),npm_config_cache=str(scratch/'receiver-npm-cache'))
def run(name,cmd,cwd=repo,env=None,allow=()):
 dest=out/name
 code=subprocess.run([sys.executable,str(scratch/'receiver-measure.py'),'--out',str(dest),'--cwd',str(cwd),'--',*map(str,cmd)],env=env).returncode
 print(json.dumps({'phase':name,'exit':code}),flush=True)
 if code and code not in allow:raise SystemExit(code)
 return code
run('controls',['node',scripts/'split-fixture-controls.cjs'])
control_dir=out/'control-snapshots';control_dir.mkdir(exist_ok=True)
shutil.copyfile(scratch/'split-fixture-controls.json',control_dir/'control.json')
run('this-controls',['node',scripts/'split-aggregate.cjs',control_dir,out/'this-controls-result.json'])
this_control=json.loads((out/'this-controls-result.json').read_text())['implementationThis']
assert this_control['reads']==1 and this_control['parseFailures']==0,this_control
# A link shares only locked dependencies, not generated declaration outputs.
if not (base/'node_modules').exists():(base/'node_modules').symlink_to(split/'node_modules',target_is_directory=True)
run('dts-baseline',['node_modules/.bin/hereby','dts'],base)
run('dts-split',['node_modules/.bin/hereby','dts'],split)
declarations=[]
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
 a=(base/'built/local'/name).read_bytes();b=(split/'built/local'/name).read_bytes();declarations.append({'file':name,'identical':a==b,'baselineBytes':len(a),'prototypeBytes':len(b),'sha256':hashlib.sha256(b).hexdigest()})
(out/'declarations.json').write_text(json.dumps(declarations,indent=2)+'\n');assert all(d['identical'] for d in declarations)
if not counting.exists():shutil.copytree(split,counting,ignore=lambda directory,names: [name for name in names if pathlib.Path(directory)==split and name in ('node_modules','built','.parallelperf.json')])
run('instrument',['node',scripts/'split-instrument.cjs',counting,scratch/'receiver-heat.json',out/'instrument.json'])
# Ensure the exact test input/reference corpus survives copying and instrumentation.
inventory={}
for folder in ['tests/cases','tests/baselines/reference']:
 original={str(p.relative_to(split)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (split/folder).rglob('*') if p.is_file()}
 actual={str(p.relative_to(counting)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (counting/folder).rglob('*') if p.is_file()}
 assert original==actual,folder
 inventory[folder]={'files':len(original),'identical':True,'inventorySha256':hashlib.sha256(json.dumps(original,sort_keys=True).encode()).hexdigest()}
(out/'test-inputs.json').write_text(json.dumps(inventory,indent=2)+'\n')
# Controls use separate processes/directories; only the unfiltered oracle gets these snapshots.
env=dict(os.environ,NODE_OPTIONS='--require='+str(scripts/'split-hooks.cjs'),RECEIVER_PRICE_DIR=str(out/'snapshots'))
run('counting-lane',[sys.executable,repo/'stage3/oracle/run.py',counting,out/'oracle'],env=env,allow=(1,))
run('aggregate',['node',scripts/'split-aggregate.cjs',out/'snapshots',out/'counts-full.json'])
run('census-baseline',[sys.executable,scripts/'split-census.py','baseline',base,out/'census'])
run('census-split',[sys.executable,scripts/'split-census.py','split',split,out/'census'])
