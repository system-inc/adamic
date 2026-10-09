import pathlib,json,difflib
r=pathlib.Path('/workspace/adamic');o=r/'review/test-audit/internal-native-arguments_length';o.mkdir(parents=True,exist_ok=True);p=[]
def add(id,file,old,new,mode='bool'):
 s=(r/file).read_text();assert s.count(old)==1,(id,s.count(old));p.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,mode=mode))
b='internal/native/borrow.go';e='internal/native/element_borrow.go';c='internal/native/emit.go';n='internal/native/native.go'
add('M01',c,'program.ClosureConventionNeeded()','!program.ClosureConventionNeeded()')
add('M02',c,'if regexCallbacks {','if !regexCallbacks {','condition')
add('M03','internal/native/library.go','strings.Contains(source, "#define "+feature+" 1\\n")','strings.Contains(source, feature)')
add('M04',n,'"-ffp-contract=off"','"-ffp-contract=fast"','string')
add('M05',b,'case ir.Read, ir.Conditional, ir.Coalesce, ir.Box, ir.Narrow, ir.Unwrap, ir.CheckedCast, ir.MaybeOf, ir.Defined, ir.Undefined,\n\t\tir.NumberConstant, ir.BooleanConstant, ir.StringConstant:\n\t\treturn false','case ir.Read, ir.Conditional, ir.Coalesce, ir.Box, ir.Narrow, ir.Unwrap, ir.CheckedCast, ir.MaybeOf, ir.Defined, ir.Undefined,\n\t\tir.NumberConstant, ir.BooleanConstant, ir.StringConstant:\n\t\treturn true','block')
add('M06',b,'default:\n\t\treturn false\n\t}\n\treturn true','default:\n\t\treturn true\n\t}\n\treturn true','block')
add('M07',b,'expression.Method || expression.Optional || expression.Object.Type() == ir.Weak','expression.Method || !expression.Optional || expression.Object.Type() == ir.Weak')
add('M08',b,'ok && names[write.Name]','ok && !names[write.Name]')
add('M09',b,'if targets.Unknown {\n\t\t\t\t\t\tsafe = false','if false && targets.Unknown {\n\t\t\t\t\t\tsafe = false','block')
add('M10',e,'if element || chain {','if element || false && chain {','condition')
add('M11',e,'local.Global || local.Captured || local.Function != function || !lendable(local.Type) || assigned[declare.Local]','local.Global || !local.Captured || local.Function != function || !lendable(local.Type) || assigned[declare.Local]')
add('M12','internal/native/region.go','!function.Closure && function.RestElement == 0 && function.Returns == ir.Object','!function.Closure && function.RestElement == 0 && function.Returns != ir.Object')
add('M13','internal/native/region.go','if !known || escapes {','if !known || !escapes {','condition')
add('M14','internal/native/reuse.go','consumed = consumed || plan.consumed[program.Functions[target].Parameters[position]]','consumed = consumed && plan.consumed[program.Functions[target].Parameters[position]]','statement')
add('M15','internal/native/emit_objects.go','valueType == ir.Union || valueType == ir.MaybeBoolean && !needed','valueType == ir.Union || valueType == ir.MaybeBoolean && needed')
add('M16','internal/native/emit_objects.go','"(argument_count > %d ? %s : %s)"','"(argument_count >= %d ? %s : %s)"','string')
add('M17','internal/native/emit_functions.go','count = ", size_t argument_count"','','statement')
add('M18','internal/native/runtime/case.c',"character - 'a' + 'A'","character - 'a' + 'B'",'c-expression')
add('M19','internal/native/runtime/case.c','encode(0x3c2, out)','encode(0x3c3, out)','c-expression')
add('M20','internal/native/runtime/case_tables.h','#define CASE_UNICODE_VERSION "17.0.0"','#define CASE_UNICODE_VERSION "16.0.0"','static')
(o/'plan.json').write_text(json.dumps(p,indent=2))
for m in p:
 s=(r/m['file']).read_text();new=s.replace(m['old'],m['new']);(o/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
print('20 production mutants predeclared; sources unchanged')
