from pathlib import Path
import re,subprocess
files=subprocess.check_output(['git','diff','--name-only','--diff-filter=U'],text=True).splitlines()
for name in files:
 p=Path(name);s=p.read_text();i=0
 def resolve(m):
  global i
  i+=1;a,b=m[1],m[2]
  if name.endswith('counts.md'):return a+b
  if name=='internal/ir/ir.go':
   extra='\t\tViewRead ArrayViewRead\n' if 'ViewRead' in a else '\t\tOptionalComparator bool\n'
   return extra+b
  if name=='internal/javascript/javascript.go':
   if 'counted := false' in b:
    pre=b[:b.index('\tif counted {\n\t\tbuilder.WriteString("const adamicCallee')]
    return '\tbuilder.WriteString(recordRuntime)\n'+pre+a[a.index('\t// object.name'):]
   return b
  if name=='internal/lower/expression.go':
   if 'isArgumentsLength' in b:return a+'\t}\n'+b
   if 'func (l *lowering)' in a or a.startswith('// functionValue'):return b
   if 'callArguments(call.Arguments' in b:
    return '\tif rest := l.censusRestDeclaration(function); rest != nil {\n\t\treturn l.censusRestCall(call, function, rest)\n\t}\n'+b.replace('return ir.Call{','return l.censusOverloadResult(call, ir.Call{').replace('Returns: l.result.Functions[function].Returns}, nil','Returns: l.result.Functions[function].Returns})')
   return b
  if name=='internal/lower/functions.go':return a+'\t\t}\n'+b
  if name=='internal/lower/object.go':
   field=b.split('CallbackType: ',1)[1].rsplit('}, true, nil',1)[0]
   return a.replace('}, true, nil',', CallbackType: '+field+'}, true, nil')
  if name=='internal/lower/lower_test.go':return b
  if name=='internal/native/emit_functions.go':
   if 'function.ArgumentsCount' in b:return '\t\te.line("(void)argument_count;")\n'+b
   return a
  if name=='internal/native/emit_objects.go':return b.replace('value = unslotted(local.Type, fmt.Sprintf("arguments[%d].%s", index-1, member(local.Type)))','value = closureArgument(local.Type, index-1)')
  if name=='internal/native/emit_expressions.go':
   pre=a[:a.index('\t\te.line(')]
   return pre+b.replace('fmt.Sprintf("%s->elements[%s]", source, index)','argument')
  return b
 s=re.sub(r'<<<<<<< HEAD\n(.*?)=======\n(.*?)>>>>>>> 9534e8ab\n',resolve,s,flags=re.S)
 assert '<<<<<<<' not in s,(name,i);p.write_text(s)
