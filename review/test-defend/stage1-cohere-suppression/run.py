import pathlib,subprocess,json,difflib,time,os
p=pathlib.Path(__file__).resolve().parent
(p/'base.txt').write_text(subprocess.check_output(['git','rev-parse','origin/main'],text=True))
for n in ['REPORT.md','rows.json','frozen-plan.json','matrix.json','list.log']:(p/('prior-'+n)).write_bytes(subprocess.check_output(['git','show','origin/test-audit/stage1-cohere-suppression:review/test-audit/stage1-cohere-suppression/'+n]))
path=pathlib.Path('internal/native/emit_expressions.go');s=path.read_text();old='return fmt.Sprintf("((%s) ? &adamic_string_true : &adamic_string_false)", e.value(expression.Value))';new='return fmt.Sprintf("((%s) ? &adamic_string_false : &adamic_string_true)", e.value(expression.Value))';assert s.count(old)==1
r=dict(mutant='D01',file_line=str(path)+':'+str(s[:s.index(old)].count('\n')+1),change='change BooleanToString format constant to swap true and false output strings',old=old,new=new)
dp=p/'D01.diff';dp.write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(old,new).splitlines(True),fromfile='a/'+str(path),tofile='b/'+str(path))))
subprocess.run(['git','apply','--check',str(dp)],check=True);subprocess.run(['git','apply',str(dp)],check=True)
try:
 with (p/'D01-vet.log').open('w') as f:assert subprocess.run(['timeout','120','go','vet','./internal/native/'],stdout=f,stderr=subprocess.STDOUT).returncode==0
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/workspace/suppression-defend-cache/D01'
 r['command']='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/suppression/ -run . > D01.log 2>&1'
 start=time.monotonic()
 with (p/'D01.log').open('w') as f:ret=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/suppression/','-run','.'],env=env,stdout=f,stderr=subprocess.STDOUT)
 r['wall_seconds']=time.monotonic()-start;r['exit']=ret.returncode;events=[]
 for l in (p/'D01.log').read_text().splitlines():
  try:events.append(json.loads(l))
  except ValueError:pass
 r['rows_failed']=sorted({e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test')});r['rows_passed']=sorted({e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']});r['subcases_passed']=[e['Test'] for e in events if e.get('Action')=='pass' and '/' in e.get('Test','')];r['failures']=[{'test':e.get('Test'),'line':e['Output'].strip()} for e in events if e.get('Action')=='output' and e.get('Test','').split('/')[0] in r['rows_failed'] and '.go:' in e.get('Output','')];r['skipped']=[e['Test'] for e in events if e.get('Action')=='skip' and e.get('Test')];r['binary_seconds']=[e.get('Elapsed') for e in events if not e.get('Test') and e.get('Action')=='fail']
 (p/'matrix.json').write_text(json.dumps([r],indent=2)+'\n');print(json.dumps(r,indent=2))
finally:subprocess.run(['git','apply','-R',str(dp)],check=True)
