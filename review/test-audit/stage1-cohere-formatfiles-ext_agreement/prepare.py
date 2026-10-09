from pathlib import Path
import subprocess,json,difflib
p=Path('review/test-audit/stage1-cohere-formatfiles-ext_agreement');file='stage1/cohere/formatfiles/golang.ts';src=subprocess.check_output(['git','show','HEAD:'+file],text=True)
plan=[('M01','index >= 0','index > 0','off-by-one bound'),('M02',"path[index] === '.'","path[index] !== '.'",'flip condition'),('M03',"path[index] !== '/'","path[index] !== '\\\\'",'change constant')]
metadata=[]
for mid,old,new,menu in plan:
 assert src.count(old)==1
 changed=src.replace(old,new)
 (p/(mid+'.diff')).write_text(''.join(difflib.unified_diff(src.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 metadata.append(dict(id=mid,file=file,line=src[:src.index(old)].count('\n')+1,before=old,after=new,menu=menu))
(p/'plan.json').write_text(json.dumps(metadata,indent=2))
(p/'reached-functions.json').write_text(json.dumps({'functions':[{'file':file,'line':45,'function':'ext','calls':['String.slice (Node builtin)']}],'scope':'Target row imports golang.ts but calls only ext; clean and other exports are not invoked by ext. Go filepath.Ext and the copied-driver helpers are oracle and harness.'},indent=2))
start=src.index('export function ext(');end=src.index('\n// filepath.Join',start)
probe=src[:start]+"export function ext(path: string): string {\n\treturn '';\n}\n"+src[end:]
(p/'P01.diff').write_text(''.join(difflib.unified_diff(src.splitlines(True),probe.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
# Source instrumentation retains existing harness replacement markers exactly once.
changed=src.replace("import { clean }", "import { readTextFile } from 'adamic';\nconst auditRead = readTextFile('/workspace/u090-selector');\nconst auditMutant = auditRead.kind === 'Ok' ? auditRead.text : '';\n\nimport { clean }")
changed=changed.replace('index >= 0 &&',"index >= 0 && (auditMutant !== 'M01' || index > 0) &&")
changed=changed.replace("path[index] === '.'", "(path[index] === '.') !== (auditMutant === 'M02')")
changed=changed.replace("path[index] !== '/'", "path[index] !== (auditMutant === 'M03' ? '\\\\' : '/')")
changed=changed.replace('export function ext(path: string): string {',"export function ext(path: string): string {\n\tif (auditMutant === 'P01') { return ''; }")
Path(file).write_text(changed)
(p/'selector.diff').write_text(''.join(difflib.unified_diff(src.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
Path('/workspace/u090-selector').write_text('control')
(p/'scope.json').write_text(json.dumps({'base':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'target':'TestExtAgreesWithGo','selector_file':'/workspace/u090-selector','code_under_test':'TypeScript ext executed directly on Node; native builds checked for all standalone source diffs','oracle':'Go filepath.Ext executed at runtime; built-in negative control also expects hand-written empty JSON output','setup_seconds':0,'nproc':5},indent=2))
