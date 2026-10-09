import pathlib,json,subprocess,os,time,difflib
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/internal-oracle-stack';source=root/'internal/native/runtime/stack.c';base=source.read_text();plan=json.loads((p/'plan.json').read_text());results=[];env=os.environ.copy()
env.pop('ADAMIC_NATIVE_SPLIT',None);env.pop('ADAMIC_NATIVE_JOBS',None);env['ADAMIC_GATE_UNCACHED']='1'
patterns=[('stack-rows','^(TestLongArgumentsLeaveTheStackItsLimit|TestSmallStacksStillPanic)$'),('recursion-neighbors','^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(stack_overflow[.]a|library_fnexpr_recurse[.]a|route_targets_recursive[.]a)$'),('new-tests','^(TestFractionalPowersReachRuntime|TestReviewPrograms.*)$')]
def run(ident,group,cmd):
 e=env.copy();e['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-stack/cache/'+ident;start=time.monotonic();log=ident+'-'+group+'.log'
 with (p/log).open('w') as out:q=subprocess.run(cmd,cwd=root,env=e,stdout=out,stderr=subprocess.STDOUT)
 ev=[]
 for s in (p/log).read_text().splitlines():
  try:ev.append(json.loads(s))
  except:pass
 r=dict(mutant=ident,group=group,command=cmd,cache=e['ADAMIC_BUILD_CACHE_DIR'],wall=round(time.monotonic()-start,3),exit=q.returncode,failed=[x['Test'] for x in ev if x['Action']=='fail' and x.get('Test')],passed=[x['Test'] for x in ev if x['Action']=='pass' and x.get('Test')],skipped=[x['Test'] for x in ev if x['Action']=='skip' and x.get('Test')],outputs=[x['Output'].rstrip() for x in ev if x.get('Test','').startswith('TestLongArgumentsLeaveTheStackItsLimit') and ('stack_test.go:' in x.get('Output','') or 'native:' in x.get('Output',''))]);results.append(r);(p/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(ident,group,r['wall'],q.returncode,r['failed'],flush=True);return r
# Uncached clean reached matrix before mutation, including every added top-level test.
for group,pattern in patterns[:2]:
 rr=run('clean',group,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern]);assert rr['exit']==0
flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls']
try:
 for m in plan:
  assert base.count(m['old'])==1
  new=base.replace(m['old'],m['new'],1);source.write_text(new)
  (p/(m['mutant']+'.diff')).write_text(''.join(difflib.unified_diff(base.splitlines(True),new.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
  for mode,options in [('release',['-O2']),('sanitized',['-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all'])]:
   rr=run(m['mutant'],'validate-'+mode,['clang']+flags+options+['-I','internal/native/runtime','-c',m['file'],'-o','/tmp/defend-stack/'+m['mutant']+'-'+mode+'.o']);assert rr['exit']==0
  for group,pattern in patterns:run(m['mutant'],group,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern])
  source.write_text(base)
  runs=[x for x in results if x['mutant']==m['mutant'] and not x['group'].startswith('validate')]
  top={t.split('/')[0] for x in runs for t in x['failed']}
  if top=={'TestLongArgumentsLeaveTheStackItsLimit'} and all(x['exit']==0 for x in runs if x['group']!='stack-rows'):break
finally:source.write_text(base)
