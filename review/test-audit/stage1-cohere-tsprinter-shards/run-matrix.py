import json,os,subprocess,time
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-tsprinter-shards')
selector=Path('/workspace/u141-tmp/selector')
rows=json.loads((p/'rows-plan.json').read_text())
full=(p/'regex.txt').read_text()
fast='^('+'|'.join(name for name,regex,kind in rows if kind!='production' and name!='TestStatementsAgainstGoAndPrettier_Setup')+')$'
for id in ['neutral','M1','M2','M3','M4','P1','P2','S1','S2','S3','S4','S5','W1','W2','W3','W4','W5']:
 selector.write_text(id if id.startswith(('M','P')) else '')
 env=os.environ.copy();env['ADAMIC_MUTANT']=id if id.startswith(('S','W')) else ''
 regex=fast if id.startswith('W') else full
 command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/tsprinter/','-run',regex]
 start=time.monotonic()
 with (p/f'{id}.log').open('w') as out:result=subprocess.run(command,env=env,stdout=out,stderr=subprocess.STDOUT)
 (p/f'{id}.json').write_text(json.dumps({'id':id,'selector':selector.read_text(),'ADAMIC_MUTANT':env['ADAMIC_MUTANT'],'command':command,'exit':result.returncode,'seconds_wall':time.monotonic()-start})+'\n')
 if id=='neutral' and result.returncode:break
selector.write_text('')
