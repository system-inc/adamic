"""Mutate real projection guards; valid Go binaries must fail semantic witnesses."""
from pathlib import Path
import os,subprocess
p=Path('internal/lower/shape_flow.go');source=p.read_text()
mutants={
 'projection-drop-stores':source.replace('sources = append(sources, index.stores[site][name]...)',''),
 'projection-ignore-opaque-store':source.replace('if index.opaque[name] {','if false {'),
 'projection-ignore-missing-field':source.replace('if !found {','if !found && false {'),
 'projection-ignore-array-mutation':source.replace('if array && index.arrayMutation {','if false {'),
}
try:
 for name,changed in mutants.items():
  assert changed!=source,name
  p.write_text(changed)
  with Path('stage3/shape-conformance/logs/'+name+'.log').open('w') as log:
   status=subprocess.run(['go','test','./internal/lower','-run','^TestShapeProjection','-count=1','-v'],stdout=log,stderr=subprocess.STDOUT).returncode
  text=Path('stage3/shape-conformance/logs/'+name+'.log').read_text()
  assert status!=0 and 'build failed' not in text,name+' escaped or failed to build'
  print(name,'caught by semantic query assertion in valid Go test binary',flush=True)
finally:p.write_text(source)
