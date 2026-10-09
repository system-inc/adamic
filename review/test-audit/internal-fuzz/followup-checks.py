import pathlib,json,subprocess,time,os,difflib
p=pathlib.Path('review/test-audit/internal-fuzz')
f=pathlib.Path('internal/lower/lower.go');original=f.read_text();entry='func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {';new=original.replace('\t"context"','\t"context"\n\t"os"').replace(entry,entry+'\n\tif os.Getenv("ADAMIC_MUTANT") == "PLower" { return nil, nil }')
(p/'PLower.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),new.splitlines(True),fromfile='a/'+str(f),tofile='b/'+str(f))))
try:
 f.write_text(new)
 with (p/'PLower-vet.log').open('w') as log:subprocess.run(['timeout','90','go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT,check=True)
 env=os.environ.copy();env['ADAMIC_MUTANT']='PLower';env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u020/cache/PLower';start=time.monotonic()
 with (p/'PLower.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/fuzz/','-run','.'],stdout=log,stderr=subprocess.STDOUT,env=env)
 (p/'lower-probe-run.json').write_text(json.dumps(dict(id='PLower',status=r.returncode,wall=time.monotonic()-start,log=str(p/'PLower.log'),regex='.'),indent=2)+'\n')
finally:f.write_text(original)
with (p/'coverage.log').open('w') as log:subprocess.run(['timeout','120','go','test','-count=1','-timeout','90s','-coverprofile='+str(p/'coverage.out'),'./internal/fuzz/','-run','.'],stdout=log,stderr=subprocess.STDOUT,check=True)
with (p/'coverage-functions.txt').open('w') as log:subprocess.run(['go','tool','cover','-func='+str(p/'coverage.out')],stdout=log,stderr=subprocess.STDOUT,check=True)
# Separate diagnostic witness, never part of the original row matrix.
w=pathlib.Path('internal/fuzz/audit_behavior_test.go')
witness="""package fuzz
import "testing"
func TestAuditBehavior(t *testing.T) {
 t.Logf("signature Exact=true Text=abc matches abc-suffix: %v", (Signature{Text:"abc",Exact:true}).matches("abc-suffix"))
 t.Logf("simplest Number: %s", simplest(Number).String())
 t.Logf("bytesShared(63,128): %v", bytesShared(63,128))
 t.Logf("keeps different signature: %v", keeps(Signature{Kind:"refusal",Text:"first",Exact:true},Observation{Verdict:Invalid},Observation{Verdict:Invalid,Lines:map[string]string{"refusal":"second"}}))
}
"""
(p/'survivor-witness.go.txt').write_text(witness)
try:
 w.write_text(witness)
 for id in ['clean','M05','M12','M14','M16']:
  m=next((m for m in json.loads((p/'mutation-plan.json').read_text()) if m['id']==id),None)
  if m:
   f=pathlib.Path(m['file']);old=f.read_text();f.write_text(old.replace(m['old'],m['new']))
  try:
   with (p/(id+'-witness.log')).open('w') as log:subprocess.run(['timeout','90','go','test','-v','-count=1','./internal/fuzz/','-run','^TestAuditBehavior$'],stdout=log,stderr=subprocess.STDOUT,check=True)
  finally:
   if m:f.write_text(old)
finally:w.unlink(missing_ok=True)
