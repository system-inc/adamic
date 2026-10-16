#!/usr/bin/env python3
"""One bounded named witness per template or ownership regression."""
import json, os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[3];path=root/'internal/native/constant_tables.go';source=path.read_text()
mutants=[
 ('number-table-identity','e.sharedDeclaration("adamic_number_table", data)','e.sharedDeclaration("adamic_number_table", "constant")','TestConstantTablesSplitSanitized'),
 ('record-table-identity','first.layout+"/"+data+"/"+strings.Join(numbers, ",")+"/"+strings.Join(offsets, ",")','first.layout','TestConstantTablesSplitSanitized'),
 ('truncated-number-copy','index < count','index + 1 < count','TestConstantNumbersFreshAndSignedZero'),
 ('lost-signed-zero','value = -value','value = 0 - value','TestConstantNumbersFreshAndSignedZero'),
 ('runtime-call-as-constant','case ir.NumberConstant:','case ir.Call: return 0, true\n case ir.NumberConstant:','TestConstantEvaluationOrder'),
 ('boolean-layout','"{.boolean = %t}"','"{.number = %t}"','TestConstantTablesSplitSanitized'),
 ('wrong-record-row','object->slots[%d] = values[row][%d];','object->slots[%d] = values[0][%d];','TestConstantTablesSplitSanitized'),
 ('wrong-array-offset','offsets[row * %d + %d]','offsets[0 * %d + %d]','TestConstantTablesSplitSanitized'),
 ('unbounded-constructor','const constantRecordFields = 16','const constantRecordFields = 256','TestConstantRecordConstructorsStayBounded'),
 ('lost-program-placement','literal.Record || literal.ProgramRegion ||','literal.Record ||','TestConstantRecordsKeepPlacement'),
 ('lost-graph-placement','e.adoptGraph(array, "sizeof *"+array, e.graphTypes(literal.GraphTypes))','// lost adoption','TestConstantNumbersKeepPlacement'),
]
selected=os.environ.get('ADAMIC_CONSTANT_MUTANTS','').split(',')
with tempfile.TemporaryDirectory(prefix='constant-mutants-') as d:
 for name,before,after,test in mutants:
  if selected!=[''] and name not in selected:continue
  if before not in source:raise RuntimeError('anchor '+name)
  changed=source.replace(before,after,1)
  if name=='boolean-layout':changed=changed.replace('fmt.Sprintf("{.number = %t}", value.Value)','fmt.Sprintf("{.number = %d}", map[bool]int{true:1,false:0}[value.Value])')
  f=Path(d)/(name+'.go.txt');f.write_text(changed);overlay=Path(d)/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(path):str(f)}}))
  log=Path('/tmp/constant-mutant-'+name+'.log')
  with log.open('w') as out:r=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/native','-run','^'+test+'$','-count=1','-timeout=55s','-v'],cwd=root,stdout=out,stderr=subprocess.STDOUT,timeout=58)
  text=log.read_text()
  if r.returncode==0 or '--- FAIL: '+test not in text or '[build failed]' in text:raise RuntimeError(name+' not caught: '+text)
  print(name+' caught by '+test,flush=True)
