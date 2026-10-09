from pathlib import Path
import subprocess,json,tempfile
out=Path(__file__).parent
cli='/tmp/adamic-train-plus-3-fixed-cli'
results=[]
def run(label,command):
 result=subprocess.run(command,capture_output=True,text=True,timeout=40)
 data=dict(label=label,command=command,exit=result.returncode,stdout=result.stdout,stderr=result.stderr)
 results.append(data)
 (out/'extra-fixtures.json').write_text(json.dumps(results,indent=2)+'\n')
 return result
node=lambda p:['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(p)]
with tempfile.TemporaryDirectory(prefix='adamic-train-plus-3-') as directory:
 for name,path in [('p19','review/compiler/fx6-key-read/p19-native-call-gap.a'),('p70-copy','internal/lower/testdata/tuple_object_views/p70_copy.a')]:
  truth=run(name+'-node',node(path)); assert truth.returncode==0
  binary=Path(directory)/name
  build=run(name+'-build',[cli,'build',path,'-o',str(binary)]); assert build.returncode==0
  native=run(name+'-native',[str(binary)])
  js=run(name+'-js-build',[cli,'js',path]); assert js.returncode==0
  generated=Path(directory)/(name+'.mjs'); generated.write_text(js.stdout)
  backend=run(name+'-javascript',node(generated))
  assert (native.returncode,native.stdout,native.stderr)==(truth.returncode,truth.stdout,truth.stderr)
  assert (backend.returncode,backend.stdout,backend.stderr)==(truth.returncode,truth.stdout,truth.stderr)
 for path in ['review/compiler/fx6-valid-programs/p70.a','internal/lower/testdata/callable_producers/p135_object_result.a','internal/lower/testdata/unsupported_view/p37.a']+['internal/lower/testdata/view_diagnostics/oct9_views_'+name+'.a' for name in ['p17','p48','p64']]:
  name=Path(path).stem
  truth=run(name+'-node',node(path)); assert truth.returncode==0
  for backend in ['c','js']:
   refused=run(name+'-'+backend+'-refused',[cli,backend,path])
   assert refused.returncode==1 and ('unsupported member' in refused.stderr or 'copy pair.length' in refused.stderr or 'checked callable producer with a different scalar storage' in refused.stderr)
print('PASS: two additional fixtures agree with Node in both backends; six source controls refuse before emission',flush=True)
