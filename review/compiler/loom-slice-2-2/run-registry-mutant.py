import json, subprocess, time
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-slice-2-2/registry-mutant';out.mkdir(exist_ok=True)
path=root/'internal/oracle/weak_differential_test.go';source=path.read_text()
before='[]string{"4ddd17f_weak_union_narrowed", "4ddd17f_weak_single_narrowed"}'
assert source.count(before)==1
mutant=out/'source.go.txt';mutant.write_text(source.replace(before,'[]string{"4ddd17f_weak_union_narrowed"}'))
overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(path):str(mutant)}}))
command=['go','test','./internal/oracle','-overlay='+str(overlay),'-run','^TestCountsAreRecordedTable$','-count=1','-v','-timeout=90s']
start=time.monotonic()
with (out/'test.log').open('w') as log:ran=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=150)
text=(out/'test.log').read_text();caught=ran.returncode==1 and "want the row of" in text and '[build failed]' not in text and 'test timed out' not in text
row=dict(mutant='weak-counter-registration-omitted',caught=caught,exit=ran.returncode,command=command,seconds=round(time.monotonic()-start,3),catcher='TestCountsAreRecordedTable')
(out/'result.json').write_text(json.dumps(row,indent=2)+'\n');print(json.dumps(row),flush=True);assert caught
