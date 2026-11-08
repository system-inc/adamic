"""Push a checkpoint ref without adding history to the final report branch."""
import json,subprocess,sys
from pathlib import Path
out=Path(__file__).resolve().parent
repo=out.parents[2]
event=sys.argv[1]
assert event.replace('-','').isalnum()
subprocess.run(['git','add','-f','--',str(out)],cwd=repo,check=True)
tree=subprocess.check_output(['git','write-tree'],cwd=repo,text=True).strip()
sha=subprocess.check_output(['git','commit-tree',tree,'-p','b5c0aa64'],cwd=repo,input='census: checkpoint '+event+'\n',text=True).strip()
ref='refs/heads/typescript/branded-call-census-checkpoint-'+event
with (out/('checkpoint-'+event+'.log')).open('w') as f:
 rc=subprocess.run(['git','push','origin',sha+':'+ref],cwd=repo,stdout=f,stderr=subprocess.STDOUT).returncode
print('push child_exit='+str(rc),sha,ref,flush=True)
p=out/'checkpoints.json';rows=json.loads(p.read_text()) if p.exists() else []
rows.append(dict(event=event,sha=sha,ref=ref,push_exit=rc));p.write_text(json.dumps(rows,indent=2)+'\n')
sys.exit(rc)
