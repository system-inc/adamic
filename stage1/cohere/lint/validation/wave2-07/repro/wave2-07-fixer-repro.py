import pathlib,json,subprocess
root=pathlib.Path.cwd();d=pathlib.Path('/tmp/wave2-07-repro');rows=[]
sources=['const a=1;\u00a0const b=2;','let a;\u00a0let b','var a;\u00a0var b;','let a;\u2002let b;','let a;\ufefflet b;', '\ufeffconst é = 1;\u00a0\u2002\u2003\nconst x =\u000b\u000b1;\u2028\u2029']
for i,s in enumerate(sources):
 p=d/f'fixer-{i}.ts';p.write_text(s);rows.append(str(p)+'\tone-var');rows.append(str(p)+'\tall')
manifest=d/'fixer-manifest';manifest.write_text('\n'.join(rows)+'\n');results=[]
for name,cmd in [('Go',[str(d/'oracle'),'--manifest',str(manifest)]),('Node',['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(root/'stage1/cohere/lint/main.ts'),'--manifest',str(manifest)])]:
 with (d/('fixer-'+name+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 results.append({'backend':name,'exit':r.returncode,'command':cmd})
(d/'fixer-repro.json').write_text(json.dumps({'sources':sources,'results':results},indent=2));print(results)
