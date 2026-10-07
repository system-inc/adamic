"""Run each proof mutation independently and restore sources even on failure."""
from pathlib import Path
import os, subprocess, json
root = Path(__file__).resolve().parents[5]
borrow = root / 'internal/native/borrow.go'
element = root / 'internal/native/element_borrow.go'
original = {borrow: borrow.read_text(), element: element.read_text()}
mutants = [
 ('field-write', borrow, 'if write, ok := statement.(ir.SetProperty); ok && names[write.Name] {', 'if write, ok := statement.(ir.SetProperty); ok && names[write.Name] && false {', 'oracle', 'write'),
 ('stored-override', borrow, 'if write, ok := statement.(ir.SetProperty); ok && names[write.Name] {', 'if write, ok := statement.(ir.SetProperty); ok && names[write.Name] && false {', 'oracle', 'store'),
 ('root-assignment', element, 'assigned[assign.Local] = true', 'assigned[assign.Local] = false', 'oracle', 'reassigned'),
 ('override', borrow, 'program.CallTargets(expression)', '[]int{expression.Function}', 'oracle', 'override'),
 ('unknown', borrow, 'if targets.Unknown {', 'if targets.Unknown && false {', 'oracle', 'unknown'),
 ('disable-chains', element, 'chain := declaration && borrowableChain(program, index, declare, assigned)', 'chain := false', 'native', ''),
]
results=[]
try:
 for name,path,before,after,package,fixture in mutants:
  for p,s in original.items(): p.write_text(s)
  assert path.read_text().count(before)==1, name
  path.write_text(path.read_text().replace(before,after))
  # Keep the local declaration live when the positive rule is disabled.
  if name == 'disable-chains': element.write_text(element.read_text().replace('chain := false','chain := declaration && false && borrowableChain(program, index, declare, assigned)'))
  pattern = 'TestBorrowChainDeclarations' if package=='native' else 'TestNativeAgreesWithNode/internal/oracle/testdata/borrow_chain_'+fixture+r'\.a$'
  log=Path('/tmp/borrow-chains-mutant-'+name+'.log')
  with log.open('wb') as output:
   result=subprocess.run(['go','test','./internal/'+package,'-run',pattern,'-count=1','-timeout','30m'],cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'},stdout=output,stderr=subprocess.STDOUT)
  text=log.read_text()
  killed=result.returncode != 0 and ('heap-use-after-free' in text if package=='oracle' else 'did not borrow' in text)
  results.append({'mutant':name,'exit':result.returncode,'killed':killed,'log':str(log)})
finally:
 for p,s in original.items(): p.write_text(s)
Path('/tmp/borrow-chains-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))
assert all(result['killed'] for result in results)
