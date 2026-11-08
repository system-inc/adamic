import json
from pathlib import Path
out=Path(__file__).resolve().parent;repo=Path('/workspace/adamic');result=[]
for row in json.loads((out/'source-excerpts.json').read_text()):
 reason=row['reason']
 if reason.startswith('the host binding process, pending '):
  e=dict(file='internal/lower/host_bindings.go',line=139,function='hostBindingStop',function_line=124,caller='(*lowering).hostBindingStatement',caller_file='internal/lower/host_bindings.go',caller_line=107,note='hostBindingStop walks row.Targets() at lines 126-129, collecting Pending columns and their owners. It builds the exact reason at line 139. hostBindingStatement (starts at line 96) matches the ambient declarator to the table row, calls hostBindingStop at line 106, and emits l.notYet(declaration, reason) at line 107. ambientRefusal (internal/lower/ambient.go:61) calls hostBindingStatement at line 76; the rule is registered at ordinal 2200, line 94. The embedded process row has native Pending (Runtime), javascript Provided, wasm32 Pending (Runtime and Platforms), and model node:process. This reason is determined by the compiler-embedded table, not by this box missing a Node global.')
  (out/'host-bindings-table.json').write_bytes((repo/'internal/load/host_bindings.json').read_bytes())
  (out/'ambient.go.txt').write_bytes((repo/'internal/lower/ambient.go').read_bytes())
 elif reason.startswith('a namespace read before runtime initialization'):
  e=dict(file='internal/lower/namespaces.go',line=269,function='(*lowering).namespaceInitialization',function_line=260,caller='lowerWithElementKey',caller_file='internal/lower/lower.go',caller_line=113,note='The checkCall closure follows graph.reach(target). A reached namespace declaration absent from initialized emits line 269 at the reached read. The direct-read alternative emits the same reason at line 321. lowerWithElementKey calls this walk before refusal and function-body lowering.')
 elif reason.startswith('reading an enum before its runtime initialization'):
  e=dict(file='internal/lower/namespaces.go',line=318,function='(*lowering).namespaceInitialization',function_line=260,caller='lowerWithElementKey',caller_file='internal/lower/lower.go',caller_line=113,note='The visit closure finds a namespace value node whose namespaceRuntimeEnum declaration is not initialized and emits line 318. The reached-call alternative emits the same reason at line 267. lowerWithElementKey calls this walk before refusal and function-body lowering.')
 else:raise RuntimeError('Unmapped new group: '+str(row))
 result.append(dict(where=row['where'],kind=row['kind'],reason=reason,**e))
 for name in {e['file'],e.get('caller_file',e['file'])}:
  p=repo/name;(out/(p.name+'.txt')).write_bytes(p.read_bytes())
(out/'new-stop-emitters.json').write_text(json.dumps(result,indent=2)+'\n');print('mapped_new_groups='+str(len(result)))
