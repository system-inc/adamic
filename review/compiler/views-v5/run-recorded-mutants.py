import ast, concurrent.futures, json, os, pathlib, subprocess, time
root=pathlib.Path.cwd(); evidence=root/'review/compiler/views-v5'; directory=evidence/'recorded-mutants'; directory.mkdir(exist_ok=True); cases=[]
def add(test,fixture,variable,value): cases.append((test,fixture,variable,value))
for kind,fixture in [('skip','emit-root-wrong'),('shape','emit-root-wrong'),('nested','emit-root-nested')]: add('TestCheckedViewIntersectionSource',fixture,'SOURCE',kind)
for kind,fixture in [('skip','recursive-read'),('shape','recursive-read'),('nested','recursive-read'),('canonical','recursive-read'),('optional','recursive-optional-root'),('literal-mode','recursive-literal-wrong'),('literal-code','recursive-number-literal'),('literal-enabled','recursive-boolean-literal')]: add('TestCheckedViewIntersectionRecursiveDemand',fixture,'RECURSIVE',kind)
for kind in ['skip','shape','nested','tag']: add('TestCheckedViewIntersectionSelectedArms','leading-access-unknown-tag' if kind=='tag' else 'leading-access-wrong','ARM',kind)
for kind,fixture in [('skip','tracker-wrong'),('shape','tracker-wrong'),('nested','tracker-wrong'),('presence','tracker-missing'),('outer','tracker-root-wrong'),('absence','tracker-absent')]: add('TestCheckedViewIntersectionOriginalPairs',fixture,'ORIGINAL',kind)
source=ast.parse((root/'stage3/interface-downcasts/lane7/original/run-pair-mutants.py').read_text()); context=dict(bindable='TestCheckedViewIntersectionOriginalBindable',emit_class='TestCheckedViewIntersectionOriginalNodes',true='',zero='')
for node in source.body:
 if isinstance(node,ast.Assign) and any(isinstance(n,ast.Name) and n.id=='runs' for n in node.targets):
  for test,kind,fixture,_ in eval(compile(ast.Expression(node.value),'<pinned-mutant-inventory>','eval'),{'__builtins__':{}},context): add(test,fixture,'ORIGINAL',kind)
for test,fixture,variable,value in [('OriginalAbsorption','wrong','ABSORPTION','1'),('OriginalIdentifier','wrong','IDENTIFIER','1'),('OriginalIdentifier','helper-only','IDENTIFIER','1'),('OriginalIsolatedMembers','9454$/^wrong','ISOLATED','9454'),('OriginalIsolatedMembers','9657$/^wrong','ISOLATED','9657'),('OriginalHeritageUnion','wrong','HERITAGE_UNION','1'),('OriginalRankedImport','wrong','RANKED','9245')]: add('TestCheckedViewIntersection'+test,fixture,variable,value)
for identity in ['8920','40931','8923']: add('TestCheckedViewIntersectionOriginalRankedOperands',identity+'$/^wrong(-bigint)?','RANKED',identity)
for fixture in ['read','helper','callback','destructure']: add('TestCheckedViewIntersectionDeferredMember',fixture,'DEFERRED','1')
def run(case):
 test,fixture,variable,value=case; label=f'{variable}-{value}-{fixture}'.replace('/','_').replace('$','').replace('^',''); log=directory/(label+'.log'); env=dict(os.environ,ADAMIC_INTERSECTION_ORIGINAL_DECLS='/tmp/views-v5-intersection-declarations'); env['ADAMIC_INTERSECTION_'+variable+'_MUTANT']=value
 command=['go','test','./internal/oracle','-run','^'+test+'$/^'+fixture+'$','-count=1','-timeout=80s','-v']; start=time.monotonic()
 with log.open('w') as output:
  try: result=subprocess.run(command,env=env,stdout=output,stderr=subprocess.STDOUT,timeout=88).returncode
  except subprocess.TimeoutExpired: result=124
 text=log.read_text(); pins=text.count('got oracle.run'); skip='--- SKIP:' in text
 status='caught' if result not in (0,124) and pins>=3 and text.count('stderr:[]uint8{}, exitCode:0') >= 3 and not skip else 'blocked-or-unproven'
 item=dict(test=test,fixture=fixture,variable=variable,value=value,command=command,exit=result,seconds=round(time.monotonic()-start,3),pins=pins,status=status,log=str(log.relative_to(root)))
 print(json.dumps(item),flush=True); return item
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool: results=list(pool.map(run,cases))
(evidence/'recorded-mutant-results.json').write_text(json.dumps(results,indent=2)+'\n')
