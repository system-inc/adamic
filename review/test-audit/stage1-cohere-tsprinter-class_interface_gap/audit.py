import pathlib,subprocess,time,json,os,difflib,re
R=pathlib.Path('/workspace/adamic');P=R/'review/test-audit/stage1-cohere-tsprinter-class_interface_gap';PKG='./stage1/cohere/tsprinter/'
plans=json.loads((P/'plan.json').read_text());original={x['file']:(R/x['file']).read_text() for x in plans};original['internal/lower/lower.go']=(R/'internal/lower/lower.go').read_text(); results=[]
rows=['TestClassInterfaceMethodGap','TestClosedOptionalBooleanFunctionGap','TestTrackedCorpusRoots','TestDocumentsAgainstGoAndPrettier','TestTSPrinterCachedMetadataRelocation','TestExpressionUnitBoxSelection','TestExpressionUnitPlantedDisagreement','TestExpressionUnitUnionRejectsMissingAndRepeated','TestCompilerGaps','TestClosedPrefixUpdateValueGap','TestNumberConstructor'];regex='^('+'|'.join(rows)+')$'
def run(id,cmd,env=None):
 t=time.monotonic()
 with (P/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=R,env=env,stdout=f,stderr=subprocess.STDOUT)
 out=dict(id=id,command=cmd,seconds=time.monotonic()-t,exit=r.returncode);results.append(out);(P/'runs.json').write_text(json.dumps(results,indent=2));print(id,r.returncode,round(out['seconds'],2),flush=True);return r.returncode
def diff(id,f,new):
 (P/(id+'.diff')).write_text(''.join(difflib.unified_diff(original[f].splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
def replace_body(s,start,end,body):
 a=s.index(start);b=s.index(end,a);return s[:a]+start+'\n'+body+'\n}\n\n'+s[b:]
def standalone(x):
 s=original[x['file']]
 if x['id']=='M3':s=replace_body(s,x['from_'],'// Known factories','\treturn nil')
 elif x['id']=='W2':s=replace_body(s,'func expressionUnion(plan *corpusPlan, outputs [][]byte, want []byte) error {','func expressionDisagreement','\treturn nil')
 else:
  assert x['from_'] in s,x;s=s.replace(x['from_'],x['to'],1)
 return s
for x in plans:diff(x['id'],x['file'],standalone(x))
# Validate standalone Go diffs separately against pristine source, before the switch.
for x in plans:
 if not x['file'].endswith('.go'):continue
 (R/x['file']).write_text(standalone(x))
 try:run(x['id']+'-vet',['timeout','90','go','vet','./'+str(pathlib.Path(x['file']).parent)+'/'])
 finally:(R/x['file']).write_text(original[x['file']])
# One Go switch build, compiler selectors are not included in standalone diffs.
audit=R/'internal/lower/audit_switch.go';audit.write_text('package lower\nimport "os"\nfunc auditActive(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\nfunc auditNumber(id string, value, changed float64) float64 { if auditActive(id) { return changed }; return value }\nfunc auditOption(id, value, changed string) string { if auditActive(id) { return changed }; return value }\n')
for x in plans:
 if x['id'] not in ['M3','M4','M5','M6']:continue
 s=original[x['file']]
 if x['id']=='M3':s=s.replace(x['from_'],x['from_']+'\n\tif auditActive("M3") { return nil }',1)
 if x['id']=='M4':s=s.replace(x['from_'],'Right: ir.NumberConstant{Value: auditNumber("M4", 1, 2)}}, Checked: l.checked(local)',1)
 if x['id']=='M5':s=s.replace(x['from_'],'return ir.NumberCall{Function: auditOption("M5", "convert", "parseFloat"), Arguments: []ir.Expression{value}}, nil',1)
 if x['id']=='M6':s=s.replace(x['from_'],'if auditActive("M6") { return ir.MaybeOf{Value: ir.BooleanConstant{Value: false}, Of: to} }; '+x['from_'],1)
 (R/x['file']).write_text(s)
s=original['internal/lower/lower.go'];s=s.replace('func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\tif auditActive("P-Lower") { return nil, nil }',1);(R/'internal/lower/lower.go').write_text(s)
try:
 for id in ['M3','M4','M5','M6','P-Lower']:
  env=os.environ.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/workspace/u140-cache/'+id
  run(id,['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run',regex],env)
  if 'panic:' in (P/(id+'.log')).read_text():
   for row in rows:run(id+'-'+row,['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run','^'+row+'$'],env)
finally:
 for f,s in original.items():(R/f).write_text(s)
 audit.unlink()
# Port mutants each have their own build and remain separate from Go mutations.
for x in plans[:2]:
 (R/x['file']).write_text(standalone(x));env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/workspace/u140-cache/'+x['id']
 try:
  run(x['id']+'-build',['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/tsprinter/docMain.ts','-o','/workspace/u140-tmp/'+x['id'],'--sanitize'],env)
  run(x['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run',regex],env)
 finally:(R/x['file']).write_text(original[x['file']])
# Empty document answer, replace the complete body so narrowing is not invalidated by unreachable code.
f='stage1/cohere/tsprinter/doc.ts';s=original[f];a=s.index('    print(root: number): string {');b=s.index('    /** @mutates output',a);new=s[:a]+"    print(root: number): string { return ''; }\n"+s[b:];diff('P-Documents',f,new);(R/f).write_text(new)
try:
 run('P-Documents-build',['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/tsprinter/docMain.ts','-o','/workspace/u140-tmp/P-Documents','--sanitize'])
 run('P-Documents',['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run','^TestDocumentsAgainstGoAndPrettier$'])
finally:(R/f).write_text(original[f])
# Construction and witness edits are isolated and restored individually.
for x in plans[6:]:
 (R/x['file']).write_text(standalone(x))
 tests={'S1':'TestTrackedCorpusRoots','S2':'TestTSPrinterCachedMetadataRelocation','S3':'TestExpressionUnitBoxSelection','W1':'TestExpressionUnitPlantedDisagreement','W2':'TestExpressionUnitUnionRejectsMissingAndRepeated','W3':'TestMutants_000'}
 try:run(x['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run','^'+tests[x['id']]+'$'])
 finally:(R/x['file']).write_text(original[x['file']])
