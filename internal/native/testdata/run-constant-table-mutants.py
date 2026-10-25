#!/usr/bin/env python3
"""One bounded named witness per template or ownership regression."""
import json, os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[3]
mutants=[
 ('small-function-table','if e.function != nil && len(literal.Elements) < constantNumbers {','if false && e.function != nil && len(literal.Elements) < constantNumbers {','TestConstantNumbersSmallFunctionsInline'),
 ('shared-template-not-copy','adamic_array *array = adamic_array_new(count, false);\n for (size_t index = 0; index < count; index++) {\n  adamic_array_push(array, (adamic_value){.number = values[index]});\n }\n return array;','static adamic_array template; template.heap.kind = adamic_kind_array; template.length = count; template.capacity = count; template.elements = (adamic_value *)(void *)values; return &template;','TestConstantTablesLoopFreshness'),
 ('number-table-identity','e.sharedDeclaration("adamic_numbers", data)','e.sharedDeclaration("adamic_numbers", "constant")','TestConstantTablesSplitSanitized'),
 ('record-table-identity','e.sharedDeclaration("adamic_record_table", data)','e.sharedDeclaration("adamic_record_table", "constant")','TestConstantTablesSplitSanitized'),
 ('truncated-number-copy','index < count','index + 1 < count','TestConstantNumbersFreshAndSignedZero'),
 ('lost-signed-zero','value = -value','value = 0 - value','TestConstantNumbersFreshAndSignedZero'),
 ('runtime-call-as-constant','case ir.NumberConstant:','case ir.Call: return 0, true\n case ir.NumberConstant:','TestConstantEvaluationOrder'),
 ('boolean-layout','"{.boolean = %t}"','"{.number = %t}"','TestConstantTablesSplitSanitized'),
 ('wrong-record-row','object->slots[%d] = values[row][%d];','object->slots[%d] = values[0][%d];','TestConstantTablesSplitSanitized'),
 ('wrong-array-offset','offsets[row * %d + %d]','offsets[0 * %d + %d]','TestConstantTablesSplitSanitized'),
 ('unbounded-constructor','const constantRecordFields = 16','const constantRecordFields = 256','TestConstantRecordConstructorsStayBounded'),
 ('lost-program-placement','literal.Record || literal.ProgramRegion ||','literal.Record ||','TestConstantRecordsKeepPlacement'),
 ('lost-graph-placement','e.adoptGraph(array, "sizeof *"+array, e.graphTypes(literal.GraphTypes) && !literal.ProgramRegion)','// lost adoption','TestConstantNumbersKeepPlacement'),
]
selected=os.environ.get('ADAMIC_CONSTANT_MUTANTS','').split(',')
with tempfile.TemporaryDirectory(prefix='constant-mutants-') as d:
 for name,before,after,test in mutants:
  if selected!=[''] and name not in selected:continue
  file='emit_arrays.go' if name=='number-table-identity' else 'constant_tables.go'
  path=root/'internal/native'/file;source=path.read_text()
  if before not in source:raise RuntimeError('anchor '+name)
  changed=source.replace(before,after,1)
  if name=='boolean-layout':changed=changed.replace('fmt.Sprintf("{.number = %t}", value.Value)','fmt.Sprintf("{.number = %d}", map[bool]int{true:1,false:0}[value.Value])')
  f=Path(d)/(name+'.go.txt');f.write_text(changed);overlay=Path(d)/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(path):str(f)}}))
  log=Path(os.environ.get('ADAMIC_CONSTANT_MUTANT_LOG_DIR','/tmp'))/('constant-mutant-'+name+'.log')
  with log.open('w') as out:r=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/native','-run','^'+test+'$','-count=1','-timeout=55s','-v'],cwd=root,stdout=out,stderr=subprocess.STDOUT,timeout=58)
  text=log.read_text()
  if r.returncode==0 or '--- FAIL: '+test not in text or '[build failed]' in text:raise RuntimeError(name+' not caught: '+text)
  print(name+' caught by '+test,flush=True)
