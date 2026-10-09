import pathlib,subprocess,json,os
root=pathlib.Path('/workspace/adamic');src=root/'stage1/cohere/json';p=root/'review/test-audit/stage1-cohere-json-shards';f=src/'shards_test.go';s=f.read_text()
probe='''package json
import("testing"; "fmt"; "strings")
func TestU104ConstructionProbe(t *testing.T) {
 cases:=make([]textCase,39)
 for i:=range cases {cases[i]=textCase{Name:fmt.Sprintf("case-%d.json",i),Text:"{}"}}
 cases[17].Text=strings.Repeat(" ",128*1024+1)
 t.Logf("actual partition: %+v", jsonPortShards(cases))
}
'''
(p/'survivor-probe.go.txt').write_text(probe);probe_path=src/'u104_probe_test.go'
try:
 probe_path.write_text(probe)
 for id,source in [('S1-before',s),('S1-after',s.replace('end-start < 16','end-start < 15'))]:
  f.write_text(source)
  with (p/(id+'.log')).open('w') as log:subprocess.run(['go','test','-v','-count=1','./stage1/cohere/json/','-run','^TestU104ConstructionProbe$'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)
finally:
 f.write_text(s);probe_path.unlink(missing_ok=True)
