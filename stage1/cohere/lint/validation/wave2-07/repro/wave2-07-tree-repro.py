import pathlib,json,subprocess
root=pathlib.Path.cwd();d=pathlib.Path('/tmp/wave2-07-repro');cohere=root/'cohere/TypeScript/tsc';virtual=cohere/'adamic_wave2_parser_oracle.go';overlay=d/'parser-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(root/'stage1/typescript/parser/testdata/oracle.go')}}));p=d/'adjacent.tsx';p.write_text('<a/><a/>');binary=d/'parser-oracle'
with (d/'parser-build.log').open('w') as f:r=subprocess.run(['go','build','-overlay='+str(overlay),'-o',str(binary),str(virtual)],cwd=cohere,stdout=f,stderr=subprocess.STDOUT)
assert r.returncode==0
results=[]
for name,cmd in [('Go-tree',[str(binary),str(p),'--whole','--jsx-recovery']),('Node-tree',['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(root/'stage1/typescript/parser/main.ts'),str(p),'--whole'])]:
 with (d/(name+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 results.append({'backend':name,'command':cmd,'exit':r.returncode,'output':(d/(name+'.log')).read_text()})
(d/'parser-repro.json').write_text(json.dumps(results,indent=2));print([(r['backend'],r['exit']) for r in results])
