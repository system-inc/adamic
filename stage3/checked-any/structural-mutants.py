import json,pathlib,subprocess,tempfile
root=pathlib.Path(__file__).resolve().parents[2];scratch=pathlib.Path(tempfile.mkdtemp(prefix='checked-json-mutants-'))
cases=[
 ('prepared-aliases','internal/lower/checked_json.go','if !l.jsonContractPrepared(schema) {','if false && !l.jsonContractPrepared(schema) {','TestCheckedAnyUnsupportedContracts/json_unprepared'),
 ('array-method','internal/lower/checked_json.go','if l.checker.IsArrayType(l.checker.GetTypeAtLocation(receiver)) || l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "ReadonlyArray") {','if false && (l.checker.IsArrayType(l.checker.GetTypeAtLocation(receiver)) || l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "ReadonlyArray")) {','TestCheckedAnyUnsupportedContracts/json_array_method'),
 ('object-fields','internal/native/runtime/checked_json.c','i < contract->count; i++) {\n   const adamic_json_contract_field','i < 0; i++) {\n   const adamic_json_contract_field','TestCheckedJSON/json_object_misfit'),
 ('array-elements','internal/native/runtime/checked_json.c','strcmp(kind, "array") == 0 && (contract->element != NULL || domain)','strcmp(kind, "array") == 0 && (contract->element != NULL || domain) && false','TestCheckedJSON/json_list_misfit'),
 ('literal-value','internal/native/runtime/checked_json.c','if(!matches) {if(terminal)', 'if(!matches && false) {if(terminal)', 'TestCheckedJSON/json_literal_misfit'),
 ('union-alternative','internal/native/runtime/checked_json.c','if (inspect(value, contract->alternatives[i], path, depth, false)) return true;', 'if (true || inspect(value, contract->alternatives[i], path, depth, false)) return true;', 'TestCheckedJSON/json_union_misfit'),
 ('finite-number','internal/native/runtime/checked_json.c','strcmp(kind, "number") == 0 && !isfinite', 'strcmp(kind, "number") == 0 && false && !isfinite','TestCheckedJSON/json_domain_misfit'),
 ('boolean-element','internal/native/runtime/checked_json.c','case 2: return slot->boolean ? &adamic_box_true.heap : &adamic_box_false.heap;', 'case 2: return adamic_box_number(slot->boolean ? 1 : 0);','TestCheckedJSON/json_boolean_list'),
 ('recovery-tags','internal/native/checked_json.go','fmt.Sprintf("(%s)adamic_retain(%s)", cType(to), value)', 'fmt.Sprintf("(%s)adamic_retain(%s == &adamic_null ? NULL : %s)", cType(to), value, value)','TestCheckedJSON/json_recovery'),
 ('alias-write','internal/lower/checked_json.go','if node.Kind == ast.KindBinaryExpression && ast.IsAssignmentOperator', 'if false && node.Kind == ast.KindBinaryExpression && ast.IsAssignmentOperator','TestCheckedAnyUnsupportedContracts/json_alias_write'),
 ('alias-update','internal/lower/checked_json.go','if node.Kind == ast.KindPrefixUnaryExpression || node.Kind == ast.KindPostfixUnaryExpression {', 'if false && (node.Kind == ast.KindPrefixUnaryExpression || node.Kind == ast.KindPostfixUnaryExpression) {','TestCheckedAnyUnsupportedContracts/json_alias_update'),
 ('reflection','internal/lower/checked_json.go','if l.isLibraryGlobal(base, "Object") || l.isLibraryGlobal(base, "Reflect") {','if false && (l.isLibraryGlobal(base, "Object") || l.isLibraryGlobal(base, "Reflect")) {','TestCheckedAnyUnsupportedContracts/json_reflection'),
 ('iteration','internal/lower/checked_json.go','if node.Kind == ast.KindForOfStatement || node.Kind == ast.KindSpreadElement {','if false && (node.Kind == ast.KindForOfStatement || node.Kind == ast.KindSpreadElement) {','TestCheckedAnyUnsupportedContracts/json_iteration'),
 ('callback-tags','internal/lower/library_regexp_replace.go','return 64 | 1 | 2 | 4 | 8','return 64 | 1 | 4 | 8','TestCheckedJSON/stock_decode_entities'),
]
for name,file,before,after,selector in cases:
 source=root/file;text=source.read_text();assert before in text,(name,before)
 mutated=scratch/(name+source.suffix);mutated.write_text(text.replace(before,after,1))
 overlay=scratch/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(source):str(mutated)}}))
 log=scratch/(name+'.log')
 with log.open('wb') as stream:result=subprocess.run(['go','test','-overlay',str(overlay),'./internal/oracle','-run','^'+selector.replace('/','$/')+'$','-count=1','-v'],cwd=root,stdout=stream,stderr=stream)
 output=log.read_text()
 if result.returncode==0 or '[build failed]' in output or 'fatal error:' in output:raise RuntimeError((name,result.returncode,output[-1500:]))
 if 'native ' not in output and 'unsafe contract became accepted' not in output and 'JavaScript ' not in output:raise RuntimeError((name,'unexpected catcher',output[-1000:]))
 print(name,'caught',flush=True)
