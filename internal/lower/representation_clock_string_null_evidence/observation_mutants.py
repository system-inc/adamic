from pathlib import Path
import subprocess
p=Path('internal/lower/representation_clock_string_null.go'); original=p.read_text()
mutants=[('clock-string-null-stale-null-observation',original.replace('AlwaysFalse: !nullComparison','AlwaysFalse: nullComparison || !nullComparison')),('clock-string-null-stale-undefined-observation',original.replace('case ir.IsUndefined:\n\t\theld = test.Value','case ir.IsUndefined:\n\t\treturn value'))]
try:
 for name,source in mutants:
  assert source!=original,name+' changed nothing'
  p.write_text(source)
  with open('/tmp/'+name+'.log','w') as log:
   r=subprocess.run(['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode/internal/oracle/testdata/representation_clock_string_null_observations.a$','-count=1','-v'],stdout=log,stderr=subprocess.STDOUT)
  assert r.returncode!=0,name+' survived'
  assert 'stdout differs' in Path('/tmp/'+name+'.log').read_text(),name+' failed outside stdout'
  print(name+' killed by Node stdout')
finally: p.write_text(original)
