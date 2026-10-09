exec(open('/tmp/u047-run.py').read().split("if __name__")[0])
import difflib
menu=[]
def add(id,file,old,new,sw,kind='mutant'):
 s=(root/file).read_text(); assert old in s,(id,old); menu.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,switch=sw,kind=kind))
f='internal/native/element_borrow.go'
add('M1',f,'if function.Closure {','if !function.Closure {','if function.Closure != u047Mutant("M1") {')
add('M2',f,'if _, isStore := statement.(ir.SetIndex); isStore {','if _, isStore := statement.(ir.SetIndex); !isStore {','if _, isStore := statement.(ir.SetIndex); isStore != u047Mutant("M2") {')
add('M3',f,'case ir.ObjectLiteral, ir.ArrayPush, ir.MakeClosure, ir.MakeError, ir.Defined:\n\t\treturn true','case ir.ObjectLiteral, ir.ArrayPush, ir.MakeClosure, ir.Defined:\n\t\treturn true','case ir.MakeError:\n\t\treturn !u047Mutant("M3")\n\tcase ir.ObjectLiteral, ir.ArrayPush, ir.MakeClosure, ir.Defined:\n\t\treturn true')
add('M4',f,'if targets.Unknown {','if !targets.Unknown {','if targets.Unknown != u047Mutant("M4") {')
f='internal/native/fields.go'
add('M5',f,'offsets[field.Name] = -1','offsets[field.Name] = 0','offsets[field.Name] = -1\n\t\t\t\tif u047Mutant("M5") { offsets[field.Name] = 0 }')
add('M6',f,'if len(program.Regexps) != 0 || recordStorage {','if recordStorage {','if (len(program.Regexps) != 0 && !u047Mutant("M6")) || recordStorage {')
add('M7',f,'"symbolicLink": 3','"symbolicLink": 2','"symbolicLink": u047Number("M7", 3, 2)')
f='internal/native/emit.go'
add('M8',f,'bodies.WriteString("\\treturn 0;\\n}\\n")','bodies.WriteString("\\treturn 0;\\n}")','if u047Mutant("M8") { bodies.WriteString("\\treturn 0;\\n}") } else { bodies.WriteString("\\treturn 0;\\n}\\n") }')
f='internal/native/runtime/heap.c'
add('M9',f,'\tPOISON(slot, size);','\t(void)size;','\tif (!adamic_u047("M9")) { POISON(slot, size); }')
add('M10',f,'chunk *each = spares;','chunk *each = NULL;','chunk *each = adamic_u047("M10") ? NULL : spares;')
f='internal/native/runtime/ieee754.c'
add('M11',f,'return kernel_sin(x, z, 0);','return kernel_sin(z, x, 0);','return adamic_u047("M11") ? kernel_sin(z, x, 0) : kernel_sin(x, z, 0);')
f='internal/native/runtime/map.c'
add('M12',f,'\titerator->map->iterating--;','\t/* dropped exhausted iteration decrement */','\tif (!adamic_u047("M12")) { iterator->map->iterating--; }')
f='internal/lower/class.go'
sig='func (l *lowering) absentOptionalWrite(target *ast.Node) error {\n'
old='if l.omittedOptionals[member] && !presentBeforeWrite(target) && !l.literalGivesField(target) {\n\t\treturn l.notYet(target, "writing a possibly absent optional own field")\n\t}'
add('M13',f,old,'',old.replace('if l.omittedOptionals', 'if !u047Mutant("M13") && l.omittedOptionals'))
add('P1','internal/native/element_borrow.go','func planElementBorrows(program *ir.Program) (map[*ir.Statement]bool, map[int]bool) {\n','func planElementBorrows(program *ir.Program) (map[*ir.Statement]bool, map[int]bool) {\n\treturn nil, nil\n','func planElementBorrows(program *ir.Program) (map[*ir.Statement]bool, map[int]bool) {\n\tif u047Mutant("P1") { return nil, nil }\n','probe')
add('P2','internal/native/emit.go','func C(program *ir.Program) string {\n','func C(program *ir.Program) string {\n\treturn ""\n','func C(program *ir.Program) string {\n\tif u047Mutant("P2") { return "" }\n','probe')
add('P3','internal/native/fields.go','func uniformFieldOffsets(program *ir.Program) map[string]int {\n','func uniformFieldOffsets(program *ir.Program) map[string]int {\n\treturn nil\n','func uniformFieldOffsets(program *ir.Program) map[string]int {\n\tif u047Mutant("P3") { return nil }\n','probe')
add('P4','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\treturn nil, nil\n','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\tif u047Mutant("P4") { return nil, nil }\n','probe')
f='internal/native/runtime/string_build_impl.h'
for id,sig in [('P5','adamic_string *adamic_string_concat(size_t count, adamic_string *const parts[]) {\n'),('P6','adamic_string *adamic_string_allocate(size_t length) {\n')]:add(id,f,sig,sig+'\treturn NULL;\n',sig+'\tif (adamic_u047("'+id+'")) { return NULL; }\n','probe')
sig='adamic_map *adamic_map_new(bool string_keys, bool reference_values) {\n';add('P7','internal/native/runtime/map.c',sig,sig+'\treturn NULL;\n',sig+'\tif (adamic_u047("P7")) { return NULL; }\n','probe')
# Each exposed math function gets its own entry probe, selected in a single runtime build.
for file in ['internal/native/runtime/ieee754.c','internal/native/runtime/hypot.c']:
 for m in re.finditer(r'double (adamic_math_\w+)\([^\n]*\) \{\n',(root/file).read_text()):
  id='P_'+m[1]; sig=m[0];add(id,file,sig,sig+'  return 0;\n',sig+'  if (adamic_u047("'+id+'")) { return 0; }\n','probe')
add('S1','internal/native/heap_test.go','return value / 1024','return value / 1000','return value / 1000','setup')
(out/'menu.json').write_text(json.dumps(menu,indent=2)); (out/'diffs').mkdir(exist_ok=True)
for m in menu:
 s=(root/m['file']).read_text(); changed=s.replace(m['old'],m['new'],1)
 (out/'diffs'/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
covered=[]
for line in (out/'coverage-functions.log').read_text().splitlines():
 parts=line.split()
 if len(parts)==3 and parts[2]!='0.0%' and parts[0].startswith('github'):covered.append(line)
(out/'reached-functions.txt').write_text('\n'.join(covered)+'\n\nRuntime: string_build_impl.h allocate, adamic_string_allocate, adamic_string_concat, adamic_string_put; heap.c list_chunk, unlist_chunk, new_chunk, take, give, adamic_allocate, deallocate, adamic_retain, list, let_go, free_one, release_last, adamic_release; map.c allocate_zeroed, adamic_map_new, hash_key, same_key, find, rebuild, adamic_map_get, adamic_map_set, adamic_map_delete, adamic_map_free_children, adamic_map_iterate, adamic_map_iterator_next; all 21 adamic_math functions in ieee754.c and hypot.c and their static numerical helpers. Runtime inventory is conservative and not C coverage measured. Lower.Lower and absentOptionalWrite are reached by the refusal row. Load is preparation.\n')
print('Fixed menu saved:',len([m for m in menu if m['kind']=='mutant']),'mutants;',len([m for m in menu if m['kind']=='probe']),'entry probes')
