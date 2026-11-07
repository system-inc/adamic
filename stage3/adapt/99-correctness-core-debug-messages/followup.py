from pathlib import Path
import sys,os,time,json,shutil,difflib
sys.dont_write_bytecode=True;sys.path.insert(0,'/workspace/scratch/debug-messages')
from work import run,s,repo
unit=repo/'stage3/adapt/99-correctness-core-debug-messages';env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
start=time.monotonic();last=start
while not (s/'raw/after-apply.json').exists():
 if time.monotonic()-start>1490:raise SystemExit('candidate apply wait exceeded bound')
 if time.monotonic()-last>30:print('waiting for accepted full apply',flush=True);last=time.monotonic()
 time.sleep(2)
assert json.loads((s/'raw/after-apply.json').read_text())['exit']==0
run('after-build-wrapper',['python3',s/'work.py','build','after'],env=env)
run('normal-build-comparison',['python3',unit/'compare.py','builds'],env=env)
run('final-witness-run',['node',unit/'witness.cjs',s/'before',s/'after',s/'raw/final-witness.json'],env=env)
run('after-types-wrapper',['python3',s/'work.py','types','after'],env=env)
run('after-census-wrapper',['python3',unit/'census.py','after'],env=env)
# Restore exactly one callback expression in a separate private source tree.
shutil.copytree(s/'after/src',s/'mutant/src');file=s/'mutant/src/compiler/debug.ts';text=file.read_bytes().decode();pins=json.loads((unit/'pins.json').read_text())['pins'];pin=pins[0];assert text.count(pin['after'])==1;mutated=text.replace(pin['after'],pin['before']);assert mutated!=text;file.write_bytes(mutated.encode());(s/'raw/restoring-mutant.patch').write_text(''.join(difflib.unified_diff(text.splitlines(True),mutated.splitlines(True),fromfile='after/debug.ts',tofile='mutant/debug.ts')))
run('mutant-census-wrapper',['python3',unit/'census.py','mutant'],env=env)
run('census-comparison-run',['python3',unit/'compare.py','census'],env=env)
run('after-instrument',['node',unit/'instrument-built.cjs',s/'after',s/'measure-after',s/'raw/after-instrumentation.json'],env=env)
control=dict(env,NODE_OPTIONS='--require='+str(unit/'preload.cjs'),DEBUG_MESSAGES_COUNTS=str(s/'raw/after-counter-control'))
run('after-counter-control-run',['node',unit/'counter-control.cjs',s/'measure-after/built/local/typescript.js'],env=control)
processes=[json.loads(p.read_text()) for p in (s/'raw/after-counter-control').glob('*.json')];assert len(processes)==1;assert all(v==1 for v in processes[0]['counts'].values());(s/'raw/after-counter-control.json').write_text(json.dumps(processes,indent=2)+'\n')
print('FOLLOWUP COMPLETE',flush=True)
