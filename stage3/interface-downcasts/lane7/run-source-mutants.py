#!/usr/bin/env python3
"""Each IR mutant must fail the independent source refusal test, in both backends."""
import os,subprocess,sys
from pathlib import Path
root=Path(__file__).resolve().parents[3]
logs=Path(sys.argv[1]).resolve();logs.mkdir(parents=True,exist_ok=True)
for kind,fixture in [('skip','emit-root-wrong'),('shape','emit-root-wrong'),('nested','emit-root-nested')]:
 env=dict(os.environ,ADAMIC_INTERSECTION_HOOKS='1',ADAMIC_INTERSECTION_SOURCE_MUTANT=kind)
 cmd=['go','test','./internal/oracle','-run','^TestCheckedViewIntersectionSource/'+fixture+'$','-count=1','-v']
 path=logs/(kind+'.log')
 with path.open('w') as log: result=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 text=path.read_text()
 if result.returncode==0 or text.count('got oracle.run') != 3:
  raise SystemExit('mutant did not fail the refusal pin: '+str(path))
 print(kind+': pinned source assertion failed as required; '+str(path))
