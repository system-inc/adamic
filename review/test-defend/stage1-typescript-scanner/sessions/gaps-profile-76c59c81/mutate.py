from pathlib import Path
import subprocess,os,json,time,difflib
p=Path('/tmp/defend-scanner/evidence');pkg='./stage1/typescript/scanner/';rows=['TestGapStandsWhereGapsMdSays','TestBigintGapStandsWhereGapsMdSays','TestProfileSnapshotsAgree'];plan=[
 dict(id='D1',row=rows[0],file='internal/lower/object.go',old='"push with other than one value"',new='"push with two values"',reason='exclusive multi-argument push refusal diagnostic',menu='change constant'),
 dict(id='D2',row=rows[1],file='internal/lower/expression.go',old='"a value of type "+l.checker.TypeToString(l.checker.GetTypeAtLocation(node))',new='"a representation of type "+l.checker.TypeToString(l.checker.GetTypeAtLocation(node))',reason='exclusive unsupported bigint representation diagnostic',menu='change constant'),
 dict(id='D3',row=rows[2],file='internal/native/runtime/heap.c',old='each->free = ((free_slot *)slot)->next;',new='',reason='release slab free-list reuse, bypassed by sanitizer allocator',menu='drop statement'),
 dict(id='D4',row=rows[2],file='internal/native/runtime/heap.c',old='each->fresh += size;',new='each->fresh += size - 1;',reason='release slab fresh-slot bound, bypassed by sanitizer allocator',menu='off by one'),
 dict(id='D5',row=rows[2],file='internal/native/native.go',old='return append(flags, "-O2")',new='return append(flags, "-O2", "-funsafe-math-optimizations")',reason='release-only compiler floating-point option; sanitized branch returns before it',menu='change option')]
for x in plan:
 old=Path(x['file']).read_text();assert old.count(x['old'])==1,x;new=old.replace(x['old'],x['new'],1) if x['new'] else old.replace('\t\t'+x['old']+'\n','',1);x['line']=old[:old.index(x['old'])].count('\n')+1;(p/(x['id']+'.diff')).write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+x['file'],tofile='b/'+x['file'])))
(p/'defense-plan.json').write_text(json.dumps(plan,indent=2));results=json.loads((p/'defense-runs.json').read_text())
for x in plan:
 if any(r['id']==x['id'] for r in results):continue
 # Once the profile row is uniquely defended, further attempts on it are unnecessary.
 if x['row']==rows[2] and any(r['unique_target'] and r['target']==rows[2] for r in results):break
 id=x['id'];diff=p/(id+'.diff');subprocess.run(['git','apply','--check',str(diff)],check=True);subprocess.run(['git','apply',str(diff)],check=True)
 committed=x['file'].endswith('.c')
 if committed:
  subprocess.run(['git','add',x['file']],check=True);subprocess.run(['git','commit','-m','Record '+id+' runtime variant for defender replay'],check=True,stdout=subprocess.DEVNULL)
 try:
  if x['file'].endswith('.go'):cmd=['go','vet','./'+str(Path(x['file']).parent)+'/']
  else:cmd=['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2','-I','internal/native/runtime','-c',x['file'],'-o','/tmp/defend-scanner/'+id+'.o']
  start=time.monotonic()
  with (p/(id+'-validation.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
  assert r.returncode==0,(id,'validation')
  env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/tmp/defend-scanner/typescript',ADAMIC_SCANNER_BENCH='1',ADAMIC_SCANNER_PROFILE_DIR='/tmp/defend-scanner/artifacts/'+id,ADAMIC_SCANNER_PROFILE_SNAPSHOTS='/tmp/defend-scanner/artifacts/'+id,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-scanner/cache/'+id)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.']
  with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
  events=[]
  for line in (p/(id+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except ValueError:pass
  failed=[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']];passed=[e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']];failing_lines=[e['Output'].strip() for e in events if e.get('OutputType')=='error'];result=dict(id=id,target=x['row'],command=cmd,environment=env|{},exit=r.returncode,seconds=time.monotonic()-start,failed=failed,passed=passed,failing_lines=failing_lines,unique_target=failed==[x['row']]);result['environment']={k:v for k,v in env.items() if k.startswith('ADAMIC_')};results.append(result);(p/'defense-runs.json').write_text(json.dumps(results,indent=2));print({k:result[k] for k in ['id','exit','seconds','failed','unique_target']},flush=True)
 finally:
  subprocess.run(['git','apply','-R',str(diff)],check=True)
  if committed:
   subprocess.run(['git','add',x['file']],check=True);subprocess.run(['git','commit','-m','Restore runtime after '+id+' defender measurement'],check=True,stdout=subprocess.DEVNULL)
assert subprocess.run(['git','diff','--exit-code']).returncode==0
with (p/'restored.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','^(TestGapStandsWhereGapsMdSays|TestBigintGapStandsWhereGapsMdSays)$'],stdout=log,stderr=subprocess.STDOUT)
assert r.returncode==0
