import sys,json,time,os
from pathlib import Path
sys.path.insert(0,'/workspace/scratch/elementat-declaration')
from run import run
s=Path('/workspace/scratch/elementat-declaration');binary=s/'bin/adamic'
while not any(json.loads(l)['name']=='build-tools' for l in (s/'commands.jsonl').read_text().splitlines()):time.sleep(1)
assert binary.exists()
results={}
for variant in ['p1','p2']:
 source=s/(variant+'.ts');d=s/variant;d.mkdir(exist_ok=True)
 row={}
 for name,args in [('types',[str(binary),'types',str(source)]),('native-build',[str(binary),'build',str(source),'-o',str(d/'native'),'--sanitize','--explain-checks']),('javascript-build',[str(binary),'js',str(source)]),('node-build',['tsc','--ignoreConfig','--target','es2024','--module','commonjs','--strict','--skipLibCheck','--outDir',str(d),str(source)])]:
  row[name]=run(variant+'-'+name,args,limit=300)
 if row['node-build']['exit']==0:row['node']=run(variant+'-node',['node',str(d/(variant+'.js'))])
 if row['native-build']['exit']==0:row['native']=run(variant+'-native',[str(d/'native')])
 if row['javascript-build']['exit']==0:
  js=d/'adamic.js';js.write_bytes((s/(variant+'-javascript-build.log')).read_bytes());row['javascript']=run(variant+'-javascript',['node','--disable-warning=ExperimentalWarning','/workspace/adamic/oracle/node.mjs',str(js)])
 if all(k in row and row[k]['exit']==0 for k in ['native','javascript','node']):
  row['nativeAgrees']= (s/(variant+'-native.log')).read_bytes()==(s/(variant+'-node.log')).read_bytes()
  row['javascriptAgrees']= (s/(variant+'-javascript.log')).read_bytes()==(s/(variant+'-node.log')).read_bytes()
 results[variant]=row;(s/'probes.json').write_text(json.dumps(results,indent=2)+'\n')

results['nodeVariantsAgree']=(s/'p1-node.log').read_bytes()==(s/'p2-node.log').read_bytes()
lines=(s/'p2-node.log').read_text().splitlines();results['rows']=len(lines)//3;results['expectedValueAndIdentityChecks']=sum(line=='same' for line in lines[2::3])
(s/'probes.json').write_text(json.dumps(results,indent=2)+'\n')
assert results['p1']['types']['exit']==0 and results['p1']['native-build']['exit']==1
assert results['p2']['nativeAgrees'] and results['p2']['javascriptAgrees'] and results['nodeVariantsAgree'] and results['rows']==30 and results['expectedValueAndIdentityChecks']==30
