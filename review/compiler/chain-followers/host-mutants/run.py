import json,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
cases=json.loads((out/'cases.json').read_text())
results=[]
for case in cases:
 command=['go','test','-overlay',case['overlay'],*case['packages'],'-run',case['pattern'],'-count=1','-timeout','90s','-v']
 log=out/(case['topic']+'.log')
 with log.open('w') as stream:
  completed=subprocess.run(command,stdout=stream,stderr=subprocess.STDOUT,timeout=150)
 content=log.read_text()
 caught=completed.returncode==1 and '--- FAIL: Test' in content and '[build failed]' not in content and 'no tests to run' not in content
 results.append(dict(topic=case['topic'],command=command,exit=completed.returncode,caught=caught))
 (out/'results.json').write_text(json.dumps(results,indent=2))
 print(case['topic'], 'caught' if caught else 'NOT CAUGHT', flush=True)
 if not caught:raise SystemExit(1)
