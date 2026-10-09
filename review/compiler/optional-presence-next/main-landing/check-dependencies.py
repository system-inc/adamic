from pathlib import Path
import sys,types,json,subprocess
root=Path('/workspace/adamic');tools=Path('/workspace/scratch/landing-gate-tools');sys.path.insert(0,str(tools/'cloud/fast-gate'));import run
candidate=json.loads((root/'cloud/fast-gate/compiler-dependencies.json').read_text());parents=[json.loads(subprocess.check_output(['git','show',rev+':cloud/fast-gate/compiler-dependencies.json'],cwd=root,text=True,timeout=30)) for rev in ['0292780a','b8bcadb2']]
for parent in parents:
 for package,deps in parent['packages'].items():assert set(deps)<=set(candidate['packages'][package]),package
paths=run.git(str(root),'ls-files').splitlines();directories={str(Path(p).parent) for p in paths if p.endswith('.go')};changed=run.git(str(root),'diff','--name-only','origin/main...HEAD').splitlines();gate=run.Gate.__new__(run.Gate);gate.arguments=types.SimpleNamespace(tree=str(root),tools=str(tools));gate.result={};gate.git=lambda tree,*args:run.git(tree,*args)
declarations=gate.compilerDependencies(directories,changed);missing=[]
for package in candidate['packages']:assert declarations[package]==candidate['packages'][package]
print('Actual gate dependency reader accepted both parent maps; new/changed undeclared consumers:',missing)
print(json.dumps({'candidate':candidate,'census':gate.result},indent=2))
