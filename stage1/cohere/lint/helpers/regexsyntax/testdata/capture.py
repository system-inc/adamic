#!/usr/bin/env python3
"""Capture every actual regexsyntax helper call made by consuming upstream cases.
Only declarations are renamed; original bodies and dependency calls are unchanged.
An added wrapper records arguments/results after delegating to that body.
"""
import json, pathlib, re, subprocess, sys
root = pathlib.Path(__file__).resolve().parents[6]
cohere = root / 'cohere'
out = pathlib.Path(sys.argv[1]).resolve(); out.mkdir(parents=True, exist_ok=True)
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=cohere,text=True).strip() == '7945d102a6c18dd36adf9114a758ce646e8b2359', 'cohere pin drift'
pkg = cohere / 'internal/lint/ecmascript/regexsyntax'
flags = 'RegexFlags{}'
# name, signature, arguments, byte input, index, end, capacity, flags, result names
specs = [
 ('IsHexDigit','b byte','b','[]byte{b}','0','0','0',flags,['value']),
 ('AllHexDigits','s string','s','[]byte(s)','0','0','0',flags,['value']),
 ('ParseHexUint','s string','s','[]byte(s)','0','0','0',flags,['value']),
 ('hexValue','b byte','b','[]byte{b}','0','0','0',flags,['value']),
 ('ParseRegexFlags','flags string','flags','[]byte(flags)','0','0','0',flags,['value']),
 ('PatternAndFlags','text string','text','[]byte(text)','0','0','0',flags,['pattern','flags']),
 ('ClassEnd','pattern string, start int, flags RegexFlags','pattern,start,flags','[]byte(pattern)','start','0','0','flags',['end','ok']),
 ('SkipPatternEscape','pattern string, i int, flags RegexFlags','pattern,i,flags','[]byte(pattern)','i','0','0','flags',['step','ok']),
 ('readRawClassChar','pattern string, i int, flags RegexFlags','pattern,i,flags','[]byte(pattern)','i','0','0','flags',['element','step','ok']),
 ('readClassEscape','pattern string, i int, flags RegexFlags','pattern,i,flags','[]byte(pattern)','i','0','0','flags',['element','step','ok']),
 ('parseRegexCharacterClass','pattern string, start int, flags RegexFlags, elementCapacity int','pattern,start,flags,elementCapacity','[]byte(pattern)','start','0','elementCapacity','flags',['elements','end','ok']),
 ('ParseRegexCharacterClassWithEnd','pattern string, start, end int, flags RegexFlags','pattern,start,end,flags','[]byte(pattern)','start','end','0','flags',['elements','parsedEnd','ok']),
]
returns = {'IsHexDigit':'bool','AllHexDigits':'bool','ParseHexUint':'uint32','hexValue':'uint32','ParseRegexFlags':'RegexFlags','PatternAndFlags':'(string,string)','ClassEnd':'(int,bool)','SkipPatternEscape':'(int,bool)','readRawClassChar':'(RegexCharElement,int,bool)','readClassEscape':'(RegexCharElement,int,bool)','parseRegexCharacterClass':'([]RegexCharElement,int,bool)','ParseRegexCharacterClassWithEnd':'([]RegexCharElement,int,bool)'}
replace = {}; wrappers=[]
for file in ['hex.go','literal.go','charclass.go']:
 text=(pkg/file).read_text()
 for name,*_ in specs:
  text=text.replace('func '+name+'(', 'func adamicOriginal_'+name+'(')
 text=text.replace('func (f RegexFlags) UV()', 'func (f RegexFlags) adamicOriginalUV()')
 side=out/file; side.write_text(text); replace[str(pkg/file)]=str(side)
for name,sig,args,b,i,end,cap,f,values in specs:
 vals=','.join(values)
 output='[]any{'+vals+'}'
 if name=='PatternAndFlags': output='[]any{adamicBytes([]byte(pattern)),adamicBytes([]byte(flags))}'
 if name=='ParseRegexFlags': output='[]any{value.Unicode,value.UnicodeSets}'
 if values[0]=='element': output='[]any{adamicElement(element),step,ok}'
 if values[0]=='elements': output='[]any{adamicElements(elements),'+','.join(values[1:])+'}'
 wrappers.append(f'func {name}({sig}) {returns[name]} {{ {vals} := adamicOriginal_{name}({args}); adamicRecord("{name}",{b},{i},{end},{cap},{f},{output}); return {vals} }}')
