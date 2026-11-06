"""Prepare Go overlays for existing internal/oracle helpers; runs no programs."""
import json
import sys
from pathlib import Path

root = Path(__file__).resolve().parents[3]
report = Path(__file__).resolve().parent
out = Path(sys.argv[1]).resolve()
out.mkdir(parents=True, exist_ok=True)
for name, source in [('oracle', 'oracle_test.go.txt'), ('signatures', 'signature_trace_test.go.txt'), ('refusal', 'refusal_trace_test.go.txt')]:
    replace = {str(root / 'internal/oracle/combine_gap_report_test.go'): str(report / source)}
    if name == 'refusal':
        path = root / 'internal/lower/invariance.go'
        code = path.read_text().replace('import (', 'import (\n "fmt"\n "os"\n "runtime/debug"', 1)
        needle = 'func (l *lowering) wideningRefusal(node *ast.Node, own, contextual *checker.Type, found *widening) error {'
        trace = '''
 if l.checker.TypeToString(own) == "T[]" && l.checker.TypeToString(contextual) == "{}[]" {
  fmt.Fprintf(os.Stderr,"REFUSAL TRACE %s own=%s contextual=%s mapper-present=%v\\n%s",l.program.Where(node),l.checker.TypeToString(own),l.checker.TypeToString(contextual),l.typeMapper!=nil,debug.Stack())
  if node.Parent.Kind==ast.KindCallExpression {
   sig:=l.checker.GetResolvedSignature(node.Parent)
   if sig!=nil {for i,p:=range sig.Parameters(){fmt.Fprintf(os.Stderr,"resolved parameter %d: %s\\n",i,l.checker.TypeToString(l.checker.GetTypeOfSymbol(p)))}}
  }
 }
'''
        assert code.count(needle) == 1
        patched = out / 'invariance.go.txt'
        patched.write_text(code.replace(needle, needle + trace, 1))
        replace[str(path)] = str(patched)
    (out / (name + '.json')).write_text(json.dumps({'Replace': replace}, indent=2) + '\n')
