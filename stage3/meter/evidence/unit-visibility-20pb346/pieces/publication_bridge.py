"""Handshake for non-force publication through the authenticated GitHub app.

The census runner commits locally first, then waits for the task agent to create
the same tree through the connector and fast-forward the requested branch with
force=false and an expected parent SHA. Anonymous fetch reconciles the new
connector-authored commit only after verifying identical tree content.
"""
from pathlib import Path
import json, subprocess, time
root=Path('/workspace/scratch/20pb346-2')
branch='typescript/meter-stream-records-2'
def publish_commit(repo, logfile, stop_event=None):
 def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
 sha=git('rev-parse','HEAD');parent=git('rev-parse','HEAD^')
 folder=root/'connector-requests';folder.mkdir(exist_ok=True)
 request=folder/(sha+'.json');ack=folder/(sha+'.ack.json')
 data=dict(local_sha=sha,parent=parent,tree=git('rev-parse','HEAD^{tree}'),base_tree=git('rev-parse','HEAD^^{tree}'),message=git('show','-s','--format=%B','HEAD'),branch=branch)
 if not request.exists():
  temporary=request.with_suffix('.pending');temporary.write_text(json.dumps(data,indent=2)+'\n');temporary.replace(request)
 while not ack.exists():
  if stop_event is not None and stop_event.is_set():raise RuntimeError('publication interrupted; local commit retained')
  time.sleep(1)
 response=json.loads(ack.read_text());assert response['local_sha']==sha
 ref='refs/remotes/origin/'+branch
 with logfile.open('a') as log:
  subprocess.run(['git','-c','credential.helper=','fetch','origin','refs/heads/'+branch+':'+ref],cwd=repo,check=True,stdout=log,stderr=subprocess.STDOUT)
 assert git('rev-parse',ref)==response['remote_sha']
 assert git('rev-parse',ref+'^{tree}')==data['tree'],'connector content changed'
 subprocess.run(['git','reset','--soft',ref],cwd=repo,check=True,stdout=subprocess.DEVNULL)
 return response['remote_sha']
