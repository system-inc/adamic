from pathlib import Path
import json,subprocess,os,time,re,difflib
root=Path('/workspace/adamic');p=root/'review/test-audit/stage1-cohere-estree-deep';os.chdir(root)
menu=json.loads((p/'menu.json').read_text());base={m['file']:subprocess.check_output(['git','show','origin/main:'+m['file']],text=True) for m in menu}
helper=root/'stage1/cohere/estree/auditMutant.ts';selector=Path('/tmp/u085/selector');selector.write_text('')
helper.write_text("import { readTextFile } from 'adamic';\nconst choice = readTextFile('/tmp/u085/selector');\nexport const auditMutant = choice.kind === 'Error' ? '' : choice.text.trim();\n")
for m in menu:
 f=m['file'];s=base[f];old=m['old'];new=m['new'];mid=m['id']
 s="import { auditMutant } from './auditMutant.ts';\n"+s
 if mid=='M1':new="node.start = auditMutant === 'M1' ? 1 : 0;"
 if mid=='M2':new="        if(auditMutant !== 'M2') {\n"+old+"        }\n"
 if mid=='M3':new="defaulted ? (auditMutant === 'M3' ? 'ExportNamedDeclaration' : 'ExportDefaultDeclaration') : 'ExportNamedDeclaration'"
 if mid=='M4':new="    if(auditMutant !== 'M4') {\n"+old+"    }\n"
 s=s.replace(old,new,1)
 if mid=='M4':s=s.replace('export function answer(path: string, text: string): string {',"export function answer(path: string, text: string): string {\n    if(auditMutant === 'P1') { return ''; }",1)
 Path(f).write_text(s);(p/('scratch-'+Path(f).name+'.txt')).write_text(s)
(p/'scratch-auditMutant.ts.txt').write_text(helper.read_text())
env=dict(os.environ,ADAMIC_ESTREE_LIBRARY='/tmp/u085/library',ADAMIC_BUILD_CACHE_DIR='/tmp/u085/cache/switched')
runs=[]
def run(label,row,variant='',overlay=None):
 selector.write_text(variant)
 cmd=['timeout','120','go','test']
 if overlay:cmd+=['-overlay='+overlay]
 cmd+=['-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run','^'+row+'$']
 start=time.monotonic()
 with (p/(label+'.log')).open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
 wall=time.monotonic()-start;runs.append(dict(label=label,row=row,variant=variant,exit=r.returncode,wall=wall,command='ADAMIC_ESTREE_LIBRARY=/tmp/u085/library ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/switched '+' '.join(cmd),selector=variant));(p/'matrix-runs.json').write_text(json.dumps(runs,indent=2)+'\n')
 print(label,r.returncode,round(wall,3),flush=True);return r.returncode
try:
 # One complete agreement row proves that the selector's empty choice is transparent.
 if run('selector-control','TestDecoratedExports'):raise RuntimeError('selector clean control failed')
 matrixrows=['TestDeepGrammar','TestGeneratedAgreement','TestDecoratedExports','TestRecoveredExpressions','TestUnattachedDecorator']
 for mid in ['M1','M2','M3','M4']:
  rows=matrixrows[:2] if mid=='M2' else matrixrows[:4] if mid in ['M1','M3'] else matrixrows
  for row in rows:run(mid+'-'+row,row,mid)
 for row in matrixrows:run('P1-'+row,row,'P1')
finally:
 selector.write_text('')
 for f in base:subprocess.run(['git','restore','--source=origin/main','--',f],check=True)
 helper.unlink(missing_ok=True)
# Weakened-check witnesses run on restored source, independent of production mutants.
for row in ['TestDecoratedExportMutant','TestRecoveredExpressionMutant']:
 run('W1-'+row,row,overlay='/tmp/u085/weak/W1.json')
run('W2-TestUnattachedDecoratorControl','TestUnattachedDecoratorControl',overlay='/tmp/u085/weak/W2.json')
# Validate each standalone diff through the native tool, not only Go vet.
checks=[]
for mid in ['M1','M2','M3','M4','P1']:
 diff=p/'diffs'/f'{mid}.diff'
 with (p/('apply-'+mid+'.log')).open('w') as out:subprocess.run(['git','apply','--check',str(diff)],stdout=out,stderr=subprocess.STDOUT,check=True)
 subprocess.run(['git','apply',str(diff)],check=True)
 start=time.monotonic();cmd=['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/estree/main.ts','-o','/tmp/u085/standalone-'+mid,'--sanitize']
 try:
  with (p/('native-build-'+mid+'.log')).open('w') as out:r=subprocess.run(cmd,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u085/cache/'+mid),stdout=out,stderr=subprocess.STDOUT)
  checks.append(dict(id=mid,exit=r.returncode,wall=time.monotonic()-start,command='ADAMIC_BUILD_CACHE_DIR=/tmp/u085/cache/'+mid+' '+' '.join(cmd)))
  print('standalone',mid,r.returncode,round(checks[-1]['wall'],3),flush=True)
 finally:
  for f in base:subprocess.run(['git','restore','--source=origin/main','--',f],check=True)
 (p/'native-builds.json').write_text(json.dumps(checks,indent=2)+'\n')
# Record checks for the allowed witness overlays.
for wid in ['W1','W2']:
 with (p/('vet-'+wid+'.log')).open('w') as out:subprocess.run(['go','vet','-overlay=/tmp/u085/weak/'+wid+'.json','./stage1/cohere/estree/'],stdout=out,stderr=subprocess.STDOUT,check=True)
