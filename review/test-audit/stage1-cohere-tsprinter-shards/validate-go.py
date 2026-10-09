import json,subprocess,time
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-tsprinter-shards')
for item in json.loads((p/'mutation-plan.json').read_text()):
 if item['kind']=='production':continue
 diff=str(p/'diffs'/f"{item['id']}.diff")
 subprocess.run(['git','apply','--check',diff],check=True)
 subprocess.run(['git','apply',diff],check=True)
 start=time.monotonic()
 try:
  with (p/f"vet-{item['id']}.log").open('w') as out:r=subprocess.run(['timeout','90','go','vet','./stage1/cohere/tsprinter/'],stdout=out,stderr=subprocess.STDOUT)
  (p/f"vet-{item['id']}.json").write_text(json.dumps({'exit':r.returncode,'seconds':time.monotonic()-start})+'\n')
 finally:subprocess.run(['git','apply','-R',diff],check=True)
