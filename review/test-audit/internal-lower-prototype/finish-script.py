import pathlib,subprocess,os,time,json
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/internal-lower-prototype';menu=json.loads((p/'menu.json').read_text());env=os.environ.copy();runs=[]
def run(cmd,log,id=None):
 ev=env.copy()
 if id:ev['ADAMIC_BUILD_CACHE_DIR']='/tmp/u043/cache/witness-'+id
 start=time.monotonic()
 with (p/log).open('w') as f:q=subprocess.run(cmd,cwd=r,env=ev,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start));(p/'finish-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,flush=True);assert q.returncode==0
src='''package lower
import("testing";"os/exec")
func TestU043SourceObservation(t *testing.T) {
 got:=escapeRegexSource("[[/]/", "u")
 node,err:=exec.Command("node","-e","console.log(new RegExp('[[/]/','u').source)").CombinedOutput()
 if err!=nil {t.Fatalf("Node: %v %s",err,node)}
 t.Logf("input=%q flags=u lowering=%q Node=%q", "[[/]/", got, string(node))
}
func TestU043RefusalObservation(t *testing.T) {
 _,err:=lowerSource(t,"function made(pattern: string): RegExp { return new RegExp(pattern); }")
 t.Logf("refusal=%v",err)
}
'''
w=r/'internal/lower/u043_observation_test.go';(p/'observation-source.go.txt').write_text(src)
try:
 assert not (r/'internal/lower/u043_mutant.go').exists()
 w.write_text(src)
 for id,test in [('clean-source','TestU043SourceObservation'),('M20','TestU043SourceObservation'),('clean-refusal','TestU043RefusalObservation'),('M17','TestU043RefusalObservation')]:
  m=next((m for m in menu if m['id']==id),None)
  if m:
   f=r/m['file'];base=f.read_text();assert base==subprocess.check_output(['git','show','origin/main:'+m['file']],cwd=r).decode();f.write_text(base.replace(m['old'],m['new'],1))
  try:run(['timeout','120','go','test','-v','-count=1','-timeout','90s','./internal/lower/','-run','^'+test+'$'],'witness-'+id+'.log',id)
  finally:
   if m:f.write_text(base)
finally:
 if w.exists():w.unlink()
run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],'final-baseline.log')
run(['go','vet','./internal/lower/'],'final-vet.log')
run(['git','diff','--check'],'source-diff-check.log')
