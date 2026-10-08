from pathlib import Path
import json,subprocess
root=Path(__file__).resolve().parents[3]; original=root/'internal/flow/build.go'
s=original.read_text(); target='\t\t\tthrows = throws || ir.NumberFormatMayThrow(value.Interface())\n';assert s.count(target)==1
path=Path('/tmp/math-number-flow-mutant.go');path.write_text(s.replace(target,''))
overlay=Path('/tmp/math-number-flow-overlay.json');overlay.write_text(json.dumps({'Replace':{str(original):str(path)}}))
with open('/tmp/math-number-flow-mutant.log','w') as log:
 p=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/flow','-run','TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/library_math_number_prototype','-count=1','-v','-timeout=10m'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
s=Path('/tmp/math-number-flow-mutant.log').read_text()
assert p.returncode==1 and 'no edge from bb' in s,s
print('Missing formatting throw edges caught only by the Node trace; exit 1')
