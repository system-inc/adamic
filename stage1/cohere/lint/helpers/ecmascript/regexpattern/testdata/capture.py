"""Capture every actual regexpattern call in both consuming Go suites and package controls."""
import collections, gzip, json, os, pathlib, subprocess, tempfile
HERE=pathlib.Path(__file__).resolve().parent
ROOT=HERE.parents[6]; COHERE=ROOT/'cohere'; PKG=COHERE/'internal/lint/ecmascript/regexpattern'
PURE={
 'consumeDigits':('pattern string, index int','pattern,index','int',{'arr':'adamicBytes(pattern)','n':'index'}),
 'braceQuantifierEnd':('pattern string, start int','pattern,start','(int,bool)',{'arr':'adamicBytes(pattern)','n':'start'}),
 'isAsciiLetter':('value byte','value','bool',{'n':'int(value)'}),
 'octalEscapeValue':('digits string','digits','(uint32,bool)',{'arr':'adamicBytes(digits)'}),
 'extendOctalEscape':('pattern string, start int, end int','pattern,start,end','int',{'arr':'adamicBytes(pattern)','n':'start','m':'end'}),
 'groupPrologueEnd':('pattern string, index int','pattern,index','int',{'arr':'adamicBytes(pattern)','n':'index'}),
 'decodeRune':('text string, index int','text,index','(uint32,int)',{'arr':'adamicBytes(text)','n':'index'}),
 'quantifierAt':('pattern string, index int','pattern,index','(int,bool)',{'arr':'adamicBytes(pattern)','n':'index'}),
 'escapeValue':('text string, kind CharacterKind','text,kind','(uint32,bool)',{'arr':'adamicBytes(text)','n':'int(kind)'}),
 'unicodeEscapeValue':('text string','text','(uint32,bool)',{'arr':'adamicBytes(text)'})}
