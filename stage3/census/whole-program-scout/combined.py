"""Combined controls expose the cycle-stop/count-target dependency explicitly."""
import json, pathlib, subprocess, sys
here=pathlib.Path(__file__).resolve().parent
binary=str(pathlib.Path(sys.argv[1]).resolve())
out=pathlib.Path(sys.argv[2]).resolve();out.mkdir(parents=True,exist_ok=True)
for bind in ('stop','clean'):
 for record in ('stop','clean'):
  name='combined-bind-'+bind+'-record-'+record
  subprocess.run([sys.executable,str(here/'measure.py'),str(out/name),binary,str(here/'fixtures'/(name+'.ts'))],check=True)
  r=json.loads((out/(name+'.json')).read_text());w=[s['Words'] for s in r['Stops'] or []]
  assert any('@ts-check' in s for s in w),r
  assert any('a record written into a graph slot' in s for s in w)==(record=='stop'),r
  assert any('bind of a function that declares' in s for s in w)==(bind=='stop' and record=='clean'),r
  assert r['Completed']==(bind=='clean' and record=='clean'),r
  print(name, r['Boundary'] or 'completed')
