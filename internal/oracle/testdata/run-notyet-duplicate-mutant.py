#!/usr/bin/env python3
from pathlib import Path
import json,os,subprocess
root=Path(__file__).resolve().parents[3]
logs=Path('/tmp/destructuring-duplicate-mutant');logs.mkdir(exist_ok=True)
file=root/'internal/lower/object.go';before='if prior.Name == fieldName {'
original=file.read_text();assert original.count(before)==1
source=logs/'object.go';source.write_text(original.replace(before,'if prior.Name == fieldName && false {',1))
overlay=logs/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(file):str(source)}}))
with (logs/'mutant.log').open('w') as log:
 result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lower','-run','^TestComputedDuplicateFieldStaysExplicit$','-count=1'],cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
output=(logs/'mutant.log').read_text()
assert result.returncode!=0 and 'want the repeated-field gap' in output and '[build failed]' not in output,(result.returncode,output)
print('duplicate-field: caught by want the repeated-field gap')
