from pathlib import Path
import os,subprocess,json,time
s=Path('/tmp/watch-file-check-cead');repo=s/'work';sha=(s/'implementation-sha.txt').read_text().strip()
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip()==sha
assert not subprocess.check_output(['git','status','--porcelain'],cwd=repo,text=True).strip()
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/tmp/watch-functions-1532/TypeScript',STAGE3_CACHE=str(s/'cache'),ADAMIC_BUILD_CACHE_DIR=str(s/'cache'),TMPDIR=str(s/'tmp'),GOTMPDIR=str(s/'tmp'),GOCACHE='/tmp/watch-file-check-1532/go-cache',NODE_PATH=str(repo/'stage3/api/node_modules'),npm_config_cache='/tmp/watch-functions-1532/npm-cache',PYTHONDONTWRITEBYTECODE='1',NODE_OPTIONS='--max-old-space-size=1536')
(s/'memory-events-before-bounded-lane.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
cmd=[str(repo/'stage3/lane/run.sh'),str(s/'lane-bounded')];start=time.monotonic()
with (s/'whole-lane-bounded.log').open('w') as log:r=subprocess.run(cmd,cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
(s/'memory-events-after-bounded-lane.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
(s/'bounded-lane-command.json').write_text(json.dumps({'command':cmd,'exit':r.returncode,'seconds':time.monotonic()-start,'implementation':sha,'environment':{k:env[k] for k in ['ADAMIC_TYPESCRIPT_SOURCE','STAGE3_CACHE','TMPDIR','GOTMPDIR','GOCACHE','NODE_PATH','npm_config_cache','NODE_OPTIONS']}},indent=2)+'\n');assert r.returncode==0
assert json.loads((s/'lane-bounded/report.json').read_text())['status']=='pass'
remote=subprocess.check_output(['git','ls-remote','--heads','origin','refs/heads/typescript/watch-file-check-1532-1'],cwd=repo,text=True).strip();assert not remote,'implementation branch name already exists remotely'
with (s/'push-implementation.log').open('w') as log:r=subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/watch-file-check-1532-1'],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
assert r.returncode==0
(s/'ALL_PROOFS_OK').write_text(sha+'\n')
