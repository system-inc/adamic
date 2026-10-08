from pathlib import Path
import subprocess
p=Path('internal/lower/representation_clock_string_null.go')
original=p.read_text()
mutants=[('clock-string-null-null-guard',original.replace('if !l.includesNull(proven) {','if false {'), 'present-literals'),('clock-string-null-member-guard',original.replace('if member.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNull) == 0 {','if member.Flags() == 0 {'), 'three-state')]
try:
 for name,source,target in mutants:
  assert source != original
  p.write_text(source)
  with open('/tmp/'+name+'.log','w') as log:
   result=subprocess.run(['go','test','./internal/lower','-run','^TestClockStringNullRepresentation/'+target+'$','-count=1','-v'],stdout=log,stderr=subprocess.STDOUT)
  assert result.returncode != 0, name+' survived'
  log=Path('/tmp/'+name+'.log').read_text()
  assert 'admission for ' in log, name+' failed outside targeted assertion'
  print(name+' killed by '+target)
finally:
 p.write_text(original)
