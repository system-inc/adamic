from pathlib import Path
import subprocess,json,time,os
p=Path('review/test-audit/stage1-cohere-lint-profile_compilation_main');root=Path.cwd();base=lambda f:subprocess.check_output(['git','show','HEAD:'+f],text=True)
f='stage1/cohere/lint/main.ts';a=base(f);b=a.replace("import { RuleContext }", "import { RuleContext }")
b=b.replace("import { panic, programArguments, readTextFile } from 'adamic';", "import { panic, programArguments, readTextFile } from 'adamic';\nimport { auditSelected } from './audit_selector.ts';")
b=b.replace('function run(row: string, countOnly: boolean): number {','function run(row: string, countOnly: boolean): number {\n    if(auditSelected(\'P01\')) { return 0; }')
b=b.replace('const column = start - (offsets[lines[line - 1] ?? 0] ?? 0) + 1;','const column = start - (offsets[lines[line - 1] ?? 0] ?? 0) + (auditSelected(\'M01\') ? 2 : 1);')
b=b.replace('const fixed = linter.fixed();',"const fixed = auditSelected('M02') ? '' : linter.fixed();")
b=b.replace('        console.log(`fixed\\t${written(fixed)}`);',"        if(!auditSelected('M02')) { console.log(`fixed\\t${written(fixed)}`); }")
Path(f).write_text(b)
f='stage1/cohere/lint/lint.ts';a=base(f);b="import { auditSelected } from './audit_selector.ts';\n"+a;b=b.replace('        rules.visit(index, parent);',"        if(!auditSelected('M03')) { rules.visit(index, parent); }");Path(f).write_text(b)
helper=Path('stage1/cohere/lint/audit_selector.ts');helper.write_text("import { readTextFile } from 'adamic';\nconst selection = readTextFile('/workspace/u111-selector');\nconst selected = selection.kind === 'Ok' ? selection.text : '';\nexport function auditSelected(id: string): boolean { return selected === id; }\n")
Path('/workspace/u111-selector').write_text('control')
(p/'selector.diff').write_text(subprocess.check_output(['git','diff'],text=True)+'--- /dev/null\n+++ b/'+str(helper)+'\n@@ -0,0 +1,4 @@\n'+''.join('+'+l+'\n' for l in helper.read_text().splitlines()))
rows=['TestProfileCompilationUnion','TestProfileCompilationPlantedFailure','TestProfileCompilationBuildLower','TestProfileCompilationBuildC','TestProfileCompilationBuildJavaScript','TestProfileCompilationBuildNative','TestProfileCompilation_Setup','TestProfileCompilation_000','TestCommentFoldMutant','TestPositionIndexMutant','TestRegistrationMutant','TestFactoryHooks_Setup','TestNestedOutsideModuleCopy']
results=[]
def run(mid,names,logname):
 env=dict(os.environ);env.pop('ADAMIC_LINT_PROFILE_DIR',None);env.pop('ADAMIC_LINT_PROFILE_SNAPSHOTS',None)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run','^('+'|'.join(names)+')$'];st=time.monotonic();logpath=p/(logname+'.log')
 with logpath.open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
 events=[]
 for line in logpath.read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 x=dict(id=mid,log=logpath.name,command=cmd,status=r.returncode,wall=time.monotonic()-st,binary_seconds=next((e.get('Elapsed') for e in reversed(events) if e.get('Action') in ['pass','fail'] and 'Test'not in e),None),failed_tests=[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test') in names],ran=[e['Test'] for e in events if e.get('Action') in ['pass','fail'] and e.get('Test') in names],cooked=any('test timed out' in e.get('Output','') for e in events));results.append(x);(p/'matrix.json').write_text(json.dumps(results,indent=2));return x
# Bound cold stages independently; keep the actual switch neutral during preparation.
x=run('build-c',['TestProfileCompilationBuildC'],'switch-build-c')
if x['status']:raise SystemExit('switched C preparation failed or cooked')
x=run('build-native',['TestProfileCompilationBuildNative'],'switch-build-native')
if x['status']:raise SystemExit('switched native preparation failed or cooked')
x=run('control',rows,'control')
if x['status']:raise SystemExit('switched control failed or cooked')
for mid in ['M01','M02','M03','P01']:
 Path('/workspace/u111-selector').write_text(mid);x=run(mid,rows,mid)
 if x['cooked']:
  for row in ['TestProfileCompilation_000','TestProfileCompilationBuildLower','TestProfileCompilationBuildC','TestProfileCompilationBuildJavaScript','TestProfileCompilationBuildNative','TestProfileCompilation_Setup']:
   run(mid,[row],mid+'-'+row)
Path('/workspace/u111-selector').write_text('control')
