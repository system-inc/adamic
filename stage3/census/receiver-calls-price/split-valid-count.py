"""Accepted counting attempt: original guards and exact arrow local types."""
import pathlib,subprocess,sys,os,json,hashlib
repo=pathlib.Path('/workspace/scratch/receiver-calls-price');scratch=repo.parent;out=scratch/'receiver-split-results';tree=scratch/'receiver-split-counting-measured';scripts=repo/'stage3/census/receiver-calls-price'
os.environ.update(GOTMPDIR=str(scratch/'receiver-tmp'),TMPDIR=str(scratch/'receiver-tmp'),GOCACHE=str(scratch/'receiver-gocache'),npm_config_cache=str(scratch/'receiver-npm-cache'))
def run(name,cmd,cwd=repo,env=None,allow=()):
 code=subprocess.run([sys.executable,str(scratch/'receiver-measure.py'),'--out',str(out/name),'--cwd',str(cwd),'--',*map(str,cmd)],env=env).returncode;print(json.dumps({'phase':name,'exit':code}),flush=True)
 if code and code not in allow:raise SystemExit(code)
run('controls-valid',['node',scripts/'split-fixture-controls.cjs'])
run('counting-stock-valid',['node',repo/'stage3/api/node_modules/typescript/lib/tsc.js','--project','src/compiler/tsconfig.json','--noEmit','--strictFunctionTypes','true','--incremental','false','--composite','false','--pretty','false'],tree)
files=['split-instrument.cjs','split-hooks.cjs','split-aggregate.cjs','split-fixture-controls.cjs']
(out/'counting-script-hashes.json').write_text(json.dumps({f:hashlib.sha256((scripts/f).read_bytes()).hexdigest() for f in files},indent=2)+'\n')
env=dict(os.environ,NODE_OPTIONS='--require='+str(scripts/'split-hooks.cjs'),RECEIVER_PRICE_DIR=str(out/'snapshots-valid'))
run('counting-lane-valid',[sys.executable,repo/'stage3/oracle/run.py',tree,out/'oracle-valid'],env=env,allow=(1,))
report=json.loads((out/'oracle-valid/report.json').read_text());assert report['phases']['install']['exit']==0 and report['phases']['build']['exit']==0 and 'tests' in report['phases'],report
run('aggregate-valid',['node',scripts/'split-aggregate.cjs',out/'snapshots-valid',out/'counts-valid-full.json'])