wrappers.append('func (f RegexFlags) UV() bool { value:=f.adamicOriginalUV(); adamicRecord("UV",nil,0,0,0,f,[]any{value}); return value }')
side=out/'record.go'; side.write_text('''package regexsyntax
import("encoding/json";"os";"sync";"sort")
var adamicMu sync.Mutex
var adamicRows = map[string]int{}
func adamicBytes(b []byte) []int { out:=make([]int,len(b)); for i,v:=range b {out[i]=int(v)}; return out }
func adamicElement(e RegexCharElement) []any { return []any{e.Kind,e.Value,e.IsUBrace,e.IsLoneSurrogate,e.Max,e.MaxIsUBrace,e.Start,e.End} }
func adamicElements(es []RegexCharElement) any { if es==nil {return nil}; out:=make([]any,len(es));for i,e:=range es {out[i]=adamicElement(e)};return out }
func adamicRecord(name string,b []byte,i,end,cap int,f RegexFlags,result any) { row:=[]any{name,adamicBytes(b),i,end,cap,f.Unicode,f.UnicodeSets,result}; data,err:=json.Marshal(row);if err!=nil {panic(err)};adamicMu.Lock();adamicRows[string(data)]++;adamicMu.Unlock() }
func AdamicDump(path string) { adamicMu.Lock();defer adamicMu.Unlock();keys:=make([]string,0,len(adamicRows));for k:=range adamicRows {keys=append(keys,k)};sort.Strings(keys);out,err:=os.Create(path);if err!=nil {panic(err)};defer out.Close();enc:=json.NewEncoder(out);for _,k:=range keys {var row any;if err=json.Unmarshal([]byte(k),&row);err!=nil {panic(err)};if err=enc.Encode([]any{adamicRows[k],row});err!=nil {panic(err)}} }
'''+'\n'.join(wrappers))
replace[str(pkg/'adamic_record.go')]=str(side)
controls=out/'controls_test.go'; controls.write_text(r"""package regexsyntax
import "testing"
func TestAdamicByteControls(t *testing.T) {
 for b:=0;b<256;b++ { IsHexDigit(byte(b));hexValue(byte(b)) }
 for _,s:=range []string{"","100000000","ffffffff","FFFFFFFFFFFFFFFF","zff"} {ParseHexUint(s);AllHexDigits(s)}
 for _,raw:=range [][]byte{{0xff},{0xe0,0x80,0x80},{0xed,0xa0,0x80},{0xf4,0x90,0x80,0x80},{0xf0,0x9f},{0xc2,0xa9},{0xf0,0x9f,0x91,0x8d},{}} {
  for bits:=0;bits<4;bits++ {flags:=RegexFlags{Unicode:bits&1!=0,UnicodeSets:bits&2!=0};readRawClassChar(string(raw),0,flags);readClassEscape("\\"+string(raw),0,flags)}
 }
}
""");replace[str(pkg/'adamic_controls_test.go')]=str(controls)
# These filters are derived from the inventory; additional helper unit tests provide controls.
inventory=json.loads((root/'stage1/cohere/lint/inventory/inventory.json').read_text())
filters={}
for rule in inventory['rules']:
 if any('/regexsyntax.' in d['symbol'] for d in rule['dependencies']):
  for path in rule['tests']['files']:
   file=root/path; package=str(file.parent.relative_to(cohere)); text=file.read_text()
   names=re.findall(r'func (Test\w+)\(',text); filters.setdefault(package,set()).update(names)
filters['internal/lint/ecmascript/regexsyntax']={'Test','Fuzz'}
for package,names in sorted(filters.items()):
 package_name=(cohere/package).name
 main=out/(package_name+'_main_test.go'); dest=out/(package_name+'.jsonl')
 main.write_text(f'''package {package_name}
import("os";"testing"; rs "github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax")
func TestMain(m *testing.M) {{ code:=m.Run(); rs.AdamicDump({json.dumps(str(dest))}); os.Exit(code) }}
''')
 if package_name == 'regexsyntax':
  main.write_text(main.read_text().replace('; rs "github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"', '').replace('rs.AdamicDump', 'AdamicDump'))
 replace[str(cohere/package/'adamic_capture_test.go')]=str(main)
overlay=out/'overlay.json'; overlay.write_text(json.dumps({'Replace':replace}))
for package,names in sorted(filters.items()):
 command=['go','test','-overlay='+str(overlay),'./'+package,'-run=^('+'|'.join(sorted(names))+')','-count=1','-timeout=10m']
 print(' '.join(command),flush=True);subprocess.run(command,cwd=cohere,check=True)
rows=[];counts={}
for file in sorted(out.glob('*.jsonl')):
 for line in file.read_text().splitlines():
  count,row=json.loads(line);rows.append(row);counts[row[0]]=counts.get(row[0],0)+count
(out/'cases.json').write_text(json.dumps(rows,separators=(',',':')))
(out/'want.txt').write_text(''.join(json.dumps(row[7],separators=(',',':'))+'\n' for row in rows))
(out/'counts.json').write_text(json.dumps(counts,sort_keys=True,indent=2))
print(json.dumps({'unique':len(rows),'calls':counts},sort_keys=True))
