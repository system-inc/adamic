"""Prove exit-1 stderr exclusion cannot hide an uncaught sanitizer failure."""
import os
from pathlib import Path
import subprocess
ROOT=Path(__file__).resolve().parents[2]
target=ROOT/'internal/native/runtime/exceptions.c'
original=target.read_bytes()
anchor='adamic_release(adamic_thrown);'
assert original.decode().count(anchor)==1
try:
 target.write_text(original.decode().replace(anchor,anchor+'\n\t(void)adamic_instanceof(adamic_thrown, &adamic_error_class);',1))
 log=ROOT/'docs/step-21-exceptions/evidence/uncaught-sanitizer-visible.log.txt'
 with log.open('w') as output:
  result=subprocess.run(['go','test','./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/step21_object_uncaught[.]a$','-count=1','-v'],cwd=ROOT,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=output,stderr=subprocess.STDOUT,timeout=360)
 observed=log.read_text()
 assert result.returncode!=0 and 'sanitizer failure' in observed and 'AddressSanitizer: heap-use-after-free' in observed,observed
 assert 'clang failed' not in observed
 print('uncaught-sanitizer-visible: exit 1, caught by sanitizer failure despite matching stdout and exit')
finally:
 target.write_bytes(original)
