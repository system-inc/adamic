from pathlib import Path
import json,subprocess,os
root=Path(__file__).resolve().parents[3];out=Path(__file__).resolve().parent
first='925f65b0d2eef508e07827ab930fba50bec1b693'
def original(path):return subprocess.check_output(['git','show',first+':'+path],cwd=root,text=True,timeout=20)
path='internal/lower/interface_cast.go';fixed=(root/path).read_text();a=original(path)
start=a.index('\t// A failed whole-object adapter');end=a.index('\tseen := map[ir.ViewContractID]bool{}',start)
needle='\tfields := map[string]bool{}\n\tseen := map[ir.ViewContractID]bool{}'
assert needle in fixed
fixed=fixed.replace(needle,'\tfields := map[string]bool{}\n'+a[start:end]+'\tseen := map[ir.ViewContractID]bool{}',1)
for function in ['view','legacyView']:
 start=fixed.index('func (l *lowering) '+function+'(');end=fixed.find('\nfunc ',start+1)
 if end<0:end=len(fixed)
 block=fixed[start:end]
 needle='\tif l.result.CheckedFields == nil {'
 assert needle in block
 block=block.replace(needle,'\tif err := l.refuseViewElementReads(fields); err != nil {\n\t\treturn nil, err\n\t}\n'+needle,1)
 fixed=fixed[:start]+block+fixed[end:]
interface=out/'combined-interface.go.txt';interface.write_text(fixed)
element=out/'combined-element.go.txt';element.write_text(original('internal/lower/element_access_fields.go'))
overlay=out/'combined-overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/path):str(interface),str(root/'internal/lower/element_access_fields.go'):str(element)}}))
command=['go','build','-overlay',str(overlay),'-o','/tmp/loom-views-combined','./cmd/adamic']
with (out/'combined-build.log').open('w') as log:subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=120,check=True)
results=[]
for name,expected in [('views_p05','view member read by element access'),('native_p37','checked view read of field'),('native_p21','checked view read of field')]:
 source=out/(name+'.a');source.write_text(original('internal/lower/testdata/loom_views_a/'+name+'.a'))
 for backend in ['c','js']:
  cmd=['/tmp/loom-views-combined',backend,str(source)]
  p=subprocess.run(cmd,cwd=root,capture_output=True,text=True,timeout=30)
  assert p.returncode==1 and expected in p.stderr and str(source) in p.stderr,(name,backend,p.returncode,p.stderr)
  results.append(dict(source=name,backend=backend,exit=p.returncode,stderr=p.stderr))
(out/'combination-results.json').write_text(json.dumps(results,indent=2)+'\n')
