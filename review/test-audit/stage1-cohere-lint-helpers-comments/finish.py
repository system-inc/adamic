import pathlib,json,subprocess,re,time,statistics,shutil
root=pathlib.Path('/workspace/adamic');pkg='stage1/cohere/lint/helpers/comments';p=root/pkg;out=root/'review/test-audit/stage1-cohere-lint-helpers-comments'
def events(file):
 a=[]
 for line in file.read_text().splitlines():
  try:a.append(json.loads(line))
  except:pass
 return a
for name in ['base.txt','discovery.log','npm-api.log','baseline.log']:
 shutil.copy('/tmp/u115/'+name,out/name)
for f in pathlib.Path('/tmp/u115').glob('clean-*.log'):shutil.copy(f,out/f.name)
base={str(f.relative_to(root)):f.read_text() for f in list(p.glob('*.ts'))+list(p.glob('*_test.go'))}
checks=[]
try:
 # Remove unreachable code from the menu's empty-answer probe.
 f=p/'comment_mutants_test.go';s=f.read_text();start=s.index('func commentMutants()');end=s.index('\n}\n',start)+2;f.write_text(s[:start]+'func commentMutants() []struct{ file, old, new string } { return nil }'+s[end:]);(out/'P2.diff').write_bytes(subprocess.check_output(['git','diff','--',pkg],cwd=root));f.write_text(s)
 for id in ['M1','M2','M3','M4','P1','W1','W2','W3','S1','P2']:
  patch=out/(id+'.diff');subprocess.run(['git','apply','--check',str(patch)],cwd=root,check=True);subprocess.run(['git','apply',str(patch)],cwd=root,check=True)
  cmd=['timeout','90','go','run','./cmd/adamic','build',str(p/'main.ts'),'-o','/tmp/u115/'+id+'-native','--sanitize'] if id.startswith('M') or id=='P1' else ['go','vet','./'+pkg+'/']
  started=time.monotonic()
  with (out/(id+'-build.log')).open('w') as fd:r=subprocess.run(cmd,cwd=root,stdout=fd,stderr=subprocess.STDOUT)
  if id=='P2':
   with (out/'P2.log').open('w') as fd: subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./'+pkg+'/','-run','^TestCommentMutantsUnion$'],cwd=root,stdout=fd,stderr=subprocess.STDOUT)
  checks.append({'id':id,'command':' '.join(cmd),'exit':r.returncode,'seconds':time.monotonic()-started});print('validation',checks[-1],flush=True)
  for path,orig in base.items():(root/path).write_text(orig)
 (out/'validation.json').write_text(json.dumps(checks,indent=2))
finally:
 for path,orig in base.items():(root/path).write_text(orig)
