"""Scratch-only trace: preserve the original branch order and every return value."""
from pathlib import Path
import json
root=Path('/workspace/orm-main');out=Path('/workspace/op2-report/stage3/census/debug-assert-overloads');build=Path('/workspace/dao-probe');build.mkdir(exist_ok=True)
s=(root/'internal/lower/overload_callback_use.go').read_text()
s=s.replace('import (','import (\n "fmt"\n "os"',1)
# Trace every failing door outside the recursive walker, with the original source line.
prefix,body=s.split('func (l *lowering) overloadCallbackHoldsAny(',1)
original=(root/'internal/lower/overload_callback_use.go').read_text().splitlines()
lines=prefix.splitlines();current=''
for i,line in enumerate(lines):
 if line.startswith('func '):current=line
 if line.strip() in ('return false','return flow') and 'implementation' in current:
  # Include function identity and mapped flow information at fallback doors.
  extra=''
  if 'overloadCallbackArgument(' in current or 'overloadCallbackElement(' in current:extra=', "flow", l.checker.TypeToString(flow)'
  lines[i]=line.replace(line.strip(),'fmt.Fprintln(os.Stderr, "DOOR", implementation.Name().Text(), '+json.dumps(current.split(') ')[1].split('(')[0])+', '+str(i+1)+extra+'); '+line.strip())
prefix='\n'.join(lines)+'\n'
# Add a wrapper solely to carry the path into the same original recursive walk.
head,body=body.split('{',1)
wrapper='''func (l *lowering) overloadCallbackHoldsAny(t *checker.Type, depth int, seen map[*checker.Type]bool) bool {
 return l.debugAssertHoldsAny(t,depth,seen,[]string{"root: "+l.checker.TypeToString(t)})
}
func (l *lowering) debugAssertHoldsAny(t *checker.Type, depth int, seen map[*checker.Type]bool, path []string) bool {'''
body=body.replace('if t.Flags()&checker.TypeFlagsAny != 0 || depth >= 32 {\n\t\treturn true', '''if t.Flags()&checker.TypeFlagsAny != 0 || depth >= 32 {
        branch := "bound"; if t.Flags()&checker.TypeFlagsAny != 0 { branch = "any" }
        fmt.Fprintf(os.Stderr,"HOLDSANY branch=%s depth=%d flags=%d path=%q\\n",branch,depth,t.Flags(),path)
        return true''')
body=body.replace('for _, member := range t.Types()', 'for index, member := range t.Types()').replace('for _, argument := range l.checker.GetTypeArguments(t)', 'for index, argument := range l.checker.GetTypeArguments(t)').replace('for _, signature := range l.checker.GetSignaturesOfType(t, kind)', 'for sigIndex, signature := range l.checker.GetSignaturesOfType(t, kind)')
replacements={
'l.overloadCallbackHoldsAny(member, depth+1, seen)':'l.debugAssertHoldsAny(member, depth+1, seen, append(append([]string{},path...),fmt.Sprintf("union/intersection[%d]: %s",index,l.checker.TypeToString(member))))',
'l.overloadCallbackHoldsAny(argument, depth+1, seen)':'l.debugAssertHoldsAny(argument, depth+1, seen, append(append([]string{},path...),fmt.Sprintf("typeArgument[%d]: %s",index,l.checker.TypeToString(argument))))',
'l.overloadCallbackHoldsAny(l.checker.GetReturnTypeOfSignature(signature), depth+1, seen)':'l.debugAssertHoldsAny(l.checker.GetReturnTypeOfSignature(signature), depth+1, seen, append(append([]string{},path...),fmt.Sprintf("signature[%d,%d].return: %s",kind,sigIndex,l.checker.TypeToString(l.checker.GetReturnTypeOfSignature(signature)))))',
'l.overloadCallbackHoldsAny(l.checker.GetTypeOfSymbol(parameter), depth+1, seen)':'l.debugAssertHoldsAny(l.checker.GetTypeOfSymbol(parameter), depth+1, seen, append(append([]string{},path...),"signature.parameter "+parameter.Name+": "+l.checker.TypeToString(l.checker.GetTypeOfSymbol(parameter))))',
'l.overloadCallbackHoldsAny(l.checker.GetTypeOfSymbol(property), depth+1, seen)':'l.debugAssertHoldsAny(l.checker.GetTypeOfSymbol(property), depth+1, seen, append(append([]string{},path...),"property "+property.Name+": "+l.checker.TypeToString(l.checker.GetTypeOfSymbol(property))))',
'l.overloadCallbackHoldsAny(index.ValueType(), depth+1, seen)':'l.debugAssertHoldsAny(index.ValueType(), depth+1, seen, append(append([]string{},path...),"index.value: "+l.checker.TypeToString(index.ValueType())))'}
for a,b in replacements.items():assert a in body,a;body=body.replace(a,b)
assert 'l.overloadCallbackHoldsAny(' not in body
s=prefix+wrapper+body
# Label top-level walker entries; full returned path is emitted only if true.
s=s.replace('return l.debugAssertHoldsAny(t,depth,seen,', 'fmt.Fprintln(os.Stderr,"HOLDSANY enter",l.checker.TypeToString(t)); return l.debugAssertHoldsAny(t,depth,seen,',1)
# Callback gate records which operands actually ran, without calling a skipped relation.
a='if l.overloadCallbackHoldsAny(parameter, 0, map[*checker.Type]bool{}) || !l.censusRelated(argument, parameter) {'
b='''holdsAny := l.overloadCallbackHoldsAny(parameter, 0, map[*checker.Type]bool{})
        related := false; if !holdsAny { related = l.censusRelated(argument, parameter) }
        fmt.Fprintln(os.Stderr,"CALLGATE",implementation.Name().Text(),"argument",l.checker.TypeToString(argument),"parameter",l.checker.TypeToString(parameter),"holdsAny",holdsAny,"relatedIfEvaluated",related)
        if holdsAny || !related {'''
assert a in s;s=s.replace(a,b)
p=build/'overload_callback_use.go';p.write_text(s);(out/'overload_callback_use.instrumented.go.txt').write_text(s)
meta=json.loads(Path('/workspace/orm-build-main/entry-overlay/overlay.json').read_text());meta['Replace'][str(root/'internal/lower/overload_callback_use.go')]=str(p)
# Label the exact overload whose strict relation and body-use exception are tested.
s=(root/'internal/lower/census_small.go').read_text();a='if !l.censusRelated(given, takes) && !l.overloadCallbackUse(implementation, overload, actual, given, takes) {'
b='''related := l.censusRelated(given,takes)
        callback := false
        if !related { fmt.Fprintln(os.Stderr,"PARAMETER",label,"index",index,"given",l.checker.TypeToString(given),"takes",l.checker.TypeToString(takes)); callback=l.overloadCallbackUse(implementation,overload,actual,given,takes); fmt.Fprintln(os.Stderr,"PARAMETER RESULT",label,index,"related",related,"callback",callback) }
        if !related && !callback {'''
assert a in s;s=s.replace(a,b).replace('import (','import (\n "os"',1)
p=build/'census_small.go';p.write_text(s);(out/'census_small.instrumented.go.txt').write_text(s);meta['Replace'][str(root/'internal/lower/census_small.go')]=str(p)
(build/'overlay.json').write_text(json.dumps(meta,indent=2)+'\n');(out/'probe-overlay.json').write_text(json.dumps(meta,indent=2)+'\n')
print('scratch instrumentation built; original branch order retained')
