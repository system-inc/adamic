import datetime,json,os,pathlib,shlex,subprocess,time
root=pathlib.Path('/workspace/adamic/evidence/d70b3516')
work=pathlib.Path('/workspace/adamic-rule-g-2594029')
pin='d70b35160df85ebc2bbc0bf195b9bf62ffe1ced9'
results={}
def run(label,args,env=None,cwd=work):
 args=list(map(str,args));(root/(label+'.command')).write_text(shlex.join(args)+'\n')
 start=datetime.datetime.now(datetime.timezone.utc).isoformat();clock=time.monotonic()
 with (root/(label+'.stdout')).open('wb') as out,(root/(label+'.stderr')).open('wb') as err:
  r=subprocess.run(args,cwd=cwd,env=env,stdout=out,stderr=err)
 (root/(label+'.exit')).write_text(str(r.returncode)+'\n')
 results[label]={'exit':r.returncode,'started_utc':start,'seconds':time.monotonic()-clock,'stdout':(root/(label+'.stdout')).read_text(errors='replace'),'stderr':(root/(label+'.stderr')).read_text(errors='replace')}
 (root/'results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(label+': exit '+str(r.returncode),flush=True)
 return r.returncode
if run('01-checkout',['git','switch','--detach',pin]):raise RuntimeError('checkout failed')
run('02-revision',['git','rev-parse','HEAD'])
run('03-submodules',['git','submodule','status','--recursive'])
if run('04-clean',['git','diff','--exit-code',pin,'--']):raise RuntimeError('dirty source')
run('05-go-version',['go','version'])
run('06-clang-version',['clang','--version'])
run('07-node-version',['node','--version'])
run('08-typescript-revision',['git','-C','/workspace/typescript-source','rev-parse','HEAD'])
run('10-gofmt',['gofmt','-l','internal/lower/'])
run('11-vet',['go','vet','./internal/lower/'])
run('12-focused',['go','test','-count=1','-run','TestAGenericTableCast|TestASetTimeout|TestAGenericMethod|TestAGenericFunction|TestAHostMethod|TestNodeWatch','-v','./internal/lower/'])
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source')
(root/'20-whole.env').write_text('ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source\n')
run('20-whole',['go','test','-count=1','-timeout','40m','-skip','^TestProduct_','./internal/lower/'],env=env)
run('21-final-restoration',['git','diff','--exit-code',pin,'--'])
run('22-final-status',['git','status','--short'])
