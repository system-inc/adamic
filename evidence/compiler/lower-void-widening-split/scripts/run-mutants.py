from pathlib import Path
import subprocess,json
out=Path('/tmp/void-split-evidence'); root=Path('/workspace/adamic')
results={}
for name,fixture in [('union','v04_widen_union.a'),('structural','container_array.a')]:
 binary='/tmp/adamic-void-split-'+name
 path='internal/oracle/testdata/void_widening_refused/'+fixture
 source=out/(name+'-source.ts'); source.write_bytes((root/path).read_bytes())
 native=out/(name+'-native'); js=out/(name+'-backend.mjs')
 commands=[('native-build',[binary,'build',path,'-o',str(native),'--sanitize']),('js-build',[binary,'js',path]),('node',['node','--experimental-transform-types',str(source)])]
 exits={}
 for label,cmd in commands:
  p=subprocess.run(cmd,cwd=root,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
  (out/(name+'-'+label+'.stdout')).write_bytes(p.stdout); (out/(name+'-'+label+'.stderr')).write_bytes(p.stderr); exits[label]=p.returncode
  if label=='js-build' and p.returncode==0:js.write_bytes(p.stdout)
 if exits['native-build']==0:
  p=subprocess.run([str(native)],cwd=root,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
  exits['native']=p.returncode;(out/(name+'-native.stdout')).write_bytes(p.stdout);(out/(name+'-native.stderr')).write_bytes(p.stderr)
 if exits['js-build']==0:
  p=subprocess.run(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(js)],cwd=root,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
  exits['js']=p.returncode;(out/(name+'-js.stdout')).write_bytes(p.stdout);(out/(name+'-js.stderr')).write_bytes(p.stderr)
 results[name]=exits
(out/'mutant-runs.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results))
