# Run from the repository root after sourcing the toolchain environment.
from pathlib import Path
import os,subprocess
path=Path('internal/lower/shape_conformance.go'); original=path.read_text()
mutants=[
 ('ignore-nonconforming-shape',original.replace(' || program.ViewContractTypes[certificate.CheckerType] != contract',''),'^TestCheckedViewShapeErasure/nonconforming$'),
 ('ignore-readiness',original.replace('field.Uninitialized || certificate == nil','certificate == nil').replace('value.View != "" && value.Readiness == "" &&','value.View != "" &&').replace('value.View = ""','value.Readiness = ""\n\t\t\t\tvalue.View = ""'),'^TestCheckedViewShapeErasure/uninitialized$'),
]
try:
 for name,changed,selection in mutants:
  assert changed!=original
  path.write_text(changed)
  with open('stage3/shape-conformance/logs/'+name+'.log','w') as log:
   result=subprocess.run(['go','test','./internal/oracle','-run',selection,'-count=1','-v'],stdout=log,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'})
  print(name,'exit',result.returncode,flush=True)
  if result.returncode==0: raise SystemExit('mutant escaped')
finally:
 path.write_text(original)
