from pathlib import Path
import json,subprocess,time
out=Path(__file__).resolve().parent;root=Path.cwd();rows=json.loads((root/'review/compiler/loom-slice-4/scout-results.json').read_text());results=[]
for i,row in enumerate(rows):
 cmd=row['command'];log=out/('scout-%d.log'%i)
 start=time.monotonic()
 with log.open('w') as capture:r=subprocess.run(cmd,stdout=capture,stderr=subprocess.STDOUT,timeout=90)
 assert r.returncode==row['expected'],(cmd,r.returncode,log.read_text())
 results.append(dict(command=cmd,exit=r.returncode,expected=row['expected'],seconds=time.monotonic()-start));print(i,r.returncode,flush=True)
(out/'scout-results.json').write_text(json.dumps(results,indent=2)+'\n')
