from pathlib import Path
import subprocess,json,time
p=Path('review/test-audit/stage1-cohere-formatfiles-ext_agreement');matrix=[]
def group(name):
 if name.startswith('TestThePortParsesAsGoCohereDoes'):return 'TestThePortParsesAsGoCohereDoes family'
 if name.startswith('TestProduct_FormatfilesNative') and name!='TestProduct_FormatfilesNativeUnsanitized':return 'TestProduct_FormatfilesNative family'
 return name
for mid in ['M01','M02','M03','P01']:
 Path('/workspace/u090-selector').write_text(mid)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/formatfiles/','-run','.'];s=time.monotonic()
 with (p/(mid+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for l in open(p/(mid+'.log')):
  try:events.append(json.loads(l))
  except:pass
 failed=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test')))
 ran=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='run' and e.get('Test')))
 end=[e for e in events if e.get('Action') in ['pass','fail'] and not e.get('Test')]
 matrix.append(dict(id=mid,selector=mid,command=cmd,status=r.returncode,wall=time.monotonic()-s,seconds=end[-1]['Elapsed'] if end else None,failed_tests=failed,failed_rows=sorted(set(group(t) for t in failed)),tests_run=ran,rows_run=sorted(set(group(t) for t in ran)),cooked=any('test timed out' in e.get('Output','') for e in events),build_events=[e['Output'].strip() for e in events if 'build formatfiles-native' in e.get('Output','')]))
 (p/'matrix.json').write_text(json.dumps(matrix,indent=2))
 if matrix[-1]['cooked']:raise SystemExit(mid+' cooked, needs narrowing')
Path('/workspace/u090-selector').write_text('control')
