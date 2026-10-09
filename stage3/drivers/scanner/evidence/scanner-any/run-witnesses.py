from pathlib import Path
import subprocess,os,json
root=Path('/workspace/scratch/scanner-any-witnesses');ts='/workspace/scratch/native3-cache/api/node_modules/typescript/lib/typescript.js';results=[]
for site in ['private','map','driver']:
 p=root/(site+'.a');s=p.read_text();js='const fs=require("fs"),ts=require(process.argv[1]);console.log(ts.transpileModule(fs.readFileSync(process.argv[2],"utf8"),{compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.CommonJS}}).outputText)'
 text=subprocess.check_output(['node','-e',js,ts,str(p)],text=True);node=root/(site+'.cjs');node.write_text(text)
 for mutant in [False,True]:
  if mutant:p.write_text(s.replace('string | number | undefined','any').replace('Map<string, Keyword>','Map<string, any>'))
  text=subprocess.check_output(['node','-e',js,ts,str(p)],text=True);node.write_text(text)
  name=site+('-restored-any' if mutant else '-concrete')
  with (root/(name+'.node.stdout')).open('wb') as out,(root/(name+'.node.stderr')).open('wb') as err:n=subprocess.run(['node',str(node)],stdout=out,stderr=err).returncode
  with (root/(name+'.build.stdout')).open('wb') as out,(root/(name+'.build.stderr')).open('wb') as err:b=subprocess.run(['/workspace/scratch/scanner-any-next-adamic','build',str(p),'-o',str(root/(name+'-native'))],cwd='/workspace/scanner-native3-next',stdout=out,stderr=err).returncode
  r={'site':site,'mutant':mutant,'node_exit':n,'native_build_exit':b,'diagnostic':(root/(name+'.build.stderr')).read_text()};results.append(r)
  if b==0:
   with (root/(name+'.native.stdout')).open('wb') as out,(root/(name+'.native.stderr')).open('wb') as err:r['native_exit']=subprocess.run([str(root/(name+'-native'))],stdout=out,stderr=err).returncode
   with (root/(name+'.diff')).open('wb') as out:r['diff_exit']=subprocess.run(['diff','-u',str(root/(name+'.node.stdout')),str(root/(name+'.native.stdout'))],stdout=out,stderr=subprocess.STDOUT).returncode
  if b==0:
   data=(root/(name+'.native.stdout')).read_bytes();assert data;changed=bytes([data[0]^1])+data[1:];(root/(name+'.byte-mutant.stdout')).write_bytes(changed)
   with (root/(name+'.byte-mutant.diff')).open('wb') as out:r['one_byte_mutant_diff_exit']=subprocess.run(['diff','-u',str(root/(name+'.node.stdout')),str(root/(name+'.byte-mutant.stdout'))],stdout=out,stderr=subprocess.STDOUT).returncode
   assert r['native_exit']==0 and r['diff_exit']==0 and r['one_byte_mutant_diff_exit']==1
  assert n==0 and (not mutant or b==1 and 'any' in r['diagnostic'])
  print(json.dumps(r),flush=True);p.write_text(s)
(root/'results.json').write_text(json.dumps(results,indent=2)+'\n')
