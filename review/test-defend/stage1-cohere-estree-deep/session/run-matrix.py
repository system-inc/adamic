import os,pathlib,subprocess,json,time,difflib
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/stage1-cohere-estree-deep/session';base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()
file='stage1/cohere/estree/convert.ts';old=(root/file).read_text();spec=[('D1','if(!assignment) {','if(false) {','Disable iterative binary conversion, retaining recursive conversion for short inputs'),('D2',"if(this.kind(child) !== 'Decorator') {","if(true) {",'Stop filtering decorators when locating the export keyword'),('D3',"                  : 'ClassExpression',","                  : 'ClassDeclaration',",'Change recovered class-expression node-kind constant to ClassDeclaration')];menu=[]
for mid,a,b,why in spec:
 assert old.count(a)==1,(mid,old.count(a));new=old.replace(a,b);line=old[:old.index(a)].count('\n')+1
 (p/(mid+'.diff')).write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 menu.append(dict(mutant=mid,file_line=file+':'+str(line),change=why))
(p/'menu.json').write_text(json.dumps(menu,indent=2));(p/'base.txt').write_text(base+'\n')
groups=[['TestDeepGrammar','TestGeneratedAgreement'],['TestDecoratedExports','TestRecoveredExpressions'],['TestRecoveredGrammar','TestUnattachedDecorator','TestDecoratedExportMutant','TestRecoveredExpressionMutant'],['TestAcceptanceGrammar','TestSyntaxGrammar'],['TestJSXAgreement','TestJSXAgreement_[0-9]+','TestBoundedPortParser_[0-9]+'],['TestThroughput'],['TestScalarEdges_[0-9]+','TestCookedSurrogates','TestOriginalLibraries','TestDecoratedExportLibraries']]
(p/'matrix-selectors.json').write_text(json.dumps(groups,indent=2));env=os.environ.copy();env.update(ADAMIC_ESTREE_LIBRARY='/tmp/estree-deep-defense/library',ADAMIC_NATIVE_SPLIT='1',ADAMIC_ESTREE_BENCHMARK='1')
def run(mid,index,names):
 sel='^('+'|'.join(names)+')$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',sel];name=mid+'-'+str(index);start=time.monotonic()
 with (p/(name+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for l in (p/(name+'.log')).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 cooked=r.returncode==124 or any('test timed out' in e.get('Output','') for e in events)
 result=dict(id=mid,group=index,command=cmd,cache=env.get('ADAMIC_BUILD_CACHE_DIR','default'),exit=r.returncode,wall=time.monotonic()-start,cooked=cooked,statuses={e['Test']:e['Action'] for e in events if e.get('Test') and '/' not in e['Test'] and e['Action'] in ['pass','fail','skip']},errors=[dict(test=e.get('Test'),line=e['Output'].strip()) for e in events if e.get('OutputType')=='error'],log=name+'.log')
 (p/(name+'.meta.json')).write_text(json.dumps(result,indent=2));return result
# Profiles already establish clean baseline for groups 0/1. All additional selected controls must pass before mutation.
clean=[]
for i in range(2,len(groups)):
 r=run('clean',i,groups[i]);clean.append(r);(p/'baseline-batches.json').write_text(json.dumps(clean,indent=2))
 if r['exit']!=0:
  print('Baseline batch failed or cooked; stopping before mutation',i,flush=True);raise SystemExit(1)
results=[]
for m in menu:
 mid=m['mutant'];env['ADAMIC_BUILD_CACHE_DIR']='/tmp/estree-deep-defense/cache/'+mid
 try:
  subprocess.run(['git','apply','--check',str(p/(mid+'.diff'))],cwd=root,check=True);subprocess.run(['git','apply',str(p/(mid+'.diff'))],cwd=root,check=True)
  batches=[]
  for i,names in enumerate(groups):
   r=run(mid,i,names);batches.append(r)
   if r['cooked']:
    # Rerun affected top-level names independently; observed passes from cooked run are retained, no timeout kills.
    pending=[n for n in (p/'list.log').read_text().splitlines() if n.startswith('Test') and __import__('re').fullmatch('('+'|'.join(names)+')',n) and r['statuses'].get(n) not in ('pass','skip')]
    for j,n in enumerate(pending):batches.append(run(mid,str(i)+'-alone-'+str(j),[n]))
   (p/(mid+'-batches.json')).write_text(json.dumps(batches,indent=2))
  results.append(dict(**m,batches=batches));(p/'matrix.json').write_text(json.dumps(results,indent=2))
 finally:(root/file).write_text(old)
print('matrix completed',flush=True)
