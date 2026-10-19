import json
from verify import run,root,out
rows=json.loads((root/'review/compiler/loom-views-fx-b/changed-counts.json').read_text())
text='package oracle\nimport "testing"\n'
for row in rows:
 name=row['path'].removeprefix('stage3/interface-downcasts/').removesuffix('.a')
 label=name.replace('/','_').replace('-','_')
 text+='func TestMovedCounts_'+label+'(t *testing.T) { t.Parallel(); program,path := interfaceFixture(t,"'+name+'"); want := onNode(t,path); if want.exitCode != 0 { t.Fatal(want) }; sanitized, leaked, comparable := nativelyCheckingLeaks(t,program,true); if leaked != "" || !comparable { t.Fatalf("leaks %s comparable %v", leaked, comparable) }; for lane,got := range map[string]run{"sanitized":sanitized,"native":released(t,program),"javascript":onJavaScriptBackend(t,program)} { t.Logf("%s exit=%d stdout=%q stderr=%q",lane,got.exitCode,got.stdout,got.stderr); if diff:=disagreement(want,got); diff!="" { t.Errorf("%s: %s",lane,diff) } } }\n'
p=out/'rows-extra_test.go.txt';p.write_text(text)
overlay=out/'rows-overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/'internal/oracle/rebased_rows_test.go'):str(p)}}))
raise SystemExit(run('moved-rows',['go','test','-overlay',str(overlay),'./internal/oracle','-run','^TestMovedCounts_','-count=1','-json','-timeout','90s'],150))
