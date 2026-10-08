"""Trace actual helper calls in all frozen consumer suites through a Go overlay."""
import json, os, re, subprocess, tempfile, pathlib, collections, gzip
HERE=pathlib.Path(__file__).resolve().parent
ROOT=HERE.parents[6]
COHERE=ROOT/'cohere'
import sys
sys.path.insert(0, str(ROOT/'stage1/cohere/lint/helpers/testdata'))
from pin import capture_pin
PIN=capture_pin(ROOT)
PKG=COHERE/'internal/lint/ecmascript/text'
specs={
 'MinimumEditDistance':('a string, b string','a,b','int',{'s':'a','t':'b'}),
 'BestMatch':('needle string, candidates []string, threshold int','needle,candidates,threshold','(string,bool)',{'s':'needle','arr':'candidates','n':'threshold'}),
 'GraphemeCount':('value string','value','int',{'s':'value'}),
 'UnescapeStringLiteralText':('text string','text','string',{'s':'text'}),
 'decodeEntity':('item string','item','(string,bool)',{'s':'item'}),
 'hexValue':('character rune','character','int',{'n':'character'}),
 'continuesCluster':('previous rune, character rune, startsWithPictograph bool, regionalIndicators int','previous,character,startsWithPictograph,regionalIndicators','bool',{'n':'previous','m':'character','b':'startsWithPictograph','k':'regionalIndicators'}),
 'hangulJoins':('previous rune, character rune','previous,character','bool',{'n':'previous','m':'character'})}
for name in ['hangulLeading','hangulVowel','hangulTrailing','hangulSyllable','hangulVowelSyllable','isGraphemeControl','isGraphemeExtend','isIndicLinker','isPictograph','isRegionalIndicator']:
 specs[name]=('character rune','character','bool',{'n':'character'})
