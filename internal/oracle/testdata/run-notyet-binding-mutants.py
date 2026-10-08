#!/usr/bin/env python3
"""Compile scratch overlays, leaving production sources unchanged."""
from pathlib import Path
import json,os,subprocess
root=Path(__file__).resolve().parents[3]
logs=Path('/tmp/destructuring-binding-mutants'); logs.mkdir(exist_ok=True)
helper='internal/lower/destructuring_bindings.go'
mutants=[
 ('evaluate-once','internal/lower/collections.go','return append([]ir.Statement{ir.Declare{Local: held, Value: value}}, declared...), nil','return append([]ir.Statement{ir.Evaluate{Value: value}, ir.Declare{Local: held, Value: value}}, declared...), nil','arrays'),
 ('holes',helper,'Index: position, Element: element','Index: ir.NumberConstant{Value: 0}, Element: element','arrays'),
 ('rest-copy',helper,'WhenNot: ir.ArraySlice{Array: array, Arguments: []ir.Expression{position}}','WhenNot: array','arrays'),
 ('sticky-exhaustion',helper,'Left: ir.Read{Local: done, Of: ir.Boolean}, Right: ir.Binary{Operator: ir.GreaterOrEqual','Left: ir.BooleanConstant{Value: false}, Right: ir.Binary{Operator: ir.GreaterOrEqual','defaults'),
 ('lazy-undefined-default',helper,'test := ir.Expression(ir.IsUndefined{Value: read})','test := ir.Expression(ir.BooleanConstant{Value: true})','defaults'),
 ('nested-field','internal/lower/collections.go','value := ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: field, Of: of, Absent: absent}','if field == "text" { field = "wrong" }\n value := ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: field, Of: of, Absent: absent}','nested'),
 ('object-union-box','internal/lower/object.go','value = fit(value, declared)','value = fit(value, declared)\n if declared == ir.Union { value = fit(ir.Undefined{}, declared) }','slots'),
 ('tuple-union-box','internal/lower/object.go','if value = fit(value, of); value.Type() != of {','if of == ir.Union { value = ir.Undefined{} }\n if value = fit(value, of); value.Type() != of {','slots'),
 ('tuple-field','internal/lower/assignments.go','return value, nil\n}\n\n// destructuringAssignment','if of == ir.Union { value = fit(ir.Undefined{}, to) }\n return value, nil\n}\n\n// destructuringAssignment','slots'),
]
for name,file,before,after,fixture in mutants:
 original=(root/file).read_text(); assert original.count(before)==1,(name,original.count(before))
 source=logs/(name+'.go');source.write_text(original.replace(before,after,1))
 overlay=logs/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(source)}}))
 with (logs/(name+'.log')).open('w') as log:
  result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/oracle','-run','^TestNativeAgreesWithNode/internal/oracle/testdata/notyet_destructuring_'+fixture+'.a$','-count=1','-timeout','10m'],cwd=root,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=log,stderr=subprocess.STDOUT)
 output=(logs/(name+'.log')).read_text()
 assert result.returncode!=0 and 'stdout differs' in output and '[build failed]' not in output and 'clang failed' not in output,(name,result.returncode,output)
 print(name+': caught by stdout differs',flush=True)
