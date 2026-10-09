from run import run,E
from pathlib import Path
selector=Path('/tmp/adamic-u083-mutant')
assert run('switch-control')['exit']==0, 'red instrumentation control'
for mid in ['M1','M2','M3','M4','M5','M6','M7','M8','M9','E1','E2']:
 selector.write_text(mid if mid in ['M1','M2','M3','M4','M5','M6','E1'] else '')
 env={'ADAMIC_U083_MUTANT':mid if mid in ['M7','M8','M9','E2'] else '', 'ADAMIC_U083_WITNESS':''}
 if mid in ['M7','M8','M9','E2']:
  # native.go enables split compilation only for literal 1; the existing cache
  # keys retain this value, forcing compiler-output products to be rebuilt.
  env['ADAMIC_NATIVE_SPLIT']='u083-'+mid
 run(mid,environment=env)
 if mid=='E2':run('E2-gap','^TestMultiPushGap$',environment=env)
selector.write_text('')
for wid in ['W1','W2']:run(wid,environment={'ADAMIC_U083_MUTANT':'','ADAMIC_U083_WITNESS':wid})
run('switch-final-control',environment={'ADAMIC_U083_MUTANT':'','ADAMIC_U083_WITNESS':''})
