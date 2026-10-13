import json,re,subprocess,time
from pathlib import Path
root=Path.cwd();evidence=root/'review/compiler/loom-v4'
files=subprocess.check_output(['git','diff','--name-only','HEAD','--','*_test.go'],text=True).splitlines()
packages={}
for file in files:
 for name in re.findall(r'^func (Test\w+)\(t \*testing.T\)',Path(file).read_text(),re.M):
  packages.setdefault(str(Path(file).parent),set()).add(name)
packages.setdefault('internal/ir',set()).update(['TestCallTargetReaders','TestProduct_CallTargetExports'])
packages.setdefault('internal/oracle',set()).update(['TestProduct_Adamic']+['TestProduct_NativeRuntimes'+kind+str(i) for kind in ['Sanitized','Release','Counted'] for i in range(3)])
(evidence/'selected-tests.json').write_text(json.dumps({k:sorted(v) for k,v in packages.items()},indent=2)+'\n')
with (evidence/'results.jsonl').open('w') as results:
 for package,names in packages.items():
  binary=evidence/(package.replace('/','-')+'.test')
  with (evidence/(package.replace('/','-')+'-build.log')).open('w') as log:
   start=time.monotonic();p=subprocess.run(['go','test','-c','-o',str(binary),'./'+package],stdout=log,stderr=subprocess.STDOUT,timeout=300)
  results.write(json.dumps(dict(package=package,build=True,seconds=round(time.monotonic()-start,3),exit=p.returncode))+'\n');results.flush()
  if p.returncode:continue
  for name in sorted(names, key=lambda name: (not name.startswith("TestProduct_"), name)):
   with (evidence/(package.replace('/','-')+'-'+name+'.log')).open('w') as log:
    start=time.monotonic()
    try:p=subprocess.run([str(binary),'-test.run','^'+name+'$','-test.count=1','-test.timeout=90s','-test.v','-test.parallel=4'],cwd=root/package,stdout=log,stderr=subprocess.STDOUT,timeout=95);code=p.returncode
    except subprocess.TimeoutExpired:code=124
   seconds=round(time.monotonic()-start,3)
   results.write(json.dumps(dict(package=package,test=name,seconds=seconds,exit=code,grain=seconds<60))+'\n');results.flush()
   print(name,code,seconds,flush=True)
  binary.unlink()
