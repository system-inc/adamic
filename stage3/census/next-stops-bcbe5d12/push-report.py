import json,subprocess
from pathlib import Path
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch/b5w3ycg')
base='bcbe5d1244ee8a218d860415dea19cca114bae46';branch='typescript/next-stops-bcbe5d12';prefix='stage3/census/next-stops-bcbe5d12/'
def git(args):
 p=subprocess.run(['git',*args],cwd=repo,capture_output=True)
 with (scratch/'final-push-git.log').open('ab') as f:f.write(p.stdout+p.stderr)
 print('git '+args[0]+' child_exit='+str(p.returncode),flush=True);p.check_returncode();return p.stdout.decode().strip()
assert git(['rev-parse','HEAD'])==git(['rev-parse','origin/main'])==base
assert git(['branch','--show-current'])==branch
assert json.loads((out/'validation.json').read_text())['remaining_census_processes']==0
for p in scratch.glob('*-status.json'):(out/p.name).write_bytes(p.read_bytes())
git(['add','stage3/census/next-stops-bcbe5d12'])
paths=git(['diff','--cached','--name-only']).splitlines();assert paths and all(p.startswith(prefix) for p in paths)
assert not any(p.endswith(('.go','.pyc')) for p in paths)
git(['commit','-m','stage3/census: measure next stops on bcbe5d12 across 79 roots'])
sha=git(['rev-parse','HEAD']);assert git(['rev-parse','HEAD^'])==base and git(['rev-list','--count',base+'..HEAD'])=='1'
changed=git(['diff','--name-only',base,'HEAD']).splitlines();assert all(p.startswith(prefix) for p in changed)
git(['push','origin','HEAD:refs/heads/'+branch])
remote=git(['ls-remote','--heads','origin','refs/heads/'+branch]).split()[0];assert remote==sha
receipt=dict(sha=sha,parent=base,branch=branch,verified_remote_sha=remote,changed_files=len(changed),only_requested_directory=True,commits_on_pinned_main=1)
(scratch/'final-push.json').write_text(json.dumps(receipt,indent=2)+'\n');print(json.dumps(receipt),flush=True)
