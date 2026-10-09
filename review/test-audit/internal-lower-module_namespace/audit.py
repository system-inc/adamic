import pathlib,json,subprocess,os,time,difflib,re
root=pathlib.Path('/workspace/adamic');os.chdir(root);p=root/'review/test-audit/internal-lower-module_namespace'
rows=json.loads((p/'rows.json').read_text())
plan=[]
def add(file,old,new,kind):
 text=(root/file).read_text();assert text.count(old)==1,(file,old,text.count(old));plan.append(dict(id='M%02d'%(len(plan)+1),file=file,line=text[:text.index(old)].count('\n')+1,old=old,new=new,kind=kind))
add('internal/lower/module_namespace.go','if l.namespaceValueNode(node) && l.moduleNamespace(node) {','if false {','flip condition')
add('internal/lower/namespace_receiver_scope.go','node.Kind != ast.KindArrowFunction','node.Kind == ast.KindArrowFunction','flip condition')
add('internal/lower/namespace_receiver_scope.go','if node.Kind == ast.KindThisKeyword {\n\t\t\treturn true','if node.Kind == ast.KindThisKeyword {\n\t\t\treturn false','change constant')
add('internal/lower/namespaces.go','ast.GetSourceFileOfNode(declaration).IsDeclarationFile','!ast.GetSourceFileOfNode(declaration).IsDeclarationFile','flip condition')
add('internal/lower/namespaces.go','if !initialized[declaration] {','if initialized[declaration] {','flip condition')
add('internal/lower/namespaces.go','if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression {','if node.Kind == ast.KindCallExpression {','flip condition')
add('internal/lower/namespaces_call_graph.go','g.walks++','g.walks += 2','change constant')
add('internal/lower/namespaces_call_graph.go','member.reaches = reaches','member.reaches = member.reads','swap value')
add('internal/lower/namespaces_call_graph.go','entry.reads[declaration] = node\n\t\t\t}\n\t\t}\n\t\tif node.Kind','_ = declaration\n\t\t\t}\n\t\t}\n\t\tif node.Kind','drop statement')
add('internal/lower/namespaces_call_graph.go','declaration == nil || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsConst)','declaration == nil || !ast.HasSyntacticModifier(declaration, ast.ModifierFlagsConst)','flip condition')
add('internal/lower/modules.go','statements = namespaceDeclarations(statements)','','drop statement')
add('internal/lower/modules.go','l.result.Locals[local].NamespaceVar = true','l.result.Locals[local].NamespaceVar = false','change constant')
add('internal/lower/locals.go','What:', 'What:', 'dummy') if False else None
add('internal/lower/locals.go','"the checker gave a declaration no symbol"','"the checker gave a declaration no binding"','change constant')
add('internal/lower/load_time_reads.go','return !l.provenModuleReads[node] && l.checked(local)','return l.checked(local)','drop condition')
add('internal/lower/namespace_callable.go','case "name", "length", "prototype", "caller", "arguments", "call", "apply", "bind", "toString":\n\t\treturn true','case "name", "length", "prototype", "caller", "arguments", "call", "apply", "bind", "toString":\n\t\treturn false','change constant')
# M08 is dropping the transitive union assignment in favor of the direct read set, a statement/value change.
plan[7]['kind']='change option (direct versus transitive read set)'
# All mutations above are fixed before any mutant run.
(p/'mutant-plan.json').write_text(json.dumps(plan,indent=2))
functions=(p/'coverage-functions.txt').read_text().splitlines();(p/'reached-functions.txt').write_text('\n'.join(s for s in functions if re.search(r'\s([\d.]+)%$',s) and float(re.search(r'\s([\d.]+)%$',s)[1])>0)+'\n')
original={m['file']:(root/m['file']).read_text() for m in plan}
(p/'diffs').mkdir(exist_ok=True)
results=[]
try:
 for m in plan:
  before=original[m['file']];after=before.replace(m['old'],m['new']);f=root/m['file'];f.write_text(after)
  (p/'diffs'/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
  start=time.monotonic()
  with (p/'logs'/('vet-'+m['id']+'.log')).open('w') as out:r=subprocess.run(['timeout','90','go','vet','./internal/lower/'],stdout=out,stderr=subprocess.STDOUT)
  m['vet_seconds']=time.monotonic()-start;m['vet_exit']=r.returncode;f.write_text(before)
  assert r.returncode==0,m['id']
 # Switch spans precisely the replaced expression/statement without changing the fixed mutants.
 switched=dict(original)
 for m in plan:
  old=m['old'];new=m['new']
  # Conditions and statements are switched by whole function duplicate: selecting one version at entry.
  text=switched[m['file']]
  # Text edits inside functions use conditional expression helpers for conditions, or conditional statements for statements.
  if old.startswith('if ') and old.endswith(' {'):
   replacement='if auditBool("'+m['id']+'", '+old[3:-2]+', '+new[3:-2]+') {'
  elif old=='g.walks++':replacement='if auditIs("'+m['id']+'") { g.walks += 2 } else { g.walks++ }'
  elif old=='member.reaches = reaches':replacement='if auditIs("'+m['id']+'") { member.reaches = member.reads } else { member.reaches = reaches }'
  elif old.startswith('entry.reads['):replacement='if !auditIs("'+m['id']+'") { entry.reads[declaration] = node }\n\t\t\t}\n\t\t}\n\t\tif node.Kind'
  elif old.startswith('statements ='):replacement='if !auditIs("'+m['id']+'") { statements = namespaceDeclarations(statements) }'
  elif old=='l.result.Locals[local].NamespaceVar = true':replacement='l.result.Locals[local].NamespaceVar = !auditIs("'+m['id']+'")'
  elif old.startswith('return !l.proven'):replacement='return (!l.provenModuleReads[node] || auditIs("'+m['id']+'")) && l.checked(local)'
  elif old.startswith('case '):replacement=old.replace('return true','return !auditIs("'+m['id']+'")')
  elif old.startswith('if node.Kind == ast.KindThisKeyword'):
   replacement=old.replace('return true','return !auditIs("'+m['id']+'")')
  elif old.startswith('"'):replacement='auditString("'+m['id']+'", '+old+', '+new+')'
  else:replacement='auditBool("'+m['id']+'", '+old+', '+new+')'
  assert old in text,m['id'];switched[m['file']]=text.replace(old,replacement)
 for f,t in switched.items():(root/f).write_text(t)
 helper=root/'internal/lower/audit_switch.go';helper.write_text('''package lower
import("os"; "github.com/microsoft/TypeScript/tsc/shim/ast")
func auditIs(id string) bool{return os.Getenv("ADAMIC_MUTANT")==id}
func auditBool(id string,a,b bool)bool{if auditIs(id){return b};return a}
func auditNode(id string,a,b *ast.Node)*ast.Node{if auditIs(id){return b};return a}
func auditString(id,a,b string)string{if auditIs(id){return b};return a}
''')
 with (p/'logs'/'compile-switch.log').open('w') as out:subprocess.run(['go','test','-c','-o','/tmp/u037.test','./internal/lower/'],stdout=out,stderr=subprocess.STDOUT,check=True)
 for m in plan:
  env=dict(os.environ,ADAMIC_MUTANT=m['id'],ADAMIC_BUILD_CACHE_DIR='/tmp/u037/cache/'+m['id']);log=p/'logs'/(m['id']+'.log');start=time.monotonic()
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.']
  with log.open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  events=[]
  for s in log.read_text().splitlines():
   if s.startswith('{'):
    try:events.append(json.loads(s))
    except:pass
  fails=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test')))
  panicked=any(e.get('Output','').startswith('panic:') for e in events)
  result=dict(id=m['id'],command=' '.join(cmd),exit=r.returncode,wall_seconds=time.monotonic()-start,kills=fails,panic=panicked,events=events)
  if panicked or r.returncode==124:
   result['bounded']=True;result['kills']=[]
   for row in rows:
    cmd2=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row['test']+'$'];l=p/'logs'/(m['id']+'-'+row['test']+'.log')
    with l.open('w') as out:rr=subprocess.run(cmd2,env=env,stdout=out,stderr=subprocess.STDOUT)
    if rr.returncode:result['kills'].append(row['test'])
  results.append(result);(p/'matrix.json').write_text(json.dumps(results,indent=2));print(m['id'],r.returncode,len(result['kills']),flush=True)
finally:
 for f,t in original.items():(root/f).write_text(t)
 (root/'internal/lower/audit_switch.go').unlink(missing_ok=True)
 (p/'mutant-plan.json').write_text(json.dumps(plan,indent=2))
