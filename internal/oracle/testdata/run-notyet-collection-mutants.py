#!/usr/bin/env python3
"""Independent scratch-overlay mutants for intrinsic lengths and collection inputs."""
from pathlib import Path
import json,os,subprocess
root=Path(__file__).resolve().parents[3]
logs=Path('/tmp/destructuring-collection-mutants');logs.mkdir(exist_ok=True)
snapshot='''function := len(l.result.Functions)
 parameter := len(l.result.Locals)
 l.result.Locals = append(l.result.Locals, ir.Local{Name: "early_value", Type: of, Function: function}, ir.Local{Name: "late_receiver", Type: ir.Object, Function: function})
 l.result.Functions = append(l.result.Functions, ir.Function{Name: "wrong_tuple_write", Parameters: []int{parameter, parameter+1}, Body: []ir.Statement{ir.SetProperty{Object: ir.Read{Local: parameter+1, Of: ir.Object}, Name: strconv.Itoa(offset), Value: ir.Read{Local: parameter, Of: of}}}})
 return ir.Evaluate{Value: ir.Call{Function: function, Arguments: []ir.Expression{value, receiver}}}, nil'''
mutants=[
 ('array-length','internal/lower/destructuring_lengths.go','value := ir.Expression(ir.Length{Array: read})','value := ir.Expression(ir.NumberConstant{Value: 99})','notyet_destructuring_lengths'),
 ('utf16-length','internal/lower/destructuring_lengths.go','value = ir.StringLength{Value: read}','value = ir.Length{Array: read}','notyet_destructuring_lengths'),
 ('map-interleaving','internal/lower/collections.go','return l.mapFromIteration(node, source, plan, key, value)','pairs, err := l.collectIteration(node, source, nil, plan, ir.Object)\n return ir.MapNew{Key: key, Value: value, Pairs: pairs}, err','notyet_map_iterable_pairs'),
 ('map-next-close','internal/lower/destructuring_map_pairs.go','step, entry := l.iterationStep(state)','step, entry := l.iterationStep(state)\n step[0] = ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: true}}','notyet_map_iterable_pairs'),
 ('absent-array','internal/lower/collections.go','Fallback: ir.ArrayLiteral{Element: element}, Of: ir.Array','Fallback: ir.ArrayLiteral{Element: element, Elements: []ir.Expression{ir.StringConstant{Index: l.constant("wrong")}}}, Of: ir.Array','notyet_iterated_optional_arrays'),
 ('absent-string','internal/lower/collections.go','Fallback: ir.StringConstant{Index: l.constant("")}, Of: ir.String','Fallback: ir.StringConstant{Index: l.constant("wrong")}, Of: ir.String','notyet_iterated_optional_arrays'),
 ('tuple-target-index','internal/lower/destructuring_tuple_targets.go','Name: strconv.Itoa(offset), Value: value','Name: "0", Value: value','notyet_tuple_member_assignment'),
 ('tuple-target-before-read','internal/lower/destructuring_tuple_targets.go','return ir.SetProperty{Object: receiver, Name: strconv.Itoa(offset), Value: value, Site: l.writeSite(access.Expression)}, nil',snapshot,'notyet_tuple_member_assignment'),
]
for name,file,before,after,fixture in mutants:
 original=(root/file).read_text();assert original.count(before)==1,(name,original.count(before))
 source=logs/(name+'.go');source.write_text(original.replace(before,after,1))
 overlay=logs/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(source)}}))
 with (logs/(name+'.log')).open('w') as log:
  result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/oracle','-run','^TestNativeAgreesWithNode/internal/oracle/testdata/'+fixture+'.a$','-count=1','-timeout','10m'],cwd=root,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=log,stderr=subprocess.STDOUT)
 output=(logs/(name+'.log')).read_text()
 assert result.returncode!=0 and 'stdout differs' in output and '[build failed]' not in output and 'clang failed' not in output,(name,result.returncode,output)
 print(name+': caught by stdout differs',flush=True)
