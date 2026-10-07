#!/usr/bin/env python3
"""Export the production rule's exact multi-file fixture programs through Go.
Generated Go fixtures and overlays stay outside the repository; no harness edit.
"""
from pathlib import Path
import argparse,json,re,subprocess,os
p=argparse.ArgumentParser();p.add_argument('directory',type=Path);a=p.parse_args();out=a.directory.resolve();out.mkdir(parents=True,exist_ok=True)
r=Path(__file__).resolve().parents[4];source=(r/'cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams_test.go').read_text()
def balanced(start):
 depth=0;quote=None;escape=False
 for at in range(start,len(source)):
  c=source[at]
  if quote:
   if escape:escape=False
   elif c=='\\' and quote!='`':escape=True
   elif c==quote:quote=None
  elif c in ['"',"'",'`']:quote=c
  elif c=='{':depth+=1
  elif c=='}':
   depth-=1
   if depth==0:return source[start:at+1]
 raise RuntimeError('unclosed fixture')
cases=['correctnessRequireBlockingStandardStreamsCase'+balanced(m.start()) for m in re.finditer(r'\{name: "',source)]
library=re.search(r'library := (correctnessNoProcessExitAfterOutputLines\(.*?\n\t\))',source,re.S)[1]
for name in ['Facets','ShardWorker','PhiSocialUpload','LintEngineParity','NewMigration','Structure']:
 for fixed in ['false','true']:cases.append('correctnessRequireBlockingStandardStreams'+name+'('+fixed+')')
cases.append('correctnessRequireBlockingStandardStreamsCase{name:"recursive call-order cache",imports:[]string{correctnessRequireBlockingStandardStreamsImportBlock},lines:[]string{"function a(){b();blockStandardStreams();}","function b(){a();}","function main(){b();console.log(output);process.exit(1);}","if(flag)a();main();"}}')
export=out/'fixtures_test.go'
export.write_text('''package nexus
import("testing";"encoding/json";"os";"fmt")
func TestWave04ExportBlocking(t *testing.T){
library := '''+library+'''
cases := []correctnessRequireBlockingStandardStreamsCase{'''+',\n'.join(cases)+''',}
for i,c := range cases{ data,err:=json.Marshal(c.files());if err!=nil{t.Fatal(err)};if err=os.WriteFile(fmt.Sprintf("%s/%03d.json",os.Getenv("WAVE04_FIXTURES"),i),data,0600);err!=nil{t.Fatal(err)}}}
''')
virtual=r/'cohere/internal/lint/rules/nexus/wave04_export_test.go';overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(export)}}))
env=dict(os.environ,WAVE04_FIXTURES=str(out))
with (out/'export.log').open('wb') as log:
 subprocess.run(['go','test','-overlay',str(overlay),'./internal/lint/rules/nexus','-run','^TestWave04ExportBlocking$','-count=1'],cwd=r/'cohere',env=env,stdout=log,stderr=log,check=True)
records=[]
for path in sorted(out.glob('[0-9][0-9][0-9].json')):
 files=json.loads(path.read_text());directory=out/('case-'+path.stem);roots=[]
 for filename,text in files.items():
  local=directory/filename.removeprefix('/repository/');local.parent.mkdir(parents=True,exist_ok=True);local.write_text(text);roots.append(local)
 config=directory/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022'],'moduleDetection':'auto','types':[]},'files':[str(x) for x in roots]}))
 manifest=directory/'manifest';subject=directory/'modules/subject/Subject.ts';manifest.write_text('\n'.join(str(x) for x in sorted(roots) if not str(x).endswith('.d.ts'))+'\n')
 records.append(dict(name='blocking-'+path.stem,config=str(config),manifest=str(manifest),subject_sha256=__import__('hashlib').sha256(subject.read_bytes()).hexdigest()))
(out/'corpora.json').write_text(json.dumps(records,indent=2)+'\n')
print(len(records),'exact Go multi-file fixture programs exported')
