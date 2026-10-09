from pathlib import Path
import json,subprocess,time,shlex
p=Path('review/test-audit/cmd-adamic-stage1-progress'); plan=json.loads((p/'plan.json').read_text());file=Path('cmd/adamic-stage1-progress/main.go'); original=subprocess.check_output(['git','show','b955d3030e79e9f01d12a96ab1fb4b167fc91c0c:'+str(file)],text=True);file.write_text(original);records=[]
try:
 for entry in plan:
  patch=p/'diffs'/(entry['id']+'.diff');subprocess.run(['git','apply','--check',str(patch)],check=True);subprocess.run(['git','apply',str(patch)],check=True);command=['timeout','90','go','vet','./cmd/adamic-stage1-progress/'];started=time.monotonic()
  with (p/(entry['id']+'-vet.log')).open('w') as out:result=subprocess.run(command,stdout=out,stderr=subprocess.STDOUT)
  record=dict(id=entry['id'],applies=True,command=shlex.join(command),exit=result.returncode,wall_seconds=time.monotonic()-started);records.append(record);(p/'validation.json').write_text(json.dumps(records,indent=2)+'\n');print(entry['id'],result.returncode,round(record['wall_seconds'],3),flush=True);file.write_text(original)
  if result.returncode:raise RuntimeError('invalid standalone diff')
finally:file.write_text(original)
