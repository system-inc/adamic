import json,os,subprocess,sys,difflib
from pathlib import Path
repo=Path(__file__).resolve().parents[3]
if len(sys.argv) != 2: raise SystemExit('provide a scratch output directory')
root=Path(sys.argv[1]).resolve();root.mkdir(parents=True,exist_ok=True)
def run(name,relative,change,pattern,catcher):
 original=repo/relative;source=original.read_text();mutated=change(source);assert mutated!=source
 (root/(name+'.patch')).write_text(''.join(difflib.unified_diff(source.splitlines(True),mutated.splitlines(True),fromfile=relative,tofile=relative)))
 replacement=root/(name+'.go.txt');replacement.write_text(mutated)
 overlay=root/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(original):str(replacement)}}))
 env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1'
 log=root/(name+'.log')
 with log.open('w') as stream:
  result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/oracle','-run',pattern,'-count=1','-timeout','30m','-v'],cwd=repo,env=env,stdout=stream,stderr=subprocess.STDOUT)
 text=log.read_text()
 assert result.returncode!=0, name+' survived'
 assert catcher in text, name+' failed without its intended catcher: '+text[-2500:]
 assert 'error:' not in text and 'build failed' not in text, name+' failed at compilation'
 print(name,'caught:',catcher,flush=True)
def snapshot(source):
 needle='\t\terr = l.readonlyArrayValue(node, value)\n'
 assert source.count(needle)==1
 return source.replace(needle,needle+'\t\tif err == nil && node.Parent != nil && node.Parent.Kind == ast.KindForOfStatement && node.Parent.AsForInOrOfStatement().Expression == node && l.readonlyArrayView(l.checker.GetTypeAtLocation(node)) != nil {\n\t\t\tvalue = ir.ArraySlice{Array: value}\n\t\t}\n',1)
def weak(source):
 needle='\t\t\tif base := l.readonlyArrayView(contextual); base != nil {\n\t\t\t\tcontextual = base\n\t\t\t}\n'
 assert source.count(needle)==1
 return source.replace(needle,'',1)
def override(source):
 needle='func (l *lowering) intrinsicIteratorWrite(node *ast.Node) error {\n'
 assert source.count(needle)==1
 return source.replace(needle,needle+'\treturn nil\n',1)
run('snapshot-view','internal/lower/expression.go',snapshot,'^TestNativeAgreesWithNode$/stage3/fixtures/iteration/array_view_stress.a$', 'stdout')
run('strong-weak-slot','internal/lower/object.go',weak,'^TestNativeAgreesWithNode$/stage3/fixtures/iteration/array_view_weak.a$', 'AddressSanitizer: SEGV')
run('allow-intrinsic-write','internal/lower/iteration_intrinsic.go',override,'^TestStep20IntrinsicIteratorWrites$/intrinsic_iterator_alias.a$', 'want ruled intrinsic iterator refusal')

def variance(source):
 needle="\t\tif base := l.readonlyArrayView(member); base != nil {\n\t\t\tfound = append(found, base)\n\t\t\tcontinue\n\t\t}\n"
 assert source.count(needle)==1
 return source.replace(needle,'',1)
run('widen-mutable-element','internal/lower/invariance.go',variance,'^TestStep20ArrayViewVariance$', 'want nested mutable slot refusal')
