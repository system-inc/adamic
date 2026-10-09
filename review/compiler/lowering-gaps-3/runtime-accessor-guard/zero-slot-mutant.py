from pathlib import Path
import subprocess,json,difflib
out=Path('review/compiler/lowering-gaps-3/runtime-accessor-guard');p=Path('internal/native/runtime/object.c');saved=p.read_bytes()
try:
 s=saved.decode();needle='\tmemset(object->slots, 0, shape->count * sizeof object->slots[0]);\n';assert needle in s;s=s.replace(needle,'',1);p.write_text(s)
 (out/'main-accessor-slots-revert.patch').write_text(''.join(difflib.unified_diff(saved.decode().splitlines(True),s.splitlines(True),fromfile=str(p),tofile=str(p))))
 with (out/'main-accessor-slots-revert.log').open('wb') as log:r=subprocess.run(['timeout','150','go','test','./internal/oracle','-run','^TestNativeAgreesWithNode$/internal/oracle/testdata/accessor_spread_throw_(first|middle|last)\\.a$','-count=1','-timeout=90s','-v'],stdout=log,stderr=subprocess.STDOUT)
 text=(out/'main-accessor-slots-revert.log').read_text();caught=r.returncode==1 and ('AddressSanitizer' in text or 'UndefinedBehaviorSanitizer' in text)
 (out/'zero-slot-mutant.json').write_text(json.dumps(dict(exit=r.returncode,caught=caught))+'\n');assert caught
finally:p.write_bytes(saved)
