import pathlib,shutil,subprocess,sys,json
s=pathlib.Path('/workspace/overload-visitor-scratch');repo=s/'adamic'
for name in ['visitNode','visitNodes','visitArray','visitArrayWorker','visitParameterList','visitEachChild']:
 p=s/name;p.mkdir()
 shutil.copytree(s/'baseline/src',p/'src')
 for f in ['package.json','tsconfig.json','tsconfig.base.json']:
  if (s/'baseline'/f).exists():shutil.copy2(s/'baseline'/f,p/f)
 with open(repo/'stage3/census/overload-visitor'/(name+'-trial.json'),'w') as out:
  r=subprocess.run(['node',str(repo/'stage3/census/overload-visitor/trial.cjs'),str(p),name],stdout=out)
 if r.returncode:sys.exit(r.returncode)
 print('prepared',name,flush=True)