with tempfile.TemporaryDirectory(prefix='text-capture-') as td:
 td=pathlib.Path(td); replace={}
 for name in ['distance.go','grapheme.go','jsx_entities.go']:
  src=(PKG/name).read_text()
  for symbol in specs: src=src.replace('func '+symbol+'(', 'func adamicOriginal'+symbol+'(')
  f=td/name;f.write_text(src);replace[str(PKG/name)]=str(f)
 code='''package text
import("encoding/json";"fmt";"os";"sync";"strings";"unicode/utf16")
var adamicMu sync.Mutex
func adamicUnits(s string) string {v:=[]string{};for _,c:=range utf16.Encode([]rune(s)){v=append(v,fmt.Sprint(c))};return strings.Join(v,",")}
func adamicRecord(row map[string]any){path:=os.Getenv("ADAMIC_TEXT_CAPTURE");if path==""{return};adamicMu.Lock();defer adamicMu.Unlock();data,e:=json.Marshal(row);if e!=nil{panic(e)};f,e:=os.OpenFile(path,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0644);if e!=nil{panic(e)};defer f.Close();_,e=f.Write(append(data,'\\n'));if e!=nil{panic(e)}}
'''
 for name,(params,args,result,fields) in specs.items():
  pair=result=='(string,bool)'
  code+=f'func {name}({params}) {result} {{\n'
  code+=('v,ok:=' if pair else 'v:=')+f'adamicOriginal{name}({args})\n'
  rendered='adamicUnits(v)+"|"+fmt.Sprint(ok)' if pair else 'adamicUnits(v)' if result=='string' else 'fmt.Sprint(v)'
  fields={'symbol':'"'+name+'"',**fields,'want':rendered,'inputSet':'os.Getenv("ADAMIC_TEXT_INPUT_SET")'}
  code+='adamicRecord(map[string]any{'+','.join('"'+k+'":'+v for k,v in fields.items())+'})\n'
  code+='return v'+(',ok' if pair else '')+'\n}\n'
 f=td/'trace.go';f.write_text(code);replace[str(PKG/'adamic_trace.go')]=str(f)
 controls='''package text
import("testing";"unicode")
func TestAdamicTextControls(t *testing.T){
 points:=map[rune]bool{-1:true,0x110000:true}
 for _,tab:=range unicode.Categories{for _,r:=range tab.R16{for _,p:=range []rune{rune(r.Lo)-1,rune(r.Lo),rune(r.Hi),rune(r.Hi)+1}{points[p]=true}};for _,r:=range tab.R32{for _,p:=range []rune{rune(r.Lo)-1,rune(r.Lo),rune(r.Hi),rune(r.Hi)+1}{points[p]=true}}}
 for _,p:=range []rune{0,13,10,0x94d,0x915,0x9cd,0xacd,0xb4d,0xc4d,0xd4d,0x200c,0x200d,0xe33,0xeb3,0x1100,0x115f,0x1160,0x11a7,0x11a8,0x11ff,0xac00,0xac01,0xd7a3,0x1f1e6,0x1f1ff,0x1f3fb,0x1f3ff,0xe0020,0xe007f,0x1f000,0x1faff,0x2600,0x27bf}{points[p-1]=true;points[p]=true;points[p+1]=true}
 for p:=range points{hexValue(p);hangulLeading(p);hangulVowel(p);hangulTrailing(p);hangulSyllable(p);hangulVowelSyllable(p);isGraphemeControl(p);isGraphemeExtend(p);isIndicLinker(p);isPictograph(p);isRegionalIndicator(p)}
 for _,a:=range []rune{0,13,0x200d,0x94d,0x1f1e6,0x1100,0x1160,0x11a8,0xac00,0xac01}{for _,b:=range []rune{10,65,0x915,0x301,0x1f600,0x1f1e6,0x1100,0x1160,0x11a8,0xac00,0xac01}{hangulJoins(a,b);for _,pic:=range []bool{true,false}{for _,n:=range []int{0,1,2,3}{continuesCluster(a,b,pic,n)}}}}
 for _,s:=range []string{"","abc","a\\r\\nb","á","각","क् क","👨‍👩‍👧","🇬🇧🇦","&amp;&unknown;&#xD800;&#1114112;", "&;&&amp;","é😀"}{GraphemeCount(s);UnescapeStringLiteralText(s);for _,b:=range []string{"",s,"cafe","ab"}{MinimumEditDistance(s,b)};BestMatch(s,[]string{s+"x",s,"abc"},1)}
 for _,s:=range []string{"amp","unknown","#0","#xD800","#55296","#x110000","#1114112","#","#x","#xF","x y","#xq","#99999999999999999999999"}{decodeEntity(s)}
 for name:=range xhtmlEntities{decodeEntity(name)}
}
'''
 f=td/'controls_test.go';f.write_text(controls);replace[str(PKG/'adamic_controls_test.go')]=str(f)
 harness=COHERE/'internal/lint/testing/rule_testing.go'
 original=harness.read_text();anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
 assert original.count(anchor)==1
 f=td/'rule_testing.go';f.write_text(original.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result',1));replace[str(harness)]=str(f)
 overlay=td/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}))
 trace=td/'trace.jsonl';env=os.environ|{'ADAMIC_TEXT_CAPTURE':str(trace)}
 groups=[('next','GoogleFontDisplay|GoogleFontPreconnect|NextScriptForGa|NoBeforeInteractiveScriptOutsideDocument|NoCssTags|NoHtmlLinkForPages|NoPageCustomFont|NoTypos|NoUnwantedPolyfillio'),('core','IdLength'),('structure','BoundaryNoProjectThemeValue|NextNoNearMissRouteExport'),('../ecmascript/text','MinimumEditDistance|BestMatch|GraphemeCount|AdamicTextControls')]
 logs=[]
 for family,pattern in groups:
  pkg='./internal/lint/rules/'+family if not family.startswith('..') else './internal/lint/ecmascript/text'
  log=td/(family.replace('/','_')+'.log')
  with log.open('w') as output:
   res=subprocess.run(['go','test','-overlay='+str(overlay),pkg,'-run','^Test('+pattern+')','-count=1','-timeout=10m','-v'],cwd=COHERE,env=env|{'ADAMIC_TEXT_INPUT_SET':family,'COHERE_DOCS_CAPTURE':str(td/'inputs')},stdout=output,stderr=subprocess.STDOUT,text=True)
  logs.append(log.read_text());print(family,res.returncode,flush=True)
  if res.returncode: raise RuntimeError(log.read_text()[-12000:])
 rows=[json.loads(line) for line in trace.read_text().split('\n') if line]
 rows.sort(key=lambda r:json.dumps(r,sort_keys=True))
 # Preserve duplicate calls: every recorded use is compared.
 with (HERE/'cases.json.gz').open('wb') as out:
  with gzip.GzipFile(filename='',mode='wb',fileobj=out,mtime=0) as z: z.write((json.dumps(rows,ensure_ascii=True,separators=(',',':'))+'\n').encode())
 inputs=[json.loads(l) for file in (td/'inputs').glob('*.jsonl') for l in file.read_text().split('\n') if l]
 fixtures={}
 for r in inputs:
  fixtures.setdefault(r['rule'],set()).add(json.dumps(r,sort_keys=True))
 consumers=[r['rule'] for r in json.loads((ROOT/'stage1/cohere/lint/helpers/comments/readiness.json').read_text())['remaining'] if any('/ecmascript/text.' in h for h in r['remaining_helpers'])]
 assert set(consumers)<=set(fixtures),set(consumers)-set(fixtures)
 (HERE/'coverage.json').write_text(json.dumps({'pin':PIN,'calls':dict(sorted(collections.Counter(r['symbol'] for r in rows).items())),'inputSets':{group:dict(sorted(collections.Counter(r['symbol'] for r in rows if r['inputSet']==group).items())) for group,_ in groups},'groups':groups,'consumerFixtures':{r:len(fixtures[r]) for r in sorted(consumers)}},indent=2)+'\n')
 (HERE/'capture.log').write_text(''.join(logs))
 print('captured',len(rows),'calls',flush=True)
