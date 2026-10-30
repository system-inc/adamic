#!/usr/bin/env python3
"""Restore only guarded columns; the TTY ruling test must be the sole new failure."""
import argparse,json,subprocess,tempfile
from pathlib import Path
parser=argparse.ArgumentParser();parser.add_argument('--log',type=Path,required=True);args=parser.parse_args()
root=Path(__file__).resolve().parents[3]
source=root/'internal/lower/library_node_process.go';text=source.read_text()
anchor='func (l *lowering) nodeProcessValue(node *ast.Node) (ir.Expression, bool, error) {\n'
restore='''
 if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionToken && l.processPath(node.AsBinaryExpression().Left) == "process.stdout.columns" {
  fallback, err := l.expression(node.AsBinaryExpression().Right)
  if err != nil { return nil,true,err }
  if fallback.Type() != ir.Number { return nil,true,l.notYet(node,"a TTY fallback outside the primitive host ABI") }
  return ir.Coalesce{Value:ir.ProcessCall{Operation:"columns",Of:ir.MaybeNumber},Fallback:fallback,Of:ir.Number},true,nil
 }
'''
assert text.count(anchor)==1;text=text.replace(anchor,anchor+restore)
anchor='\treturn l.processPath(outer) == "process.stdout._handle"'
restore=''' if l.processPath(outer) == "process.stdout.columns" && outer.Parent != nil && outer.Parent.Kind == ast.KindBinaryExpression && outer.Parent.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionToken && outer.Parent.AsBinaryExpression().Left == outer { return true }
'''
assert text.count(anchor)==1;text=text.replace(anchor,restore+anchor)
anchor='func init() {';assert text.count(anchor)==1;text=text.replace(anchor,anchor+'\n RegisterNodeLibraryMembers("node:tty.WriteStream.columns")',1)
with tempfile.TemporaryDirectory(prefix='process-tty-columns-mutant-') as tmp:
 changed=Path(tmp)/source.name;changed.write_text(text)
 overlay=Path(tmp)/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(changed)}}))
 with args.log.open('w') as log:
  result=subprocess.run(['go','test','-json','-overlay',str(overlay),'./internal/lower','-count=1','-timeout=90m','-parallel=2'],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=5460)
 events=[]
 for line in args.log.read_text().splitlines():
  try:events.append(json.loads(line))
  except json.JSONDecodeError:pass
 failures={e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']}
 baseline={'TestProgramRegionCensusMembership','TestProduct_ProgramRegionCensusPairwise','TestProduct_ProgramRegionCensusIndexed','TestCycleFinderCensusEquality','TestProduct_CycleScanningCensus'}
 assert result.returncode==1 and failures-baseline=={'TestNodeProcessTTYRefusals'},(result.returncode,failures)
 print('caught only by TestNodeProcessTTYRefusals; unchanged clean-main census failures:',sorted(failures&baseline))
