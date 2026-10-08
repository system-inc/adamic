import json, os, subprocess, tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[5]
COHERE=ROOT/'cohere'
import sys
sys.path.insert(0, str(ROOT/'stage1/cohere/lint/helpers/testdata'))
from pin import capture_pin
PIN=capture_pin(ROOT)
functions={
'DecodeCompilerRuleOptions':('compiler_rule_options.go','string(raw)','false'),
'DecodeNoMethodSetStateOptions':('no_method_set_state_options.go','string(raw)','false'),
'isHookIdentifierName':('rules_of_hooks.go','name','false'),
'isComponentIdentifierName':('rules_of_hooks.go','name', 'false'),
'mentionsRef':('error_boundaries.go','name','false'),
'isJavaScriptIdentifier':('no_deprecated.go','name','false'),
'commentValueOf':('no_deprecated.go','comment.Text','comment.IsBlock'),
'jsxAnnotationIn':('no_deprecated.go','text','false'),
'isReactComponentBaseName':('no_multi_comp.go','name','false'),
}
with tempfile.TemporaryDirectory() as tmp:
 tmp=Path(tmp); replacements={}; sources={}
 for name,(file,arg,block) in functions.items():
  path=COHERE/'internal/lint/rules/react'/file
  text=sources.get(path,path.read_text()); start=text.index('func '+name+'('); brace=text.index('{',start)
  text=text[:brace+1]+'\n recordAdamicString("'+name+'", '+arg+', '+block+')'+text[brace+1:]; sources[path]=text
 for i,(path,text) in enumerate(sources.items()):
  side=tmp/f'source{i}.go';side.write_text(text); replacements[str(path)]=str(side)
 side=tmp/'capture.go'
 side.write_text("""package react
import ("encoding/json"; "os"; "sync";"unicode/utf8")
var adamicStringMutex sync.Mutex
func recordAdamicString(name, text string, block bool) {
 if !utf8.ValidString(text){panic("invalid UTF-8 helper input: "+name)}
 adamicStringMutex.Lock(); defer adamicStringMutex.Unlock()
 f,e:=os.OpenFile(os.Getenv("ADAMIC_STRING_CAPTURE"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if e!=nil{panic(e)};defer f.Close()
 if e=json.NewEncoder(f).Encode(struct{Name,Text string;Block bool}{name,text,block});e!=nil{panic(e)}
}
""")
 replacements[str(COHERE/'internal/lint/rules/react/adamic_string_capture.go')]=str(side)
 overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
 calls=tmp/'calls.jsonl'
 with (HERE/'capture.log').open('w') as log:
  subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/react','-count=1','-timeout=10m'],cwd=COHERE,env=os.environ|{'ADAMIC_STRING_CAPTURE':str(calls)},stdout=log,stderr=subprocess.STDOUT,check=True)
 rows=[json.loads(s) for s in calls.read_text().splitlines()]
 assert set(functions)=={r['Name'] for r in rows},set(functions)-{r['Name'] for r in rows}
 (HERE/'calls.json').write_text(json.dumps(rows,ensure_ascii=True)+'\n')
 print('captured',len(rows),'actual calls', {n:sum(r['Name']==n for r in rows) for n in functions})
