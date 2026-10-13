import json, subprocess, time
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-slice-2-2';directory=out/'weak-isolated';directory.mkdir(exist_ok=True)
weak='internal/lower/weak_representations.go';obj='internal/lower/object.go'
source=(root/weak).read_text();before='return l.notYet(node, callee.Name().Text()+" callback over narrowed Weak elements whose parameter needs handle-to-target conversion; use a loop with an explicit narrowed copy")';assert source.count(before)==1
files={weak:source.replace(before,'return nil'),obj:subprocess.check_output(['git','show','origin/main:'+obj],text=True)}
mapping={}
for i,(name,text) in enumerate(files.items()):
 target=directory/(str(i)+'.go.txt');target.write_text(text);mapping[str(root/name)]=str(target)
overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':mapping}))
command=['go','test','./internal/oracle','-overlay='+str(overlay),'-run','^TestWeakDifferential(Each|Find|Map|Reduce|Some|Sort)$','-count=1','-v','-timeout=90s']
start=time.monotonic()
with (directory/'test.log').open('w') as log:run=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=300)
text=(directory/'test.log').read_text();markers=['both backends unexpectedly admit the pinned callback NotYet','AddressSanitizer: heap-buffer-overflow','sanitized: exit codes differ']
row=dict(mutant='weak-callback-isolated',exit=run.returncode,caught=run.returncode==1 and all(x in text for x in markers) and '[build failed]' not in text and 'test timed out' not in text,markers=markers,command=command,seconds=round(time.monotonic()-start,3))
(directory/'result.json').write_text(json.dumps(row,indent=2)+'\n');print(json.dumps(row),flush=True);assert row['caught']
