#!/usr/bin/env python3
import subprocess
from pathlib import Path
owned=Path(__file__).resolve().parent
repo=owned.parents[5]
scratch=Path('/tmp/wave15-design-system-cache');scratch.mkdir(exist_ok=True)
log=scratch/'mutex-build.log'
with log.open('wb') as out:
 result=subprocess.run(['go','run','./cmd/adamic','build',str(owned/'gaps/mutex.a'),'-o',str(scratch/'mutex')],cwd=repo,stdout=out,stderr=subprocess.STDOUT)
text=log.read_text()
assert result.returncode!=0,'Mutex support changed; reevaluate the blocker'
assert 'TS2305' in text and "no exported member 'Mutex'" in text,text
assert not (scratch/'mutex').exists(),'unexpected native artifact'
print('Observed: Mutex probe exits '+str(result.returncode)+' with TS2305; no native artifact emitted.')
print('This is a missing-prerequisite probe, not a semantic mutant or a completed helper.')
