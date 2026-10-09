# Run from repository root at the audit's starting commit plus evidence files.
import pathlib,subprocess
r=pathlib.Path('.');p=r/'review/test-audit/internal-oracle-checked_views_v2_source';bridge=r/'internal/lower/u058_bridge.go';main=p/'bridge-probe/main.go'
assert not bridge.exists() and not main.exists()
bridge.write_bytes((p/'bridge-probe/bridge.go.txt').read_bytes());main.write_bytes((p/'bridge-probe/main.go.txt').read_bytes())
try:
 for id,mode in [('M2','supported'),('M3','complete'),('M15','lazy')]:
  subprocess.run(['git','apply','--check',str(p/'diffs'/(id+'.diff'))],check=True)
  for variant in ['before','after']:
   if variant=='after':subprocess.run(['git','apply',str(p/'diffs'/(id+'.diff'))],check=True)
   try:
    with (p/f'survivor-{id}-direct-{variant}.log').open('w') as log:subprocess.run(['go','run','./'+str(p/'bridge-probe'),mode],stdout=log,stderr=subprocess.STDOUT,check=True)
   finally:
    if variant=='after':subprocess.run(['git','apply','-R',str(p/'diffs'/(id+'.diff'))],check=True)
finally:bridge.unlink();main.unlink()
