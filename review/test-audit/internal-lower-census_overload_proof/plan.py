import pathlib,subprocess,json,difflib
p=pathlib.Path('review/test-audit/internal-lower-census_overload_proof'); files=['census_overload_proof.go','census_small.go','census_predicate_marker.go','predicates.go','element_access_fields.go','invariance.go','lower.go'];orig={f:subprocess.check_output(['git','show','HEAD:internal/lower/'+f],text=True) for f in files};items=[]
reached=[l for l in (p/'functions.log').read_text().splitlines() if not l.startswith('total:') and float(l.split()[-1].strip('%'))>0];(p/'reached-functions.txt').write_text('\n'.join(reached)+'\n')
def add(id,f,old,new,kind='flip condition',mode='bool',occ=0):
 assert old in orig[f],id
 start=orig[f].find(old)
 if occ:start=orig[f].find(old,start+len(old))
 assert start>=0
 items.append(dict(id=id,file='internal/lower/'+f,line=orig[f][:start].count('\n')+1,old=old,new=new,menu=kind,mode=mode,occ=occ))
add('M01','census_overload_proof.go','!unchanged','false','change option')
add('M02','census_overload_proof.go','return l.censusReturnProof(implementation, overload, nil, nil)','return false','return early','return')
add('M03','census_overload_proof.go','!l.censusHasUndefined(produced) || l.censusHasUndefined(promised)','!l.censusHasUndefined(produced) && l.censusHasUndefined(promised)')
add('M04','census_small.go','l.classAssignable(from, to) && l.widened(from, to, map[[2]*checker.Type]bool{}) == nil','l.classAssignable(from, to) || l.widened(from, to, map[[2]*checker.Type]bool{}) == nil')
add('M05','census_small.go','!l.censusRelated(given, takes)','!l.censusRelated(takes, given)','swap two arguments')
add('M06','census_small.go','ordinal := 0','ordinal := 1','change constant','statement')
add('M07','census_predicate_marker.go','safe = safe && accepted','safe = safe || accepted','flip condition','statement')
add('M08','predicates.go','return found && safe','return found || safe','flip condition','return')
add('M09','predicates.go','"there is no body proving this parameter"','""','change constant','string')
add('M10','element_access_fields.go','len(names) == 0','len(names) != 0')
add('M11','census_small.go','left.Type() != ir.Boolean || right.Type() != ir.Boolean','left.Type() == ir.Boolean || right.Type() == ir.Boolean')
add('M12','invariance.go','!fresh && !l.checker.IsReadonlySymbol(viewed) && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source))','fresh && !l.checker.IsReadonlySymbol(viewed) && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source))')
add('P01','lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\treturn nil, nil','empty-answer probe','probe')
(p/'manifest.json').write_text(json.dumps(items,indent=2));(p/'diffs').mkdir(exist_ok=True)
modified=orig.copy()
for x in items:
 f=x['file'].split('/')[-1];s=orig[f]; pos=s.find(x['old']);new=s[:pos]+x['new']+s[pos+len(x['old']):]
 (p/'diffs'/(x['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/'+x['file'],tofile='b/'+x['file'])))
 old=x['old'];id=x['id'];mode=x['mode']
 if mode=='bool':repl='auditU027Bool("'+id+'", func() bool { return '+old+' }, func() bool { return '+x['new']+' })'
 elif mode=='return':repl='if auditU027Selected("'+id+'") { '+x['new']+' }; '+old
 elif mode=='statement':
  if old.startswith('ordinal') :repl=old+'; if auditU027Selected("'+id+'") { ordinal = 1 }'
  else:repl='if auditU027Selected("'+id+'") { '+x['new']+' } else { '+old+' }'
 elif mode=='string':repl='auditU027String("'+id+'", '+old+', '+x['new']+')'
 else:repl=old+'\n if auditU027Selected("P01") { return nil, nil }'
 modified[f]=modified[f].replace(old,repl,1)
(p/'plan.md').write_text('Fixed production-code plan, written before mutant runs. Lower is the only tested entry. P01 is a probe, not a mutant. Twelve mutants chosen across six production files. The whole clean package took 51.080 seconds, so the 20-minute budget limits matrix size below the 20-mutant aim.\n\n'+ '\n'.join(f"{x['id']} {x['file']}:{x['line']} {x['menu']}: {x['old']} -> {x['new']}" for x in items)+'\n')
for f,s in modified.items():(p/(f+'.switched.txt')).write_text(s)
print('Reached functions:',len(reached),'planned mutants:',len(items)-1)
