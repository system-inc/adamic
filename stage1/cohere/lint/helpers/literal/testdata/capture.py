"""Record all live calls from both consuming rules, without changing Go answers."""
import json, os, subprocess, tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[5]
COHERE=ROOT/'cohere'
with tempfile.TemporaryDirectory(prefix='literal-capture-') as scratch:
    tmp=Path(scratch)
    original=COHERE/'internal/lint/ecmascript/literal/offsets.go'
    source=original.read_text()
    for name in ['cookedBytesProducedBy','producesNoCookedBytes']:
        assert source.count('func '+name+'(')==1
        source=source.replace('func '+name+'(', 'func original_'+name+'(',1)
    (tmp/'offsets.go').write_text(source)
    recorder='''package literal
import("encoding/json";"fmt";"os";"sync")
var captureMu sync.Mutex
func record(kind,a,b,want string){ captureMu.Lock();defer captureMu.Unlock();p:=os.Getenv("LITERAL_CAPTURE");if p==""{return};f,e:=os.OpenFile(p,os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if e!=nil{panic(e)};defer f.Close();e=json.NewEncoder(f).Encode(map[string]string{"kind":kind,"a":a,"b":b,"want":want});if e!=nil{panic(e)} }
func cookedBytesProducedBy(a,b string)int{v:=original_cookedBytesProducedBy(a,b);record("produced",a,b,fmt.Sprint(v));return v}
func producesNoCookedBytes(a string)bool{v:=original_producesNoCookedBytes(a);record("tail",a,"",fmt.Sprint(v));return v}
'''
    (tmp/'recorder.go').write_text(recorder)
    controls='''package literal
import "testing"
func TestAdamicLiteralControls(t *testing.T){
for _,s:=range []string{"","\\\\","\\\\\\n","\\\\\\r","\\\\\\r\\n","\\\\\\n\\\\\\r\\n","\\\\\\r\\nX","\\\\x","é","👍"}{producesNoCookedBytes(s)}
for _,a:=range []string{"","a","ab","é","👍","\\\\n","\\\\x41","\\\\uD83D","x"}{for _,b:=range []string{"","a","b","é","👍","\\n","ä","é"}{cookedBytesProducedBy(a,b)}}
}
'''
    (tmp/'controls_test.go').write_text(controls)
    overlay={'Replace':{str(original):str(tmp/'offsets.go'),str(original.parent/'adamic_capture.go'):str(tmp/'recorder.go'),str(original.parent/'adamic_controls_test.go'):str(tmp/'controls_test.go')}}
    (tmp/'overlay.json').write_text(json.dumps(overlay))
    results={}
    for label,pkg,pattern in [('no-regex-spaces','rules/core','^TestNoRegexSpaces'),('no-misleading-character-class','rules/core','^TestNoMisleadingCharacterClass'),('controls','ecmascript/literal','^TestAdamicLiteralControls')]:
        path=tmp/(label+'.jsonl')
        run=subprocess.run(['go','test','-json','-overlay='+str(tmp/'overlay.json'),'./internal/lint/'+pkg,'-run',pattern,'-count=1','-timeout=10m'],cwd=COHERE,env=os.environ|{'LITERAL_CAPTURE':str(path)},capture_output=True,text=True)
        if run.returncode:raise RuntimeError(run.stdout+run.stderr)
        events=[json.loads(s) for s in run.stdout.splitlines() if s.startswith('{')]
        assert not [e for e in events if e.get('Action')=='skip'],label
        rows=[json.loads(s) for s in path.read_text().splitlines()]
        assert rows,label
        (HERE/(label+'.json')).write_text(json.dumps(rows,ensure_ascii=True,indent=2)+'\n')
        results[label]={'calls':len(rows),'produced':sum(r['kind']=='produced' for r in rows),'tail':sum(r['kind']=='tail' for r in rows),'passed_tests':sum(e.get('Action')=='pass' and 'Test'in e for e in events)}
    (HERE/'coverage.json').write_text(json.dumps(results,indent=2)+'\n')
    print(json.dumps(results,indent=2))
