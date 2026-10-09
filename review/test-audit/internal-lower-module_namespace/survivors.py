import pathlib,json,subprocess,os,time
os.chdir('/workspace/adamic');p=pathlib.Path('review/test-audit/internal-lower-module_namespace');matrix=json.loads((p/'matrix.json').read_text());plan=json.loads((p/'mutant-plan.json').read_text());survivors=[m for m in matrix if not m['kills']];results=[]
source='''package lower
import("context";"path/filepath";"testing";"github.com/system-inc/adamic/internal/load";"github.com/system-inc/adamic/internal/ir")
func TestAuditSurvivorObservation(t *testing.T){
path,err:=filepath.Abs("../oracle/testdata/namespaces_parser_factory.a");if err!=nil{t.Fatal(err)}
program,err:=load.Load([]string{path});if err!=nil{t.Fatal(err)}
file:=program.Files()[0];checker,release:=program.Checker(context.Background(),file);defer release()
l:=&lowering{program:program,checker:checker,result:&ir.Program{},this:-1,functionIndex:-1}
if err:=l.declareModule(file.Statements.Nodes);err!=nil{t.Fatal(err)}
count:=0;for _,local:=range l.result.Locals{if local.NamespaceVar{count++}}
t.Logf("OBSERVATION NamespaceVar bindings=%d",count)
}
'''
scratch=pathlib.Path('internal/lower/audit_observation_test.go');scratch.write_text(source);(p/'probes'/'survivor-observation.go.txt').write_text(source)
try:
 for m in survivors:
  mutation=next(q for q in plan if q['id']==m['id']);file=pathlib.Path(mutation['file']);before=file.read_text();observations=[]
  try:
   for label in ['before','after']:
    file.write_text(before if label=='before' else before.replace(mutation['old'],mutation['new']))
    cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^TestAuditSurvivorObservation$'];log=p/'logs'/(m['id']+'-witness-'+label+'.log')
    with log.open('w') as out:r=subprocess.run(cmd,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u037/cache/witness-'+m['id']+'-'+label),stdout=out,stderr=subprocess.STDOUT)
    text=log.read_text();observation=next((json.loads(s).get('Output','').strip() for s in text.splitlines() if s.startswith('{') and 'OBSERVATION' in s),None)
    observations.append(dict(label=label,exit=r.returncode,output=observation,command=' '.join(cmd)))
   changed=observations[0]['output']!=observations[1]['output'] and all(o['exit']==0 for o in observations)
   results.append(dict(id=m['id'],classification='unguarded behavior' if changed else 'equivalent candidate',observations=observations))
  finally:file.write_text(before)
finally:scratch.unlink(missing_ok=True)
(p/'survivors.json').write_text(json.dumps(results,indent=2));print(json.dumps(results,indent=2))