METHODS={'kindFromSource':('start int, end int','start,end','CharacterKind',{'n':'start','m':'end'}),'applyQuantifier':('index int','index','int',{'n':'index'}),'emit':('character Character','character','bool',{}),'emitAndQuantify':('character Character','character','int',{})}
with tempfile.TemporaryDirectory(prefix='regexpattern-capture-') as scratch:
 td=pathlib.Path(scratch); replace={}
 for filename in ['escape.go','walk.go','pattern.go']:
  source=(PKG/filename).read_text().replace('func Walk(', 'func adamicOriginalWalk(')
  for name in ['run','walkClass','readEscape']:source=source.replace('func (w *walker) '+name+'(', 'func (w *walker) adamicOriginal'+name+'(')
  for name in PURE: source=source.replace('func '+name+'(', 'func adamicOriginal'+name+'(')
  for name in METHODS: source=source.replace('func (w *walker) '+name+'(', 'func (w *walker) adamicOriginal'+name+'(')
  f=td/filename;f.write_text(source);replace[str(PKG/filename)]=str(f)
 code='''package regexpattern
import("encoding/json";"fmt";"os";"sync";"strings";"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax")
var adamicMu sync.Mutex
func adamicBytes(s string) []int {out:=[]int{};for _,b:=range []byte(s){out=append(out,int(b))};return out}
func adamicChars(chars []Character) string {out:=[]string{};for _,c:=range chars{out=append(out,fmt.Sprintf("%d:%d:%d:%d:%d:%d",c.Value,c.Kind,c.Start,c.End,c.QuantifierDepth,c.ClassDepth))};return strings.Join(out,";")}
func adamicRecord(row map[string]any){path:=os.Getenv("ADAMIC_PATTERN_CAPTURE");if path==""{return};adamicMu.Lock();defer adamicMu.Unlock();data,e:=json.Marshal(row);if e!=nil{panic(e)};f,e:=os.OpenFile(path,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0644);if e!=nil{panic(e)};defer f.Close();_,e=f.Write(append(data,'\\n'));if e!=nil{panic(e)}}
'''
 for name,(params,args,result,fields) in PURE.items():
  pair=result.startswith('(')
  code+=f'func {name}({params}) {result} {{\n'+('v,z:=' if pair else 'v:=')+f'adamicOriginal{name}({args})\n'
  rendered='fmt.Sprint(v)+"|"+fmt.Sprint(z)' if pair else 'fmt.Sprint(v)'
  fields={'symbol':'"'+name+'"',**fields,'want':rendered,'inputSet':'os.Getenv("ADAMIC_PATTERN_INPUT_SET")'}
  code+='adamicRecord(map[string]any{'+','.join('"'+k+'":'+v for k,v in fields.items())+'})\nreturn v'+(',z' if pair else '')+'\n}\n'
 for name,(params,args,result,fields) in METHODS.items():
  code+=f'func (w *walker) {name}({params}) {result} {{\n'
  fields={'symbol':'"'+name+'"','arr':'adamicBytes(w.pattern)',**fields,'inputSet':'os.Getenv("ADAMIC_PATTERN_INPUT_SET")'}
  if name in ['emit','emitAndQuantify']:
   code+='qd,cd,stopped:=w.quantifierDepth,w.classDepth,w.stopped; original:=w.callback;chars:=[]Character{};response:=true;w.callback=func(c Character)bool{chars=append(chars,c);response=original(c);return response}\n'
   code+=f'v:=w.adamicOriginal{name}({args});w.callback=original\n'
   fields|={'n':'qd','m':'cd','b':'stopped','response':'response','value':'character.Value','kind':'character.Kind','start':'character.Start','end':'character.End','qd':'character.QuantifierDepth','cd':'character.ClassDepth','want':'fmt.Sprint(v)+"|"+fmt.Sprint(w.stopped)+"|"+fmt.Sprint(w.quantifierDepth)+"|"+fmt.Sprint(w.classDepth)+"|"+adamicChars(chars)'}
  else:
   code+=f'v:=w.adamicOriginal{name}({args})\n';fields['want']='fmt.Sprint(v)'
  code+='adamicRecord(map[string]any{'+','.join('"'+k+'":'+v for k,v in fields.items())+'})\nreturn v\n}\n'
 code+=r'''func Walk(pattern string, flags regexsyntax.RegexFlags, callback func(Character)bool)bool {
 chars:=[]Character{};responses:=[]bool{};v:=adamicOriginalWalk(pattern,flags,func(c Character)bool{chars=append(chars,c);r:=callback(c);responses=append(responses,r);return r})
 adamicRecord(map[string]any{"symbol":"Walk","arr":adamicBytes(pattern),"u":flags.Unicode,"v":flags.UnicodeSets,"responses":responses,"want":fmt.Sprint(v)+"|"+adamicChars(chars),"inputSet":os.Getenv("ADAMIC_PATTERN_INPUT_SET")});return v
}
func(w *walker)run()bool {
 qd,cd,stopped:=w.quantifierDepth,w.classDepth,w.stopped;original:=w.callback;chars:=[]Character{};responses:=[]bool{}
 w.callback=func(c Character)bool{chars=append(chars,c);r:=original(c);responses=append(responses,r);return r}
 v:=w.adamicOriginalrun();w.callback=original
 adamicRecord(map[string]any{"symbol":"run","arr":adamicBytes(w.pattern),"u":w.flags.Unicode,"v":w.flags.UnicodeSets,"n":qd,"m":cd,"b":stopped,"responses":responses,"want":fmt.Sprint(v)+"|"+fmt.Sprint(w.stopped)+"|"+fmt.Sprint(w.quantifierDepth)+"|"+fmt.Sprint(w.classDepth)+"|"+adamicChars(chars),"inputSet":os.Getenv("ADAMIC_PATTERN_INPUT_SET")});return v
}
func(w *walker)walkClass(start,end int)bool {
 qd,cd,stopped:=w.quantifierDepth,w.classDepth,w.stopped;original:=w.callback;chars:=[]Character{};responses:=[]bool{}
 w.callback=func(c Character)bool{chars=append(chars,c);r:=original(c);responses=append(responses,r);return r}
 v:=w.adamicOriginalwalkClass(start,end);w.callback=original
 adamicRecord(map[string]any{"symbol":"walkClass","arr":adamicBytes(w.pattern),"u":w.flags.Unicode,"v":w.flags.UnicodeSets,"n":start,"m":end,"stateqd":qd,"statecd":cd,"b":stopped,"responses":responses,"want":fmt.Sprint(v)+"|"+fmt.Sprint(w.stopped)+"|"+fmt.Sprint(w.quantifierDepth)+"|"+fmt.Sprint(w.classDepth)+"|"+adamicChars(chars),"inputSet":os.Getenv("ADAMIC_PATTERN_INPUT_SET")});return v
}
func(w *walker)readEscape(index int)(*Character,int,bool) {
 c,next,ok:=w.adamicOriginalreadEscape(index);chars:=[]Character{};if c!=nil{chars=append(chars,*c)}
 adamicRecord(map[string]any{"symbol":"readEscape","arr":adamicBytes(w.pattern),"u":w.flags.Unicode,"v":w.flags.UnicodeSets,"n":index,"want":fmt.Sprint(c!=nil)+"|"+fmt.Sprint(next)+"|"+fmt.Sprint(ok)+"|"+adamicChars(chars),"inputSet":os.Getenv("ADAMIC_PATTERN_INPUT_SET")});return c,next,ok
}
'''
 f=td/'trace.go';f.write_text(code);replace[str(PKG/'adamic_trace.go')]=str(f)
 controls=r'''package regexpattern
import("testing";"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax")
func TestAdamicPatternControls(t *testing.T){
 for n:=0;n<256;n++{isAsciiLetter(byte(n));decodeRune(string([]byte{byte(n)}),0)}
 for a:=0xc0;a<=0xf5;a++{for b:=0;b<256;b++{for _,tail:=range []string{"","\x80","\x80\x80","\xff\x80"}{decodeRune(string([]byte{byte(a),byte(b)})+tail,0)}}}
 for _,s:=range []string{"","123","00000000000000000000000000000","4294967296","377","400","7777777777777777777777777777777","é","8","9","0","\x00","11x","20000000000"}{octalEscapeValue(s)}
 patterns:=[]string{"","abc","0123456789","123é4","{2}","{2,}","{2,4}","{}","{,2}","{x}","{12","{1,2}?","*?","+","??","a+?b","(x)","(?:x)","(?=a)","(?!a)","(?<=a)","(?<!a)","(?<name>x)","(?<name","(?","(?(x)","\\0","\\377","\\400","\\0000","\\777","\\8","\\x00","\\u0000","\\cA","\\n","\\r","\\t","\\v","\\f","\\b","\\-","😀"}
 for _,s:=range patterns{
  for n:=0;n<=len(s);n++{consumeDigits(s,n);braceQuantifierEnd(s,n);quantifierAt(s,n);groupPrologueEnd(s,n);w:=&walker{pattern:s};w.applyQuantifier(n);for _,m:=range []int{n,n+1,n+2,len(s),len(s)+1}{extendOctalEscape(s,n,m);w.kindFromSource(n,m)}}
  w:=&walker{pattern:s};w.kindFromSource(-1,len(s));
  for _,n:=range []int{0,len(s)}{for _,qd:=range []int{0,2}{for _,cd:=range []int{0,3}{for _,response:=range []bool{true,false}{for _,stopped:=range []bool{true,false}{
   w:=&walker{pattern:s,quantifierDepth:qd,classDepth:cd,stopped:stopped,callback:func(Character)bool{return response}}
   c:=Character{Value:65,Kind:KindSymbol,Start:0,End:n,QuantifierDepth:17,ClassDepth:19};w.emit(c)
   w=&walker{pattern:s,quantifierDepth:qd,classDepth:cd,stopped:stopped,callback:func(Character)bool{return response}};w.emitAndQuantify(c)
  }}}}}
 }
 for _,s:=range []string{"","\\","\\n","\\0","\\377","\\400","\\d","\\b","\\x00","\\x1f","\\xzz","\\u0000","\\u{1F}","\\u{}","\\u{100000000}","\\u{ABC","\\cA","\\c1","\\é","\\😀","[a-z]","[\\x00-\\x1f]","[\\d]","[[a]--[b]]","[^]","[","a+?b{2}","(?:a)|(?<name>.)$","\xff"}{
  for _,flags:=range []regexsyntax.RegexFlags{{},{Unicode:true},{UnicodeSets:true}}{
   for _,response:=range []bool{true,false}{
    Walk(s,flags,func(Character)bool{return response})
    for _,stop:=range []bool{true,false}{w:=&walker{pattern:s,flags:flags,callback:func(Character)bool{return response},quantifierDepth:2,classDepth:3,stopped:stop};w.run()}
    w:=&walker{pattern:s,flags:flags,callback:func(Character)bool{return response},quantifierDepth:2,classDepth:3};w.walkClass(0,len(s))
   }
   for n:=0;n<len(s);n++{w:=&walker{pattern:s,flags:flags};w.readEscape(n)}
  }
  for kind:=KindSymbol;kind<=KindIdentityEscape;kind++{escapeValue(s,kind)}
  unicodeEscapeValue(s)
 }

}
'''
 f=td/'controls_test.go';f.write_text(controls);replace[str(PKG/'adamic_controls_test.go')]=str(f)
 harness=COHERE/'internal/lint/testing/rule_testing.go';source=harness.read_text();anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
 assert source.count(anchor)==1
 f=td/'rule_testing.go';f.write_text(source.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result'));replace[str(harness)]=str(f)
 overlay=td/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}));trace=td/'calls.jsonl'
 groups=[('core','NoControlRegex|NoRegexSpaces'),('regexpattern','.*')];logs=[]
 for group,pattern in groups:
  package='./internal/lint/rules/core' if group=='core' else './internal/lint/ecmascript/regexpattern'
  result=subprocess.run(['go','test','-overlay='+str(overlay),package,'-run','^Test('+pattern+')','-count=1','-timeout=10m','-v'],cwd=COHERE,env=os.environ|{'ADAMIC_PATTERN_CAPTURE':str(trace),'ADAMIC_PATTERN_INPUT_SET':group,'COHERE_DOCS_CAPTURE':str(td/'fixtures')},stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
  logs.append(result.stdout);print(group,result.returncode,flush=True)
  if result.returncode:raise RuntimeError(result.stdout[-12000:])
 rows=[json.loads(line) for line in trace.read_text().split('\n') if line];rows.sort(key=lambda r:json.dumps(r,sort_keys=True))
 with (HERE/'cases.json.gz').open('wb') as out:
  with gzip.GzipFile(filename='',mode='wb',fileobj=out,mtime=0) as z:z.write((json.dumps(rows,separators=(',',':'))+'\n').encode())
 fixtures={}
 for f in (td/'fixtures').glob('*.jsonl'):
  for line in f.read_text().split('\n'):
   if line:
    row=json.loads(line);fixtures.setdefault(row['rule'],set()).add(json.dumps(row,sort_keys=True))
 assert {'no-control-regex','no-regex-spaces'}<=set(fixtures)
 coverage={'pin':subprocess.check_output(['git','rev-parse','HEAD'],cwd=COHERE,text=True).strip(),'calls':dict(sorted(collections.Counter(r['symbol'] for r in rows).items())),'inputSets':{group:dict(sorted(collections.Counter(r['symbol'] for r in rows if r['inputSet']==group).items())) for group,_ in groups},'consumerFixtures':{rule:len(fixtures[rule]) for rule in sorted(fixtures)},'dependency':'ecmascript/regexsyntax bca49a5ac1694099bcae388642126bf5281f1702'}
 (HERE/'coverage.json').write_text(json.dumps(coverage,indent=2)+'\n');(HERE/'capture.log').write_text(''.join(logs));print('Captured',len(rows),'actual Go calls',flush=True)
