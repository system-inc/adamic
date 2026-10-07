import json,subprocess
from pathlib import Path
repo=Path('/workspace/adamic');base=Path('/workspace/scratch/branch-shape');original=(repo/'internal/native/branch_shape.go').read_text()
changes=[('final-fraction', 'e.line("\\tif ((double)%s == %s) %s = %s;", integral, value, dispatch, integral)', 'e.line("\\t%s = %s;", dispatch, integral)'),('final-range','e.line("if (%s >= %s && %s <= %s) {", value, cNumber(float64(minimum)), value, cNumber(float64(maximum)))','e.line("if (true) {")')]
for name,old,new in changes:
 assert original.count(old)==1
 mutant=base/(name+'-mutant.go');mutant.write_text(original.replace(old,new));overlay=base/(name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(repo/'internal/native/branch_shape.go'):str(mutant)}}))
 with (base/(name+'-mutant.log')).open('w') as out:
  result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/native','-run','TestNumericSwitchMatchesNode','-count=1'],cwd=repo,stdout=out,stderr=subprocess.STDOUT)
 print(name,'exit',result.returncode,flush=True);assert result.returncode!=0
