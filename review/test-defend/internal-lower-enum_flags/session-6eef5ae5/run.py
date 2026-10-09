import pathlib,subprocess,time,json,difflib,os
root=pathlib.Path('/workspace/adamic');out=root/'review/test-defend/internal-lower-enum_flags';base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root).decode().strip()
plans=[
('D01','internal/lower/assignments.go','ast.KindAmpersandEqualsToken:                         ast.KindAmpersandToken,','ast.KindAmpersandEqualsToken:                         ast.KindBarToken,','change option: compound AND becomes OR'),
('D02','internal/lower/switch.go','for index, group := range groups {','for index, group := range groups[func() int { if unmatchedCheck != nil && l.function != nil && l.function.Returns == ir.String { return 1 }; return 0 }():] {','off-by-one bound: omit first dispatch group in checked string-returning enum switches'),
('D03','internal/lower/object.go','literal.Elements = append(literal.Elements, value)','if !(node.Parent != nil && node.Parent.Kind == ast.KindForOfStatement && l.enumMember(item) != nil && l.checker.GetConstantValue(l.enumMember(item)) == float64(0)) { literal.Elements = append(literal.Elements, value) }','drop statement: omit zero-valued enum member from inline for-of array')]
(out/'plan.json').write_text(json.dumps({'base':base,'mutants':plans},indent=2))
def covered(n):return {s.rsplit(' ',2)[0] for s in (out/(n+'.cover')).read_text().splitlines()[1:] if int(s.rsplit(' ',1)[1])>0}
for r,s in [('TestFlagEnumsDomain','TestFlagEnumInlineIteration'),('TestEnumNeverDefault','TestNumericEnumNeverProof'),('TestFlagEnumInlineIteration','TestEnumNameEnumeration')]:
 (out/(r+'.exclusive.txt')).write_text('Subsumer: '+s+'\n'+'\n'.join(sorted(covered(r)-covered(s)))+'\n')
for n in ['audit.json','audit-notes.txt','plan.json','matrix-index.json','reached-functions.txt']:
 (out/('prior-'+n)).write_bytes(subprocess.check_output(['git','show','origin/test-audit/internal-lower-enum_flags:review/test-audit/internal-lower-enum_flags/'+n],cwd=root))
results=[]
for mid,name,before,after,desc in plans:
 p=root/name;original=p.read_text();assert before in original
 changed=original.replace(before,after,1);line=original[:original.index(before)].count('\n')+1
 (out/(mid+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+name,tofile='b/'+name)))
 p.write_text(changed)
 try:
  start=time.monotonic()
  with (out/(mid+'.vet.log')).open('w') as f:v=subprocess.run(['go','vet','./internal/lower/'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
  if v.returncode:raise RuntimeError('vet failed '+mid)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/enum-defense/cache/'+mid
  with (out/(mid+'.log')).open('w') as f:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[json.loads(x) for x in (out/(mid+'.log')).read_text().splitlines() if x.startswith('{')]
  results.append({'mutant':mid,'file_line':name+':'+str(line),'change':desc,'exit':r.returncode,'seconds':time.monotonic()-start,'rows_failed':[e['Test'] for e in events if e['Action']=='fail' and 'Test' in e and '/' not in e['Test']],'rows_passed':[e['Test'] for e in events if e['Action']=='pass' and 'Test' in e and '/' not in e['Test']],'fail_lines':[e.get('Output','').strip() for e in events if e['Action']=='output' and '.go:' in e.get('Output','')],'cooked':any('panic: test timed out' in e.get('Output','') for e in events)})
  (out/'matrix.json').write_text(json.dumps(results,indent=2))
  print(mid,results[-1]['rows_failed'],flush=True)
 finally:p.write_text(original)
