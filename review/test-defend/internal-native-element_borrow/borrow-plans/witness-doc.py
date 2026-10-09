import pathlib,subprocess,os,time,json
p=pathlib.Path('review/test-defend/internal-native-element_borrow/borrow-plans'); source=pathlib.Path('internal/native/element_borrow.go'); original=source.read_text(); old='case ir.Call:\n\t\tfor _, target := range program.CallTargets(expression) {';new='case ir.Call:\n\t\tif expression.Virtual != 0 {\n\t\t\treturn false\n\t\t}\n\t\tfor _, target := range program.CallTargets(expression) {';changed=original.replace(old,new,1);assert changed!=original
scratch=pathlib.Path('/tmp/defend-elements');scratch.mkdir(exist_ok=True);results=[];env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1'
def run(label,cmd,cache):
 env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-elements/cache/'+cache;started=time.monotonic()
 with (p/(label+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 results.append({'label':label,'command':cmd,'wall_seconds':time.monotonic()-started,'exit':r.returncode});print(label,r.returncode,round(results[-1]['wall_seconds'],3),flush=True);assert r.returncode==0
try:
 for id,text in [('clean',original),('D3',changed)]:
  source.write_text(text)
  run(id+'-doc-build',['timeout','90','go','run','./cmd/adamic','build','internal/oracle/testdata/devirt_borrow_doc_claim.a','-o',str(scratch/('doc-'+id)),'--count'],id)
  run(id+'-doc-execute',['timeout','30',str(scratch/('doc-'+id))],id)
finally:source.write_text(original);(p/'doc-witness-timings.json').write_text(json.dumps(results,indent=2)+'\n')
