from pathlib import Path
import json,subprocess,os,time,difflib
p=Path('review/test-defend/internal-lower-parameter_properties')
plan=[('D01','internal/lower/parameter_properties.go','func (l *lowering) parameterPropertyStores(declaration *ast.Node, this int) ([]ir.Statement, error) {','func (l *lowering) parameterPropertyStores(declaration *ast.Node, this int) ([]ir.Statement, error) {\n\tif true { return nil, nil }','Return early without initialization stores'),('D02','internal/lower/class.go','func (l *lowering) setProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {','func (l *lowering) setProperty(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {\n\tif field := l.checker.GetSymbolAtLocation(target.Name()); field != nil && field.ValueDeclaration != nil && parameterProperty(field.ValueDeclaration) { return nil, nil }','Return early on writes to parameter-property fields'),('D03','internal/load/load.go','formatted = append(formatted, p.formatDiagnostic(diagnostic))','if diagnostic.Code() != 2341 { formatted = append(formatted, p.formatDiagnostic(diagnostic)) }','Drop collection of TS2341 private-keyword diagnostics')]
(p/'plan.json').write_text(json.dumps(plan,indent=2)+'\n');results=[]
for mid,file,old,new,change in plan:
 orig=Path(file).read_text();assert orig.count(old)==1
 patch=p/(mid+'.diff');patch.write_text(''.join(difflib.unified_diff(orig.splitlines(True),orig.replace(old,new).splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 subprocess.run(['git','apply','--check',str(patch)],check=True);subprocess.run(['git','apply',str(patch)],check=True)
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/parameter-defend/cache/'+mid
 start=time.monotonic()
 try:
  pkg='./internal/load/' if '/load/' in file else './internal/lower/'
  with open(p/(mid+'-vet.log'),'w') as log:vet=subprocess.run(['timeout','90','go','vet',pkg],stdout=log,stderr=subprocess.STDOUT,env=env).returncode
  row={'mutant':mid,'file_line':file+':'+str(orig[:orig.index(old)].count('\n')+1),'change':change,'vet_exit':vet}
  if vet==0:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'];lp=p/(mid+'.log')
   with open(lp,'w') as log:rc=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env).returncode
   es=[]
   for line in lp.read_text().splitlines():
    try:es.append(json.loads(line))
    except:pass
   row.update(exit=rc,command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),failed=[e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test')],passed=[e['Test'] for e in es if e.get('Action')=='pass' and e.get('Test')],skipped=[e['Test'] for e in es if e.get('Action')=='skip' and e.get('Test')],cooked=any('test timed out' in e.get('Output','') for e in es),diagnostics=[e['Output'].strip() for e in es if e.get('Test')=='TestParameterPropertyCheckerContracts' and '.go:' in e.get('Output','')])
  row['wall_seconds']=round(time.monotonic()-start,3);results.append(row);(p/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(mid,vet,row.get('failed'),flush=True)
 finally:subprocess.run(['git','apply','-R',str(patch)],check=True)
