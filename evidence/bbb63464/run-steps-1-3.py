import json,pathlib,shlex,subprocess
root=pathlib.Path('/workspace/adamic/evidence/bbb63464')
work=pathlib.Path('/workspace/adamic-rule-g-2594029')
pin='bbb63464cd039548dacaaa7ccbde06915c291b8a'
base='e3aa0bb7642163ebd004a71ba40a4657d27eab71'
old='e5531c8c7fd9fa0b5076bb3d0d0961e6954e9ce1'
results={}
def run(label,args):
 args=list(map(str,args));(root/(label+'.command')).write_text(shlex.join(args)+'\n')
 r=subprocess.run(args,cwd=work,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
 (root/(label+'.stdout')).write_bytes(r.stdout);(root/(label+'.stderr')).write_bytes(r.stderr)
 (root/(label+'.exit')).write_text(str(r.returncode)+'\n')
 results[label]={'exit':r.returncode,'stdout':r.stdout.decode(errors='replace'),'stderr':r.stderr.decode(errors='replace')}
 (root/'steps-1-3.json').write_text(json.dumps(results,indent=2)+'\n')
 return r.returncode
def checkout(label,sha):
 if run(label+'-checkout',['git','switch','--detach',sha]):raise RuntimeError('checkout failed')
 run(label+'-revision',['git','rev-parse','HEAD'])
 run(label+'-submodules',['git','submodule','status','--recursive'])
 if run(label+'-clean',['git','diff','--exit-code',sha,'--']):raise RuntimeError('dirty source')
checkout('01-bbb63464',pin)
run('02-go-version',['go','version'])
run('03-clang-version',['clang','--version'])
run('04-node-version',['node','--version'])
testpath=pathlib.Path('internal/lower/rule_g_test.go')
test=(work/testpath).read_bytes();(root/'rule_g_test.go.txt').write_bytes(test)
run('10-gofmt',['gofmt','-l','internal/lower/'])
run('11-vet',['go','vet','./internal/lower/'])
run('12-focused',['go','test','-count=1','-run','TestAGenericTableCast|TestASetTimeout|TestAGenericMethod|TestAGenericFunction|TestAHostMethod|TestNodeWatch','-v','./internal/lower/'])
for label,sha in [('20-main',base),('21-e5531c8',old)]:
 checkout(label,sha)
 exists=subprocess.run(['git','cat-file','-e',sha+':'+str(testpath)],cwd=work,stdout=subprocess.PIPE,stderr=subprocess.PIPE).returncode==0
 (work/testpath).write_bytes(test)
 (root/(label+'-copied-test.txt')).write_bytes(test)
 try:
  run(label+'-new-test',['go','test','-count=1','-run','^TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal$','-v','./internal/lower/'])
 finally:
  if exists:
   run(label+'-restore-test',['git','restore','--source='+sha,'--',testpath])
  else:
   (work/testpath).unlink()
  run(label+'-restoration',['git','diff','--exit-code',sha,'--'])
checkout('30-bbb63464',pin)
path=work/'internal/lower/invariance.go';original=path.read_bytes()
s=original.decode();a='if len(toSignatures[0].TypeParameters()) == 0 || l.signatureAny(instantiated, target) == nil {'
assert s.count(a)==1
path.write_text(s.replace(a,'if true {'))
try:
 run('31-mutant-patch',['git','diff','--','internal/lower/invariance.go'])
 run('32-mutant',['go','test','-count=1','-run','TestAGenericTableCastThroughAWrittenAnyKeepsMainsRefusal|TestASetTimeoutProviderIsRelatedThroughTheNoInferSlot','-v','./internal/lower/'])
finally:
 path.write_bytes(original)
run('33-final-restoration',['git','diff','--exit-code',pin,'--'])
run('34-final-status',['git','status','--short'])
print(json.dumps({k:v['exit'] for k,v in results.items()},indent=2))
