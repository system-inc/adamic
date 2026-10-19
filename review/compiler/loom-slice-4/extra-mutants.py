import json,subprocess,difflib
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-slice-4'; rows=[]
for file,evidence,test,catcher,expected in [
 ('internal/oracle/step16_generics_test.go','scout-array-result.go.txt','^TestStep16OptionalResultMutant$','caught by Node stdout',0),
 ('internal/oracle/step20_iteration_test.go','iteration-snapshot.go.txt','^TestStep20IterationOutcomes$/accessor_binding.a$','outcome changed',1),
 ('internal/lower/iteration_array_view.go','element-access-old-stop.go.txt','^TestElementAccessNodeArrayAgreement$',"stage 0 can't lower an ElementAccessExpression yet",1)]:
 original=root/file;mutant=out/evidence;mutant.write_text((root/'review/compiler/chain-slice-4'/evidence).read_text());overlay=out/(evidence+'.json');overlay.write_text(json.dumps({'Replace':{str(original):str(mutant)}}));log=out/(evidence+'.log')
 with log.open('w') as f:r=subprocess.run(['go','test','-overlay',str(overlay),'./internal/oracle','-run',test,'-count=1','-timeout','90s','-v'],stdout=f,stderr=subprocess.STDOUT,timeout=120)
 text=log.read_text();assert r.returncode==expected and catcher in text and '[build failed]' not in text,(evidence,r.returncode,text[-3000:]);print(evidence,'caught',flush=True);rows.append({'mutant':evidence,'exit':r.returncode,'catcher':catcher})
(out/'extra-mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
