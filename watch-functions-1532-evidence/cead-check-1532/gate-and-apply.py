from pathlib import Path
import os,subprocess,json,time
s=Path('/tmp/watch-file-check-cead');repo=s/'work'
while not (s/'census-done').exists():time.sleep(1)
with (s/'delta-summary.log').open('w') as log:subprocess.run(['python3',str(s/'compare-census.py')],stdout=log,stderr=subprocess.STDOUT,check=True)
d=json.loads((s/'delta-after.json').read_text());assert not d['new'] and not d['changed'], 'mandatory new findings gate failed';assert len(d['gone'])==1 and d['target1532Gone'] and d['sibling1533Unchanged']
(s/'CENSUS_OK').write_text('full census: 0 new, 1 gone, 0 changed; sibling unchanged\n')
with (s/'amend.log').open('w') as log:subprocess.run(['node',str(s/'amend.cjs')],stdout=log,stderr=subprocess.STDOUT,check=True)
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/tmp/watch-functions-1532/TypeScript',STAGE3_CACHE=str(s/'cache'),ADAMIC_BUILD_CACHE_DIR=str(s/'cache'),TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'tmp'),GOCACHE='/tmp/watch-file-check-1532/go-cache',NODE_PATH=str(repo/'stage3/api/node_modules'),npm_config_cache='/tmp/watch-functions-1532/npm-cache',PYTHONDONTWRITEBYTECODE='1')
with (s/'apply-after.log').open('w') as log:r=subprocess.run([str(repo/'stage3/apply.sh'),str(s/'after-tree')],env=env,stdout=log,stderr=subprocess.STDOUT)
(s/'apply-after-exit.txt').write_text(str(r.returncode));assert r.returncode==0
(s/'APPLY_OK').write_text('complete\n')
