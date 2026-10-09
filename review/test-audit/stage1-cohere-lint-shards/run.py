import pathlib,json,subprocess,time,os
root=pathlib.Path('/workspace/adamic'); ev=root/'review/test-audit/stage1-cohere-lint-shards'; path=root/'stage1/cohere/lint/shards/shards.go'; original=(ev/'original.go.txt').read_text(); commands=[]; validation=[]; rows=['TestMergePutsCasesBackInOrder','TestMergeKeepsLinesThatOnlyLookLikeCases','TestMergeRefusesAMissingOrRepeatedCase']; pkg='./stage1/cohere/lint/shards/'
def run(mid,selector):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run',selector]; start=time.monotonic()
 with (ev/(mid+'.log')).open('w')as log:r=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT)
 commands.append(dict(id=mid,command=' '.join(cmd)+' > '+str(ev/(mid+'.log'))+' 2>&1',exit=r.returncode,wall=time.monotonic()-start))
for row in rows:
 for i in range(1,4):run(row+'-'+str(i),'^'+row+'$')
try:
 for m in json.loads((ev/'menu.json').read_text()):
  path.write_text(original); a=subprocess.run(['git','apply','--check',str(ev/(m['id']+'.diff'))],cwd=root,capture_output=True); assert a.returncode==0
  path.write_text(original.replace(m['old'],m['new'])); start=time.monotonic()
  with (ev/(m['id']+'-vet.log')).open('w')as log:v=subprocess.run(['go','vet',pkg],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  validation.append(dict(id=m['id'],apply_exit=a.returncode,vet_exit=v.returncode,seconds=time.monotonic()-start)); assert v.returncode==0,m['id']
  run(m['id'],'.')
finally:path.write_text(original)
run('restored','.')
(ev/'commands.json').write_text(json.dumps(commands,indent=2));(ev/'validation.json').write_text(json.dumps(validation,indent=2))
