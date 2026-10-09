from pathlib import Path
import subprocess,os,json,time
out=Path('review/test-audit/stage1-cohere-markdownblocks-leaf_composition_products');scratch=Path('stage1/cohere/markdownblocks/u127_observation_test.go')
code='''package markdownblocks
import("testing";"os";"path/filepath")
func TestAuditConstructionObservation(t *testing.T) {
 root,err:=filepath.Abs(repository);if err!=nil {t.Fatal(err)}
 mode:=os.Getenv("U127_OBSERVATION")
 if mode=="S1" {path:=leafCompositionGoProduct(t,root,"adamic_markdown_lists","list_go.go");info,err:=os.Stat(path);if err!=nil{t.Fatal(err)};t.Logf("OBSERVED path=%q directory=%v",path,info.IsDir())}
 if mode=="S2" {p:=leafCompositionLoweredProduct(t);t.Logf("OBSERVED generated_C_bytes=%d",len(p.source))}
 if mode=="S3" {p:=leafCompositionLoweredProduct(t);path:=leafCompositionNative(t,p.source,true);info,err:=os.Stat(path);if err!=nil{t.Fatal(err)};t.Logf("OBSERVED path=%q directory=%v",path,info.IsDir())}
 if mode=="S4" {buildListLayoutSetup(t);t.Logf("OBSERVED sanitized=%q",listLayoutShared.sanitized)}
 if mode=="S6" {tableLayoutReady(t);t.Logf("OBSERVED input_cases=%d",len(tableLayoutInputs))}
}
'''
(out/'supplemental-observation.txt').write_text(code);result=[]
try:
 scratch.write_text(code)
 for mid in ['S1','S2','S3','S4','S6']:
  for phase in ['before','after']:
   if phase=='after':subprocess.run(['git','apply',str(out/(mid+'.diff'))],check=True)
   try:
    env=os.environ.copy();env['U127_OBSERVATION']=mid
    with (out/f'{mid}-observe-{phase}.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run','^TestAuditConstructionObservation$'],env=env,stdout=log,stderr=subprocess.STDOUT)
    observations=[]
    for line in (out/f'{mid}-observe-{phase}.log').read_text().splitlines():
     try:x=json.loads(line)
     except:continue
     if 'OBSERVED' in x.get('Output',''):observations.append(x['Output'].strip())
    result.append({'id':mid,'phase':phase,'exit':r.returncode,'observation':observations});(out/'construction-survivor-witnesses.json').write_text(json.dumps(result,indent=2))
   finally:
    if phase=='after':subprocess.run(['git','apply','-R',str(out/(mid+'.diff'))],check=True)
finally:scratch.unlink(missing_ok=True)
