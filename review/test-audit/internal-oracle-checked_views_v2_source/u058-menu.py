exec(open('/tmp/u058-run.py').read().split("if __name__")[0]);import difflib
menu=[]
def add(id,file,old,new,sw,kind='mutant'):
 s=(root/file).read_text();assert old in s,(id,old);menu.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,switch=sw,kind=kind))
f='internal/lower/view_unions_untagged.go'
add('M1',f,'if field.Optional {','if !field.Optional {','if field.Optional != u058Mutant("M1") {')
add('M2',f,'contract.Of == ir.Boolean || contract.Of == ir.MaybeNumber','contract.Of == ir.Array || contract.Of == ir.MaybeNumber','((contract.Of == ir.Boolean && !u058Mutant("M2")) || (contract.Of == ir.Array && u058Mutant("M2"))) || contract.Of == ir.MaybeNumber')
loop='for index, contract := range l.result.ViewContracts {\n\t\tif contract.Unsupported == "untagged object union" && l.supportsUntaggedRead(contract) {\n\t\t\tl.result.ViewContracts[index].Unsupported = ""\n\t\t}\n\t}'
add('M3',f,loop,'','if !u058Mutant("M3") { '+loop+' }')
add('M4',f,'contract.ProducerCertified = true','contract.ProducerCertified = false','contract.ProducerCertified = !u058Mutant("M4")')
add('M5',f,'if seen[id] {\n\t\t\treturn true','if seen[id] {\n\t\t\treturn false','if seen[id] {\n\t\t\treturn !u058Mutant("M5")')
f='internal/lower/view_unions_dispatch.go';add('M6',f,'if child.Kind == ir.ViewArray {','if child.Kind == ir.ViewObject {','if (child.Kind == ir.ViewArray && !u058Mutant("M6")) || (child.Kind == ir.ViewObject && u058Mutant("M6")) {')
f='internal/lower/view_unions_callable_members.go';add('M7',f,'len(signatures) != 1','len(signatures) != 2','len(signatures) != u058Number("M7",1,2)')
f='internal/native/view_unions_untagged.go';add('M8',f,'field.Contract, field.Optional)','field.Contract, !field.Optional)','field.Contract, field.Optional != u058Mutant("M8"))')
f='internal/javascript/view_unions_untagged.go';add('M9',f,'if(depth>128) return false;','if(depth>0) return false;','if(depth>(process.env.ADAMIC_MUTANT === "M9" ? 0 : 128)) return false;')
f='internal/lower/class_inheritance.go';add('M10',f,'&& !available[l.fieldName(child.Name())] {','&& available[l.fieldName(child.Name())] {','&& (available[l.fieldName(child.Name())] == u058Mutant("M10")) {')
f='internal/lower/iteration.go';add('M11',f,'if !l.iterationOrigin(where, members, &presence, 0) {','if l.iterationOrigin(where, members, &presence, 0) {','if l.iterationOrigin(where, members, &presence, 0) == u058Mutant("M11") {')
f='internal/lower/class_static.go';add('M12',f,'target.Parent == declaration && !ast.HasSyntacticModifier(target, ast.ModifierFlagsStatic)','target.Parent == declaration && ast.HasSyntacticModifier(target, ast.ModifierFlagsStatic)','target.Parent == declaration && (ast.HasSyntacticModifier(target, ast.ModifierFlagsStatic) == u058Mutant("M12"))')
f='internal/lower/class_features.go';add('M13',f,'if !fresh && !isClassInstance(proven)','if fresh && !isClassInstance(proven)','if (fresh == u058Mutant("M13")) && !isClassInstance(proven)')
f='internal/lower/locals.go';add('M14',f,'list.Flags&ast.NodeFlagsBlockScoped == 0','list.Flags&ast.NodeFlagsBlockScoped != 0','(list.Flags&ast.NodeFlagsBlockScoped == 0) != u058Mutant("M14")')
f='internal/lower/view_lazy.go';add('M15',f,'if strings.Contains(family, "views-v3: array element kind") {','if !strings.Contains(family, "views-v3: array element kind") {','if strings.Contains(family, "views-v3: array element kind") != u058Mutant("M15") {')
for id,f,sig,value in [('P1','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n','nil, nil'),('P2','internal/native/emit.go','func C(program *ir.Program) string {\n','""'),('P3','internal/javascript/javascript.go','func JavaScript(program *ir.Program) string {\n','""')]:
 s=(root/f).read_text();start=s.index(sig);body=start+len(sig);end=s.index('\n}\n',body)+3;old=s[start:end];new=sig+'\treturn '+value+'\n}\n';add(id,f,old,new,sig+'\tif u058Mutant("'+id+'") { return '+value+' }\n'+s[body:end],'probe')
 # Lower's empty implementation no longer needs these imports.
 if id=='P1':menu[-1]['extra_remove']=['\n\t"fmt"','\n\t"path/filepath"']
f='internal/oracle/oracle_test.go';old='func disagreement(oracle run, native run) string {\n';s=(root/f).read_text();end=s.index('\n}\n',s.index(old))+3;add('W1',f,s[s.index(old):end],old+'\treturn ""\n}\n',old+'\treturn ""\n}\n','witness')
(out/'diffs').mkdir(exist_ok=True)
for m in menu:
 s=(root/m['file']).read_text();changed=s.replace(m['old'],m['new'],1)
 for text in m.get('extra_remove',[]):changed=changed.replace(text,'')
 (out/'diffs'/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
(out/'menu.json').write_text(json.dumps(menu,indent=2))
covered=[l for l in (out/'coverage-functions.log').read_text().splitlines() if l.startswith('github') and l.split()[-1]!='0.0%'];(out/'reached-functions.txt').write_text('\n'.join(covered)+'\n')
print('Fixed menu: 15 production mutants, three entries, one comparison witness;',len(covered),'covered functions')
