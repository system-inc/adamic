import pathlib,json,hashlib
root=pathlib.Path('/workspace/scratch/namedeval-kind-S1');runtime=pathlib.Path('/workspace/scratch/namedeval-kind-runtime/S1');out=runtime/'call-trace';out.mkdir(exist_ok=True)
base=json.loads((runtime/'latent-overlay/overlay.json').read_text())
file=root/'internal/lower/overload_result_check.go'
source=pathlib.Path(base['Replace'].get(str(file),str(file))).read_text()
anchor='func (l *lowering) overloadResultCheck(node *ast.Node, value ir.Expression) (ir.Expression, error) {'
assert source.count(anchor)==1
source=source.replace(anchor,anchor.replace('overloadResultCheck(', 'kindCensusRawOverloadResultCheck('))
assert source.count('import (')==1
source=source.replace('import (','import (\n "encoding/json"\n "os"',1)
source+='''
// Measurement-only observation: call original once, preserve its value/error, emit no usable IR.
func (l *lowering) overloadResultCheck(node *ast.Node, value ir.Expression) (ir.Expression, error) {
 if l.result == nil { return l.kindCensusRawOverloadResultCheck(node, value) }
 before := l.result.OverloadResultChecks
 result, err := l.kindCensusRawOverloadResultCheck(node, value)
 resolved := l.checker.GetResolvedSignature(node)
 if resolved != nil && resolved.Declaration() != nil && resolved.Declaration().Name() != nil && resolved.Declaration().Name().Text() == "injectClassNamedEvaluationHelperBlockIfMissing" {
  message := ""
  if err != nil { message = err.Error() }
  after := l.result.OverloadResultChecks
  _ = json.NewEncoder(os.Stderr).Encode(map[string]any{
   "kind_call_trace": true, "site": l.program.Where(node), "unit": latentFindingOwner,
   "promise": l.checker.TypeToString(l.checker.GetReturnTypeOfSignature(resolved)),
   "checked_delta": after.Checked-before.Checked, "proven_delta": after.Proven-before.Proven,
   "erased_delta": after.Erased-before.Erased, "marked": l.overloadCalls[node], "error": message,
  })
 }
 return result, err
}
'''
output=out/'overload_result_check.go.txt';output.write_text(source)
base['Replace'][str(file)]=str(output)
(out/'overlay.json').write_text(json.dumps(base,indent=2)+'\n')
print(json.dumps({'sourceRevision':'dfff19bb0ab5722136f7c19ea01505a1ab0946df','originalSource':str(file),'instrumentedSha256':hashlib.sha256(source.encode()).hexdigest(),'changes':'Read counters around one original method invocation; emit trace to stderr; preserve original IR expression/error and all output guards.'},indent=2))
