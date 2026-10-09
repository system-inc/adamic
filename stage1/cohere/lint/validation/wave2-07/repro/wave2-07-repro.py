import pathlib,subprocess,json,shutil
root=pathlib.Path.cwd();d=pathlib.Path('/tmp/wave2-07-repro');d.mkdir(exist_ok=True)
oracle=d/'oracle';shutil.copy2('/tmp/adamic-gate/lint-shared-2257740487/oracle-3141954837/oracle',oracle)
rows=[]
for i,s in enumerate(['<A a /><B a />;','<A/><B/>;','<A /><B/>','<A a/><B/>','<A a/><B a/>','<a/><a/>','<A/><B/>']):
 p=d/f'case-{i}.tsx';p.write_text(s);rows.append(str(p)+'\t@typescript-eslint/no-unused-expressions\t\t\t\t\trecovery')
manifest=d/'manifest';manifest.write_text('\n'.join(rows)+'\n')
for name,cmd in [('Go',[str(oracle),'--manifest',str(manifest)]),('Node',['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(root/'stage1/cohere/lint/main.ts'),'--manifest',str(manifest)])]:
 with (d/(name+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 print(name,r.returncode,(d/(name+'.log')).read_text())
