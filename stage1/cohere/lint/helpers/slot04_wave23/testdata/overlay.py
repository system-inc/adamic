import json
from pathlib import Path
import sys
root=Path(sys.argv[1]);here=Path(sys.argv[2]);scratch=Path(sys.argv[3]);replacements={}
for name in ['cfg.go','expressions.go','patterns.go','statements.go']:
 original=root/'internal/lint/ecmascript/control_flow_graph'/name
 text=original.read_text()
 seams={
 'cfg.go':{
 'read':'if adamicDependency(b,"read",node){return}',
 'write':'if adamicDependency(b,"write",node){return}',
 'parameter':'if adamicDependency(b,"parameter",node){return}',
 'firstThrowableFork':'if adamicDependency(b,"fork",nil){return}',
 'newBlock':'if adamicMode=="build"{adamicActions=append(adamicActions,fmt.Sprintf("new:%d",len(b.blocks)))}',
 'markFinal':'if adamicMode=="build"{adamicActions=append(adamicActions,fmt.Sprintf("final:%d",blk.Index()))}'},
 'expressions.go':{
 'expr':'if adamicMode!="expr"{adamicDependency(b,"expr",node);return}',
 'binaryExpression':'if adamicDependency(b,"binary",node){return}',
 'conditionalExpression':'if adamicDependency(b,"conditional",node){return}',
 'updateExpression':'if adamicDependency(b,"update",operand){return}',
 'accessOrCall':'if adamicDependency(b,"access",node){return}',
 'classLike':'if adamicDependency(b,"class",node){return}',
 'nestedFunction':'if adamicDependency(b,"nested",node){return}',
 'visitUnknown':'if adamicDependency(b,"unknown",node){return}'},
 'patterns.go':{'bindWithDefault':'adamicActions=append(adamicActions,fmt.Sprintf("default:%d,%d",adamicID(target),adamicID(fallback)));return'},
 'statements.go':{
 'statement':'if adamicDependency(b,"statement",node){return}',
 'statements':'if adamicList(b,list){return}',
 'makeYield':'if adamicDependency(b,"yield",nil){return}'}
 }[name]
 for method,injection in seams.items():
  anchor='func (b *Builder[E]) '+method+'('
  assert text.count(anchor)==1,(name,method)
  index=text.index('{',text.index(anchor))+1
  text=text[:index]+'\n '+injection+'\n'+text[index:]
 if name in ['cfg.go','patterns.go']:text=text.replace('import (','import (\n"fmt"',1)
 target=scratch/name;target.write_text(text);replacements[str(original)]=str(target)
replacements[str(root/'internal/lint/ecmascript/control_flow_graph/adamic_slot04_wave23_exports.go')]=str(here/'exports.go')
replacements[str(root/'adamic_slot04_wave23_oracle.go')]=str(here/'oracle.go')
(scratch/'overlay.json').write_text(json.dumps({'Replace':replacements}))
