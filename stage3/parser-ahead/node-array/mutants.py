#!/usr/bin/env python3
import json,pathlib,subprocess,tempfile
root=pathlib.Path(__file__).resolve().parents[3]
p=root/'internal/lower/node_array_layout.go';source=p.read_text();needle='return slot, nil'
assert source.count(needle)==1
rows=[]
with tempfile.TemporaryDirectory(prefix='step24-array-mutant-') as scratch:
 for field in ['pos','end','hasTrailingComma','transformFlags']:
  changed=pathlib.Path(scratch)/'changed.go.txt';changed.write_text(source.replace(needle,'if field == "'+field+'" { slot.Slot++ }; return slot, nil'))
  overlay=pathlib.Path(scratch)/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(p):str(changed)}}))
  log=root/'review/compiler/step24-parser-main/logs'/('array-'+field+'.log')
  with log.open('w') as output:code=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lower','-run','^TestParserNodeArrayLayout$','-count=1','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT).returncode
  text=log.read_text();assert code!=0 and 'wrong slot for '+field in text and '[build failed]' not in text,(field,text)
  rows.append({'field':field,'exit':code,'catcher':'TestParserNodeArrayLayout slot assertion','native':'planning assertion; source metadata lowering is held to Node'})
(root/'review/compiler/step24-parser-main/node-array-mutants.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(rows,indent=2))
