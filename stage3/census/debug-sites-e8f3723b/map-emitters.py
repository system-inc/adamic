import json,re
from pathlib import Path
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent
rows=json.loads((out/'rows.json').read_text())
def entry(file,needle):
 lines=(repo/file).read_text().splitlines();hits=[i for i,s in enumerate(lines) if needle in s]
 assert len(hits)==1,(file,needle,hits)
 i=hits[0];decl=max(j for j,s in enumerate(lines[:i+1]) if s.startswith('func '))
 return {'file':file,'line':i+1,'source_line':lines[i],'function_declaration_line':decl+1,'function_declaration':lines[decl]}
def e(file,needle):return entry('internal/lower/'+file,needle)
result={}
for s in rows['sites']:
 r=s['reason']
 if r.startswith('a value of type ') and ' seen as ' not in r:a=[e('expression.go','return 0, l.notYet(node, "a value of type "')]
 elif r.startswith('overload '):a=[e('census_small.go','" cannot be served by implementation parameter "')]
 elif 'which can write' in r:a=[e('invariance.go','what := "a value of type "'),e('invariance.go','return &Refused{Where: l.program.Where(node), What: what, Fix: fix}')]
 elif r.startswith('inherited library member '):a=[e('prototype.go','What: "inherited library member "')]
 elif r.startswith('a method read as a value '):a=[e('refusals.go','What: "a method read as a value ("')]
 elif r.startswith('reading '):a=[e('expression.go','return nil, l.notYet(node, "reading "+node.Text())')]
 elif r.startswith('a template interpolating '):a=[e('expression.go','"a template interpolating an object, an array, a map, a function or undefined"')]
 elif r.startswith('a readonly view read by a run-time key '):a=[e('reflect_get.go','return l.notYet(node, dynamicKeyedRead(node)+" "+reason)')]
 elif r.startswith('in that narrows '):a=[e('unknown.go','What: "in that narrows a union of object types')]
 elif r=='a cast the runtime can\'t check':a=[e('cast_proof.go','refused := &Refused{Where: l.program.Where(node), What: "a cast the runtime')]
 elif r.startswith('Object.'):
  a=[e('library_object.go','return nil, true, &Refused{Where: l.program.Where(node), What: "Object." + name')]
  a+=[e('library_object.go','return refused("prototypes expose' if r=='Object.setPrototypeOf' else 'return refused("property descriptors can change')]
 elif r.startswith('in when '):a=[e('unknown.go','return nil, l.notYet(node, "in when "+reason)'),e('unknown.go','reason = "a possibly absent spread has no property presence descriptors"')]
 elif r=='new an Identifier':a=[e('object.go','return nil, l.notYet(node, "new "+describe(created.Expression))')]
 elif r.startswith('an unproven value assigned to a numeric literal or enum member slot '):a=[e('invariance.go','What: "an unproven value assigned to a numeric literal or enum member slot "')]
 elif r.startswith('a checked predicate target whose check would read past '):a=[e('predicates_proof.go','return nil, nil, l.notYet(node, cause+" ("'),e('predicates_proof.go','return "a checked predicate target whose check would read past the depth bound"')]
 elif r=='returning a property or element assignment':a=[e('statements.go','return nil, l.notYet(node, "returning a property or element assignment")')]
 elif r.startswith('a member typed any '):a=[e('any_relations.go','l.refusal.deferred = l.notYet(node, "a member typed any ("')]
 elif r.startswith('Array as a value '):a=[e('library_globals.go','name+" as a value outside equality or typeof')]
 elif r=='assigning to a NonNullExpression':a=[e('assignments.go','return nil, l.notYet(target, "assigning to "+describe(target))')]
 else:raise ValueError(s)
 result[s['id']]=a
(out/'emitters.json').write_text(json.dumps(result,indent=2)+'\n')
print('mapped',len(result),'debug.ts sites')
