from pathlib import Path
import subprocess,os
root=Path(__file__).resolve().parents[3]
cases=[('native-unknown-producer','internal/native/view_callables_boxing.go','body.WriteString("(void)self; (void)arguments; (void)count; (void)discard; static const char message[] = \\"callable ABI adapter: unknown producer signature\\"; adamic_panic(message, sizeof message - 1);\\n}\\n")','body.WriteString("return adamic_node_performance_invoke(self, arguments, count, discard);\\n}\\n")','./internal/native','^TestViewCallableBoxingUnknownProducer$','stdout "7\\n"'),('javascript-unknown-producer','internal/javascript/view_callables_boxing.go','panic(\\"callable ABI adapter: unknown producer signature\\");','return adamicCall(callee, arguments_);','./internal/javascript','^TestViewCallableBoxingUnknownProducer$','7'),('scalar-never-exemption','internal/lower/view_lazy.go','!(contract != 0 && program.ViewContracts[contract-1].Kind == ir.ViewCallable && viewCallableConcreteReadContract(program, contract))','!viewCallableConcreteReadContract(program, contract)','./internal/oracle','^TestCheckedViewCallableRetainsNeverWiderHelper$','wider helper lost original never refusal: <nil>')]
for name,file,before,after,package,test,needle in cases:
 before=before.replace(chr(92)*2,chr(92));after=after.replace(chr(92)*2,chr(92));needle=needle.replace(chr(92)*2,chr(92))
 p=root/file;old=p.read_text();assert old.count(before)==1,(name,old.count(before))
 try:
  p.write_text(old.replace(before,after))
  log=Path('/tmp/views-integration-callable-5d0-'+name+'-mutant.log')
  with log.open('w') as f:r=subprocess.run(['go','test',package,'-run',test,'-v','-count=1','-timeout','5m'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
  out=log.read_text();assert r.returncode!=0 and '--- FAIL:' in out and '[build failed]' not in out and 'clang failed:' not in out and needle in out,out
  if name=='scalar-never-exemption':
   env=os.environ.copy();env['ADAMIC_CALLABLE_NEVER_MUTANT']='1'
   with Path('/tmp/views-integration-callable-5d0-never-mutant-executes.log').open('w') as f:r=subprocess.run(['go','test',package,'-run',test,'-v','-count=1','-timeout','5m'],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
   assert r.returncode==0
  print(name+': executable mutant caught',flush=True)
 finally:p.write_text(old)
