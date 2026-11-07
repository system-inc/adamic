import json,pathlib,subprocess,re
D=pathlib.Path('/workspace/core-sites-main/stage3/census/core-sites-4186581e');r=json.loads((D/'rows.json').read_text());out={};searches=[]
def e(file,line,function,note=''):
 return dict(file='internal/lower/'+file,line=line,function=function,note=note)
for f in r['sites']:
 reason=f['reason'];key=f['phase']+'|'+f['kind']+'|'+reason
 if reason=="a cast the runtime can't check":
  entries=[e('cast_proof.go',188,'(*lowering).castProof','castRefusal at line 156 invokes this decision during the refusal scan')];pattern='"a cast the runtime can\'t check"'
 elif reason.startswith('a member typed any ('):
  entries=[e('any_relations.go',43,'(*lowering).refuseAnyRelation','stores a lowest-priority refusal.deferred stop; this is not deferred_generic_instantiation')];pattern='"a member typed any ("+found.path'
 elif reason=='a non-object representation passed through an opaque object parameter':
  entries=[e('detached_own.go',94,'(*lowering).detachedOwnRefusal')];pattern=reason
 elif reason.startswith('a method read as a value ('):
  entries=[e('refusals.go',282,'(*lowering).methodValueRefusal')];pattern='"a method read as a value ("'
 elif reason.startswith('a value of type T[]'):
  entries=[e('invariance.go',726,'(*lowering).wideningRefusal','builds the exact reason; returns Refused at line 740, reached by refuseWidening/refuseWholeWidening')];pattern='" seen as " + l.checker.TypeToString(contextual) + ", which can write "'
 elif reason.startswith('a BinaryExpression with '):
  entries=[e('expression.go',1132,'(*lowering).combine','typeName at line 1295 supplies union of differently held members')];pattern='describe(node)+" with a "'
 elif reason.startswith('inherited library member '):
  entries=[e('prototype.go',124,'(*lowering).prototypeRead')];pattern='"inherited library member " + name'
 elif reason.startswith('reading '):
  entries=[e('expression.go',660,'(*lowering).enumNeverValue','Identifier branch: the local is absent after an earlier failed statement was rolled back')];pattern='"reading "+node.Text()'
 elif reason=='new an Identifier':
  entries=[e('object.go',1811,'(*lowering).newExpression','new Array(length) is not handled by this fallback')];pattern='"new "+describe(created.Expression)'
 elif reason=='a value of type any':
  entries=[e('expression.go',53,'(*lowering).typeOf','the checker type has no supported runtime representation')];pattern='"a value of type "+l.checker.TypeToString(l.checker.GetTypeAtLocation(node))'
 elif reason=='assigning to an Identifier':
  entries=[e('assignments.go',65,'(*lowering).assignment','Identifier target has no represented local after rollback')];pattern='"assigning to "+describe(target)'
 elif reason.startswith('a type predicate whose return is not proven '):
  entries=[e('predicates.go',50,'(*lowering).proveFlowPredicate','decides that no function body proves the parameter'),e('predicates.go',40,'(*predicateFlowProof).refused','constructs the full reason')];pattern='"there is no body proving this parameter"'
 elif reason.startswith('the host binding process, pending '):
  entries=[e('host_bindings.go',139,'hostBindingStop','constructs the exact reason from the pinned host binding table')]
  if f['phase']=='lowering':entries.append(e('host_bindings.go',361,'(*lowering).hostBindingTypeOf','emits NotYet for typeof process'))
  elif f['site_unit_name']=='process':entries.append(e('host_bindings.go',107,'(*lowering).hostBindingStatement','emits NotYet at the declaration'))
  else:entries.append(e('host_bindings.go',269,'(*lowering).hostBindingRefusal','emits NotYet at the Identifier use'))
  pattern='"the host binding " + row.Name + ", pending "';key=f['id']
 elif reason=='an ambient value declaration':
  entries=[e('ambient.go',79,'(*lowering).ambientRefusal','selects the reason; constructs Refused at line 91')];pattern='"an ambient value declaration"'
 else:raise RuntimeError(reason)
 out[key]=entries
 if pattern not in [s['pattern'] for s in searches]:
  proc=subprocess.run(['rg','-n','-F','-g','*.go','-g','!*_test.go',pattern,'internal/lower'],text=True,capture_output=True)
  assert proc.returncode==0,(pattern,proc.stderr)
  searches.append({'pattern':pattern,'matches':proc.stdout.splitlines()})
 for item in entries:
  path=pathlib.Path(item['file']);lines=path.read_text().splitlines();item['source_line']=lines[item['line']-1]
  prior=[(n+1,t) for n,t in enumerate(lines[:item['line']]) if t.startswith('func ')]
  item['function_declaration_line']=prior[-1][0];item['function_declaration']=prior[-1][1]
(D/'emitters.json').write_text(json.dumps(out,indent=2)+'\n')
(D/'reason-searches.json').write_text(json.dumps(searches,indent=2)+'\n')
