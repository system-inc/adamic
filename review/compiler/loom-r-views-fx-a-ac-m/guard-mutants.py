from pathlib import Path
import json, subprocess, gzip
root=Path(__file__).resolve().parents[3]
out=Path(__file__).resolve().parent
cases=[('key','element_access_fields.go','\tmodules, err := l.moduleOrder(l.program.Files()[0])','\tfields = map[string]bool{}\n\tmodules, err := l.moduleOrder(l.program.Files()[0])','^TestLoomViewKeyRead$'),('unsupported','interface_cast.go','if contract := l.result.ViewContracts[id-1]; contract.Unsupported != "" {','if contract := l.result.ViewContracts[id-1]; false && contract.Unsupported != "" {','^TestLoomViewUnsupportedRead$')]
results=[]
for name,file,before,after,selector in cases:
 path=root/'internal/lower'/file
 original=path.read_text()
 assert original.count(before)==1,(name,original.count(before))
 changed=original.replace(before,after,1)
 import difflib
 (out/(name+'.patch.gz')).write_bytes(gzip.compress(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/internal/lower/'+file,tofile='b/internal/lower/'+file)).encode(),mtime=0))
 try:
  path.write_text(changed)
  command=['go','test','./internal/lower','-run',selector,'-count=1','-v','-timeout','90s']
  with (out/(name+'-mutant.log')).open('w') as log:
   result=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=110)
  text=(out/(name+'-mutant.log')).read_text()
  caught=result.returncode!=0 and 'want located refusal' in text and 'build failed' not in text
  results.append(dict(mutant=name,command=command,exit=result.returncode,caught=caught))
  assert caught,text
 finally: path.write_text(original)
(out/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
