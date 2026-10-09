"""Measure one top-level bridge test per invocation, including lazy setup."""
import json, os, re, shutil, subprocess, sys, time
from pathlib import Path
root = Path(__file__).resolve().parents[4]
evidence = root/'review/compiler/test-split-bridge-main'; evidence.mkdir(parents=True,exist_ok=True)
logs = Path('/workspace/bridge-main-leaves'); logs.mkdir(exist_ok=True)
text = (root/'bridge/tsgo/registered_units_test.go').read_text()
units = re.findall(r'\{"(TestBridge\w+)", "([^"]+)", "([^"]*)",', text)
assert len(units) == 36
jobs = [(name, 'corpus' if file and file != 'sample.ts' else 'sample') for name, _, file in units]
jobs += [('TestBridgeUnitsCoverEveryPiece','sample'), ('TestBridgeUnitsCoverEveryPiece','corpus'), ('TestBridgeProductCacheIsVerified','sample'), ('TestTSGoRequiresLink','sample')]
base = dict(os.environ, GOMAXPROCS='4', ADAMIC_UNIT_BUDGET='1', ADAMIC_BUILD_CACHE_DIR='/workspace/bridge-main-products', ADAMIC_BUILD_LOG='/workspace/bridge-main-builds.log')
base.pop('ADAMIC_TEST_SHARD', None)
cold = '--cold' in sys.argv
results = []
for name, mode in jobs:
 env = base.copy()
 if mode == 'corpus': env['ADAMIC_TSGO_CORPUS'] = '/workspace/test-split-typescript'
 else: env.pop('ADAMIC_TSGO_CORPUS',None)
 if cold:
  cache = Path('/workspace/bridge-main-cold')/(name+'-'+mode)
  if cache.exists(): shutil.rmtree(cache)
  env['ADAMIC_BUILD_CACHE_DIR'] = str(cache)
 command = ['go','test','./bridge/tsgo','-run','^'+name+'$','-count=1','-json','-timeout=90s']
 log = logs/(name+'-'+mode+('-cold' if cold else '')+'.jsonl')
 start = time.monotonic()
 with log.open('w') as output: run = subprocess.run(command,cwd=root,env=env,stdout=output,stderr=subprocess.STDOUT)
 seconds = time.monotonic()-start
 events = []
 for line in log.read_text().splitlines():
  try: events.append(json.loads(line))
  except ValueError: pass
 terminal = [e for e in events if e.get('Test') == name and e.get('Action') in ['pass','fail','skip']]
 result = dict(test=name, mode=mode, invocation_seconds=round(seconds,3), test_seconds=terminal[-1].get('Elapsed') if terminal else None, action=terminal[-1]['Action'] if terminal else 'build-fail', exit=run.returncode, command=command, log=str(log))
 results.append(result)
 (evidence/('cold-leaves.json' if cold else 'leaves.json')).write_text(json.dumps(results,indent=2)+'\n')
 if cold and cache.exists(): shutil.rmtree(cache)
 print(name,mode,result['action'],seconds,flush=True)
 if run.returncode or result['action']!='pass' or seconds>=60:
  print(log.read_text()[-6000:],flush=True)
  raise SystemExit('leaf failed or exceeded 60 seconds')
