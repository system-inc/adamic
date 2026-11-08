import json
from pathlib import Path
out=Path(__file__).resolve().parent;repo=Path('/workspace/adamic');result=[]
for row in json.loads((out/'source-excerpts.json').read_text()):
 reason=row['reason']
 if reason.startswith('a namespace read before runtime initialization'):
  e=dict(file='internal/lower/namespaces.go',line=269,function='(*lowering).namespaceInitialization',function_line=260,caller='lowerWithElementKey',caller_file='internal/lower/lower.go',caller_line=113,note='The checkCall closure follows graph.reach(target). A reached namespace declaration absent from initialized emits line 269 at the reached read. The direct-read alternative emits the same reason at line 321. lowerWithElementKey calls this walk before refusal and function-body lowering.')
 elif reason.startswith('reading an enum before its runtime initialization'):
  e=dict(file='internal/lower/namespaces.go',line=318,function='(*lowering).namespaceInitialization',function_line=260,caller='lowerWithElementKey',caller_file='internal/lower/lower.go',caller_line=113,note='The visit closure finds a namespace value node whose namespaceRuntimeEnum declaration is not initialized and emits line 318. The reached-call alternative emits the same reason at line 267. lowerWithElementKey calls this walk before refusal and function-body lowering.')
 else:raise RuntimeError('Unmapped new group: '+str(row))
 result.append(dict(where=row['where'],kind=row['kind'],reason=reason,**e))
 for name in {e['file'],e.get('caller_file',e['file'])}:
  p=repo/name;(out/(p.name+'.txt')).write_bytes(p.read_bytes())
(out/'new-stop-emitters.json').write_text(json.dumps(result,indent=2)+'\n');print('mapped_new_groups='+str(len(result)))
