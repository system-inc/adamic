#!/usr/bin/env python3
from pathlib import Path
import json,os,subprocess
root=Path(__file__).resolve().parents[3]
logs=Path('/tmp/destructuring-storage-mutant');logs.mkdir(exist_ok=True)
file=root/'internal/lower/computed_field_name.go';before='return !strings.ContainsRune(name, 0)'
original=file.read_text();assert original.count(before)==1
source=logs/'computed_field_name.go';source.write_text(original.replace(before,'return true || !strings.ContainsRune(name, 0)',1))
overlay=logs/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(file):str(source)}}))
with (logs/'mutant.log').open('w') as log:
 result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lower','-run','^TestComputedFieldStorageNamesStayExplicit$','-count=1'],cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
output=(logs/'mutant.log').read_text()
assert result.returncode!=0 and 'want the field storage gap' in output and '[build failed]' not in output,(result.returncode,output)
print('storage-field: caught by want the field storage gap')
