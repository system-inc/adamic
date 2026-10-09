#!/usr/bin/env python3
"""Focused structural-result mutants; production sources remain unchanged."""
import json,os,subprocess,tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[4]
out=Path(__file__).parent
structural=root/'internal/lower/overload_structural.go'
proof=root/'internal/lower/census_overload_proof.go'
relation=root/'internal/lower/census_small.go'
original=structural.read_text()
mutants=[
 ('drop-proof',structural,original.replace('return l.censusProveOverloadResult(implementation, overload)','return false'),'TestOverloadStructuralResults',''),
 ('trust-shape',proof,proof.read_text().replace('parameters := implementation.Parameters()', 'if promised.Flags()&checker.TypeFlagsObject != 0 { return true }; parameters := implementation.Parameters()'),'TestOverloadStructuralRefuses/missing_Block_field','got <nil>'),
 ('reverse-covariance',proof,proof.read_text().replace('return l.censusRelated(narrowed(index, flow), promised)','return l.checker.IsTypeAssignableTo(promised, narrowed(index, flow))'),'TestOverloadStructuralRefuses/reverse_readonly_variance','got <nil>'),
 ('mutable-variance',relation,relation.read_text().replace('return l.classAssignable(from, to) && l.widened(from, to, map[[2]*checker.Type]bool{}) == nil','return l.classAssignable(from, to)'),'TestOverloadStructuralRefuses/writable_variance','got <nil>'),
]
rows=[]
for name,source,changed,selector,marker in mutants:
 assert changed != source.read_text()
 with tempfile.TemporaryDirectory(prefix='structural-mutant-') as tmp:
  tmp=Path(tmp);replacement=tmp/source.name;replacement.write_text(changed)
  overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
  cmd=['go','test','-overlay',str(overlay),'./internal/lower','-run','^'+selector.split('/')[0]+'$'+('/'+selector.split('/')[1]+'$' if '/' in selector else ''),'-count=1','-v']
  path=out/(name+'.log.txt')
  with path.open('w') as log: result=subprocess.run(cmd,cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
  text=path.read_text();caught=result.returncode!=0 and '--- FAIL:' in text and '[build failed]' not in text and marker in text
  rows.append(dict(mutant=name,exit=result.returncode,caught=caught));print(name,'caught' if caught else 'NOT CAUGHT',flush=True)
(out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(row['caught'] for row in rows)
