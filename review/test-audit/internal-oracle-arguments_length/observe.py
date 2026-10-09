import pathlib,json,subprocess,os
p=pathlib.Path('review/test-audit/internal-oracle-arguments_length');plan=json.loads((p/'mutant-plan.json').read_text());m=next(q for q in plan if q['id']=='M03');file=pathlib.Path(m['file']);before=file.read_text()
helper=pathlib.Path('internal/native/audit_observer.go');test=pathlib.Path('internal/oracle/audit_observer_test.go')
helpercode='package native\nimport "github.com/system-inc/adamic/internal/ir"\nfunc AuditClosureThrownObservation() string {e:=emitter{program:&ir.Program{ClosuresMayThrow:true}};e.closureThrown();return e.out.String()}\n'
testcode='package oracle\nimport("testing";"github.com/system-inc/adamic/internal/native")\nfunc TestAuditSurvivorObservation(t *testing.T){t.Logf("OBSERVATION closureThrown=%q",native.AuditClosureThrownObservation())}\n'
helper.write_text(helpercode);test.write_text(testcode);(p/'probes'/'survivor-native-observer.go.txt').write_text(helpercode);(p/'probes'/'survivor-oracle-observer.go.txt').write_text(testcode);results=[]
try:
 for label in ['before','after']:
  file.write_text(before if label=='before' else before.replace(m['old'],m['new']));cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^TestAuditSurvivorObservation$'];log=p/'logs'/('M03-observe-'+label+'.log')
  with log.open('w') as out:r=subprocess.run(cmd,env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u055/cache/observer-'+label),stdout=out,stderr=subprocess.STDOUT)
  events=[json.loads(s) for s in log.read_text().splitlines() if s.startswith('{')];output=next((e.get('Output','').strip() for e in events if 'OBSERVATION' in e.get('Output','')),None);results.append(dict(label=label,exit=r.returncode,command=' '.join(cmd),output=output))
finally:file.write_text(before);helper.unlink(missing_ok=True);test.unlink(missing_ok=True)
(p/'survivors.json').write_text(json.dumps([dict(id='M03',classification='behavior changed, unguarded in bounded matrix; outside kills unknown',observations=results)],indent=2));print(json.dumps(results,indent=2))
