import json,time,subprocess,os
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');begin=time.monotonic()
while not (s/'control.json').exists():
 assert time.monotonic()-begin<1400,'control did not finish within bounded wait'
 print('waiting for full control before isolated count',round(time.monotonic()-begin),flush=True);time.sleep(30)
assert json.loads((s/'control.json').read_text())['exit']==0
with (s/'count-driver.log').open('w') as log:r=subprocess.run(['python3',str(s/'count.py')],cwd='/workspace/adamic',stdout=log,stderr=subprocess.STDOUT)
print('accepted count exit',r.returncode,flush=True);raise SystemExit(r.returncode)
