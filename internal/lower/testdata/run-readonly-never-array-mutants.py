from pathlib import Path
import subprocess, os
root=Path(__file__).resolve().parents[3]
mutants=[
 ('accept-writable','internal/lower/invariance.go','\t\tif mutable && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source)) {','\t\tif source.Flags()&checker.TypeFlagsNever != 0 { continue }\n\t\tif mutable && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source)) {','./internal/lower','^TestWritableNeverArrayRefused$','want pinned writable never[] refusal, got <nil>'),
 ('reject-readonly','internal/lower/invariance.go',' || from.Flags()&checker.TypeFlagsNever != 0','', './internal/oracle','^TestNativeAgreesWithNode/internal/oracle/testdata/readonly_never_array','refuses'),
 ('lose-context','internal/lower/invariance.go','if contextual := l.checker.GetContextualType(parent, checker.ContextFlagsNone); contextual != nil && l.checker.IsArrayType(l.withoutUndefined(contextual)) {','if contextual := l.checker.GetContextualType(parent, checker.ContextFlagsNone); false && contextual != nil {','./internal/oracle','^TestNativeAgreesWithNode/internal/oracle/testdata/readonly_never_array','refuses'),
 ('lose-empty-layout','internal/lower/readonly_never_array.go','return l.checker.IsArrayType(proven) &&','return false && l.checker.IsArrayType(proven) &&','./internal/oracle','^TestNativeAgreesWithNode/internal/oracle/testdata/readonly_never_array','array of never'),
 ('lose-push-ownership','internal/native/emit_expressions.go','e.line("if (%s->length == 0) { %s->references = %t; }", array, array, expression.Element.IsReference())','e.line("(void)%s;", array)','./internal/oracle','^TestNativeAgreesWithNode/internal/oracle/testdata/readonly_never_array','LeakSanitizer'),
 ('lose-splice-ownership','internal/native/emit_arrays.go','e.line("if (%s->length == 0) { %s->references = %t; }", array, array, splice.Element.IsReference())','e.line("(void)%s;", array)','./internal/oracle','^TestNativeAgreesWithNode/internal/oracle/testdata/readonly_never_array','LeakSanitizer'),
 ('lose-concat-ownership','internal/native/runtime/array.c','if (arrays[which]->length > 0) {','if (false && arrays[which]->length > 0) {','./internal/oracle','^TestNativeAgreesWithNode/internal/oracle/testdata/readonly_never_array','heap-use-after-free'),
]
for name,file,old,new,package,test,catcher in mutants:
 source=root/file;original=source.read_text();assert original.count(old)==1,(name,original.count(old))
 try:
  source.write_text(original.replace(old,new))
  with Path('/tmp/readonly-never-mutant-'+name+'.log').open('w') as output:
   result=subprocess.run(['go','test',package,'-run',test,'-count=1','-timeout','30m'],cwd=root,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=output,stderr=subprocess.STDOUT)
  text=Path('/tmp/readonly-never-mutant-'+name+'.log').read_text()
  assert result.returncode!=0 and catcher in text,(name,text)
  assert 'clang:' not in text and 'runtime error:' not in text,(name,text)
  print(name+': caught by '+catcher)
 finally:source.write_text(original)
