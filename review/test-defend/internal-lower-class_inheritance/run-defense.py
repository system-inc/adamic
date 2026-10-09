import difflib,json,pathlib,subprocess,time,os
p=pathlib.Path('review/test-defend/internal-lower-class_inheritance')
plans=[
 dict(id='D1',target='TestInheritanceKeepsCheckerConstructorRules',file='internal/load/load.go',old='func (p *Program) formatDiagnostic(diagnostic *ast.Diagnostic) string {',new='func (p *Program) formatDiagnostic(diagnostic *ast.Diagnostic) string {\n\tif diagnostic.Code() == 2511 {\n\t\treturn ""\n\t}',kind='return early',reason='Return an empty formatted diagnostic for TS2511 (abstract-class construction), leaving checker results and private-scope diagnostics intact.'),
 dict(id='D2',target='TestInheritanceRefusesThisBeforeSuperReturns',file='internal/lower/class.go',old='Fix: "call super(...) before using this"',new='Fix: "call super before using this"',kind='change constant',reason='Lose the explicit call syntax in the pre-super repair; the conditional row only requires the weaker call-super prefix.'),
 dict(id='D3',target='TestInheritanceCycleFinderIncludesInheritedFields',file='internal/lower/cycles.go',old='\t\t\tfields = append(fields, property)\n',new='\t\t\tfields = append(fields, property)\n\t\t\treturn fields\n',kind='return early',reason='Return after the first data field, losing an inherited field behind an own field, while a subclass with just an inherited generic slot can retain that slot.'),
 dict(id='D4',target='TestInheritanceCycleFinderIncludesInheritedFields',file='internal/lower/cycles.go',old='\tfor _, property := range f.l.checker.GetPropertiesOfType(proven) {\n',new='\tfor _, property := range f.l.checker.GetPropertiesOfType(proven) {\n\t\tif property.Parent != proven.Symbol() {\n\t\t\treturn fields\n\t\t}\n',kind='return early',reason='Stop on a property declared by another class, incorrectly treating the inherited-field boundary as the end of the shape.'),
 dict(id='D5',target='TestInheritanceCycleFinderIncludesInheritedFields',file='internal/lower/cycles.go',old='\treturn fields\n}',new='\treturn fields[:max(0, len(fields)-1)]\n}',kind='off by one bound',reason='Omit the final selected field, which is the inherited parent after the child label in the direct traversal witness.')]
original={m['file']:pathlib.Path(m['file']).read_text() for m in plans}
for m in plans:
 s=original[m['file']];assert s.count(m['old'])==1,(m['id'],s.count(m['old']));m['line']=s[:s.index(m['old'])].count('\n')+1
 (p/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(m['old'],m['new'],1).splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
(p/'plan.json').write_text(json.dumps(plans,indent=2)+'\n')
results=[]
def run(id,cmd,env):
 start=time.monotonic()
 with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
 events=[]
 for line in (p/(id+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except (ValueError,TypeError):pass
 result=dict(id=id,command=cmd,exit=r.returncode,wall=time.monotonic()-start,failed=sorted({e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']}),passed=sorted({e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']}),skipped=sorted({e['Test'] for e in events if e.get('Action')=='skip' and e.get('Test')}),binary_seconds=next((e.get('Elapsed') for e in reversed(events) if e.get('Action') in ['pass','fail'] and not e.get('Test')),None),fail_lines=[e['Output'].strip() for e in events if '.go:' in e.get('Output','') and e.get('Test')])
 results.append(result);(p/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(id,result['exit'],result['wall'],result['failed'],flush=True);return result
for m in plans:
 source=pathlib.Path(m['file']); env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR='/tmp/defend-class-inheritance/cache/'+m['id'])
 try:
  source.write_text(original[m['file']].replace(m['old'],m['new'],1))
  package='./internal/load/' if m['file'].startswith('internal/load') else './internal/lower/'
  checked=run(m['id']+'-vet',['go','vet',package],env)
  if checked['exit']==0:run(m['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],env)
 finally:source.write_text(original[m['file']])
