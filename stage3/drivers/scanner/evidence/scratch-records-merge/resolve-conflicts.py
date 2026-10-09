from pathlib import Path
import re
root=Path('/workspace/scanner-native3-scratch')
resolutions=[]
for name in ['internal/javascript/javascript.go','internal/lower/expression.go','internal/lower/object.go','internal/lower/refusals.go','internal/lower/statements.go','internal/native/emit_expressions.go']:
 p=root/name
 def resolve(m):
  ours,theirs=m.group(1),m.group(2)
  if name.endswith('javascript.go'): result=ours+theirs
  elif name.endswith('emit_expressions.go'): result=ours+theirs[theirs.index('\tcase ir.RecordCoalesce:'):]
  elif name.endswith('expression.go'): result=ours+'\t}\n'+theirs
  elif name.endswith('object.go'): result=ours+'\t\treturn value, true, err\n\t}\n'+theirs
  elif name.endswith('statements.go'): result=ours+'\t\tif err != nil {\n\t\t\treturn nil, err\n\t\t}\n\t\treturn []ir.Statement{ir.Evaluate{Value: value}}, nil\n\t}\n'+theirs
  elif 'errorStackSyntaxRefusal' in ours:
   result=ours[:ours.index('\t\tif refused,')]+theirs[:theirs.index('\t\tif refused,')].replace('!l.recordTarget(node.AsDeleteExpression().Expression)', '!l.recordTarget(node.AsDeleteExpression().Expression) && !l.nodeProcessEnvironmentDelete(node)')+ours[ours.index('\t\tif refused,'):]
  else: result=ours.replace('!l.methodComparisonUse(node)', '!l.methodComparisonUse(node) && !l.detachedOwnMethod(node)')
  resolutions.append({'file':name,'ours':ours,'theirs':theirs,'resolved':result})
  return result
 p.write_text(re.sub(r'^<<<<<<< HEAD\n(.*?)^=======\n(.*?)^>>>>>>> origin/codex/records-lowering\n',resolve,p.read_text(),flags=re.M|re.S))
import json
Path('/tmp/scanner-records-resolutions.json').write_text(json.dumps(resolutions,indent=2)+'\n')
