import pathlib,subprocess,tarfile,json,difflib
root=pathlib.Path.cwd();out=root/'review/compiler/chain-followers/host-mutants'; archive=tarfile.open('/tmp/host-source-evidence.tar.gz')
cases=json.loads((out/'cases.json').read_text())
for c in cases:
 target=c['file'];source='internal/lower/phantom_brands.go' if c['topic']=='phantom-primitive' else target
 base=subprocess.check_output(['git','show','6c892bee:'+source],text=True)
 mutant=archive.extractfile('mutants/'+c['topic']+'.go').read().decode()
 current=(root/target).read_text()
 for tag,a,b,x,y in difflib.SequenceMatcher(None,base.splitlines(True),mutant.splitlines(True),autojunk=False).get_opcodes():
  if tag=='equal':continue
  before=''.join(base.splitlines(True)[a:b]);after=''.join(mutant.splitlines(True)[x:y])
  if not before:
   anchor=''.join(base.splitlines(True)[max(0,a-1):a]);assert current.count(anchor)==1,(c['topic'],anchor);current=current.replace(anchor,anchor+after,1)
  else:
   assert current.count(before)==1,(c['topic'],before);current=current.replace(before,after,1)
 path=out/(c['topic']+'.go.txt');path.write_text(current)
 (out/(c['topic']+'.json')).write_text(json.dumps({'Replace':{str(root/target):str(path)}}))
