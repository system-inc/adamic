import subprocess,pathlib,time,json,os,re
root=pathlib.Path('/workspace/adamic'); ev=root/'review/test-audit/stage1-cohere-lint-registry'; rows=['TestDeterministicRegeneration','TestDescriptorRejections','TestDuplicateOracleAdapter','TestAdamicRuleModule']; records=[]
for name in rows:
 for i in range(3):
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/registry/','-run','^'+name+'$']; start=time.monotonic()
  with (ev/(name+'-'+str(i+1)+'.log')).open('w') as log: result=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT)
  records.append(dict(id=name+'-'+str(i+1),command=' '.join(command),exit=result.returncode,wall=time.monotonic()-start))
for item in json.loads((ev/'menu.json').read_text()):
 mid=item['id']; command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/registry/','-run','.']; env=dict(os.environ,ADAMIC_MUTANT=mid,ADAMIC_BUILD_CACHE_DIR='/tmp/u119/cache/'+mid); start=time.monotonic()
 with (ev/(mid+'.log')).open('w') as log: result=subprocess.run(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 records.append(dict(id=mid,command='ADAMIC_MUTANT='+mid+' ADAMIC_BUILD_CACHE_DIR=/tmp/u119/cache/'+mid+' '+' '.join(command),exit=result.returncode,wall=time.monotonic()-start))
(ev/'commands.json').write_text(json.dumps(records,indent=2))
