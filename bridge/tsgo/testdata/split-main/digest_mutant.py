"""A corrupt cached executable must fail before it can be invoked."""
import json, os, subprocess
from pathlib import Path
root=Path(__file__).resolve().parents[4]
out=root/'review/compiler/test-split-bridge-main'
cache=Path('/workspace/bridge-main-products')
inputs=[p for p in cache.glob('*.inputs') if p.read_text().splitlines()[0]=='name bridge-test-api']
p=max(inputs,key=lambda p:p.stat().st_mtime).with_suffix('')/'api'
original=p.read_bytes()
env=dict(os.environ,GOMAXPROCS='4',ADAMIC_UNIT_BUDGET='1',ADAMIC_BUILD_CACHE_DIR=str(cache))
env.pop('ADAMIC_TEST_SHARD',None);env.pop('ADAMIC_TSGO_CORPUS',None)
command=['go','test','./bridge/tsgo','-run','^TestBridgeABI$','-count=1','-json','-timeout=90s']
try:
 p.write_bytes(original+b'planted corruption')
 with (out/'cached-digest-mutant.jsonl').open('w') as log: result=subprocess.run(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 assert result.returncode!=0
 assert 'product digest: api' in (out/'cached-digest-mutant.jsonl').read_text()
finally: p.write_bytes(original)
with (out/'restored-digest.jsonl').open('w') as log: restored=subprocess.run(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
assert restored.returncode==0
(out/'cached-digest-mutant.json').write_text(json.dumps(dict(command=command,mutant_exit=result.returncode,restored_exit=restored.returncode,reason='product digest: api'),indent=2)+'\n')
print('corrupt requested product refused; restored product passes')
