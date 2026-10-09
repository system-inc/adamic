from pathlib import Path
import gzip, os, subprocess, sys
root=Path(__file__).resolve().parents[2]
logs=root/'docs/step-21-exceptions/evidence'
out=root/'docs/step-21-exceptions/landing-evidence'
out.mkdir(exist_ok=True)
originals={p:p.read_bytes() for p in logs.glob('*.log.txt')}
runners=['run-mutants.py','run-adopted-mutants.py','run-library-mutants.py','run-subclass-mutants.py','run-saved-mutants.py','prove-uncaught-sanitizer.py','run-main-mutants.py']
try:
 for runner in runners:
  with (out/(runner+'.log')).open('w') as output:
   result=subprocess.run(['python3',str(root/'docs/step-21-exceptions'/runner)],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  print(runner, 'exit',result.returncode,flush=True)
  print((out/(runner+'.log')).read_text(),flush=True)
  for p in logs.glob('*.log.txt'):
   if p not in originals or p.read_bytes()!=originals[p]:
    (out/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
  if result.returncode: sys.exit(result.returncode)
finally:
 for p in logs.glob('*.log.txt'):
  if p in originals: p.write_bytes(originals[p])
  else: p.unlink()
p=root/'docs/step-21-exceptions/census.py'
original=p.read_bytes()
try:
 for name,old,new in [('census-mutant-selector',"return reason.startswith(PREFIXES)","return reason.startswith(PREFIXES) or reason == 'reading exception'"),('census-mutant-accounting',"'diagnostics': len(catch_unknown)","'diagnostics': len(catch_unknown)-1")]:
  source=original.decode(); assert source.count(old)==1
  p.write_text(source.replace(old,new,1))
  with (out/(name+'.log')).open('w') as output:
   result=subprocess.run(['python3',str(p)],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  observed=(out/(name+'.log')).read_text()
  assert result.returncode!=0 and 'AssertionError' in observed,(name,result.returncode,observed)
  print(name,'exit',result.returncode,'caught by census assertion',flush=True)
finally: p.write_bytes(original)
