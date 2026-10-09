import json,subprocess,time
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-tsprinter-shards')
for item in json.loads((p/'mutation-plan.json').read_text()):
 if item['kind']!='production':continue
 id=item['id'];diff=str(p/'diffs'/f'{id}.diff');family='Statements' if id=='M2' else 'Expressions'
 subprocess.run(['git','apply','--check',diff],check=True)
 subprocess.run(['git','apply',diff],check=True)
 command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/tsprinter/','-run',f'^TestProduct_TSPrinter{family}(Sanitized|Release)$']
 start=time.monotonic()
 try:
  with (p/f'build-{id}.log').open('w') as out:r=subprocess.run(command,stdout=out,stderr=subprocess.STDOUT)
  (p/f'build-{id}.json').write_text(json.dumps({'id':id,'command':command,'exit':r.returncode,'seconds_wall':time.monotonic()-start})+'\n')
 finally:subprocess.run(['git','apply','-R',diff],check=True)
