import pathlib,subprocess,json,os,time,difflib
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-defend/internal-native-number'; src=root/'internal/native/runtime/dtoa.c'; original=src.read_text()
rows=[x.strip() for x in pathlib.Path('/tmp/defend-number/list.log').read_text().splitlines() if x.startswith('Test') and any(s in x for s in ['Number','Math','Precision','Exponential','String','Radix','ArtifactInputs','DecodeCacheLock','RegExpNativeStepLimitBoundary','RuntimeKeyKeepsBoundaries'])]
regex='^('+'|'.join(rows)+')$'
plan=[('D1',"(buffer[(*length) - 1] - '0') % 2 != 0","(buffer[(*length) - 1] - '0') % 2 == 0"),('D2','in_delta_room_minus = bignum_less(numerator, delta_minus);','in_delta_room_minus = bignum_less_equal(numerator, delta_minus);'),('D3','in_delta_room_plus = bignum_plus_compare(numerator, delta_plus, denominator) > 0;','in_delta_room_plus = bignum_plus_compare(numerator, delta_plus, denominator) >= 0;')]
(out/'plan.json').write_text(json.dumps([{'id':m,'line':original[:original.index(a)].count('\n')+1,'old':a,'new':b} for m,a,b in plan],indent=2));(out/'matrix-rows.json').write_text(json.dumps(rows,indent=2))
results=[]
try:
 for mid,a,b in [('clean',None,None)]+plan:
  if a:
   assert original.count(a)==1,(mid,original.count(a)); edited=original.replace(a,b);src.write_text(edited)
   (out/(mid+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),edited.splitlines(True),fromfile='a/internal/native/runtime/dtoa.c',tofile='b/internal/native/runtime/dtoa.c')))
   with (out/(mid+'-clang.log')).open('w') as f:
    rc=subprocess.run(['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-function','-Iinternal/native/runtime','-c',str(src),'-o','/tmp/defend-number/'+mid+'.o'],cwd=root,stdout=f,stderr=subprocess.STDOUT).returncode
   if rc: raise RuntimeError('clang failed '+mid)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-number/cache/'+mid
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',regex]
  t=time.monotonic()
  with (out/(mid+'.log')).open('w') as f: rc=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT).returncode
  events=[]
  for line in (out/(mid+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except ValueError:pass
  failed=[e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']];passed=[e['Test'] for e in events if e.get('Action')=='pass' and 'Test' in e and '/' not in e['Test']]
  result={'id':mid,'wall_seconds':time.monotonic()-t,'exit':rc,'command':'ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),'rows_failed':failed,'rows_passed':passed,'failure_output':[e['Output'].strip() for e in events if 'Output' in e and e.get('Test') in failed and (': native' in e['Output'] or 'disagree' in e['Output'] or 'differs' in e['Output'])]};results.append(result);(out/'results.json').write_text(json.dumps(results,indent=2)); print(mid,result['wall_seconds'],rc,failed,flush=True)
  if mid=='clean' and rc:break
  if failed==['TestNumbersFormatExactlyAsJavaScriptDoes']:break
finally:src.write_text(original)
