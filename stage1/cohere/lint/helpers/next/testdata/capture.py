"""Capture every actual query-helper call in both unchanged consumer suites."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[5]
COHERE = ROOT / 'cohere'
PIN = '7945d102a6c18dd36adf9114a758ce646e8b2359'
assert subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=COHERE, text=True).strip() == PIN
controls = [
 {'url': u, 'key': k}
 for u in ['', 'x?display=swap', '//x?display=swap', 'HTTP://x?display=swap',
           'http://x', 'https://x?', 'https://x?display', 'https://x?display=',
           'https://x?display=swap&display=block', 'https://x?other=first&display=block',
           'https://x?display=a=b?c#d', 'https://x?%64isplay=swap&display=%2C+',
           'https://x?=empty&&display=🌍&features=Promise%2CSet',
           'https://x?display=é&features=Promise%2cSet', 'https://x?display=x\ny']
 for k in ['display', 'features', '', 'missing']
]
with tempfile.TemporaryDirectory(prefix='adamic-next-query-') as scratch:
 scratch = Path(scratch)
 source = COHERE / 'internal/lint/rules/next/google_font_display.go'
 text = source.read_text()
 assert text.count('func urlQueryValue(') == 1
 text = text.replace('"strings"', '"strings"\n "encoding/json"\n "os"\n "runtime"')
 text = text.replace('func urlQueryValue(', 'func adamicOriginalQueryValue(')
 text += '''
func urlQueryValue(url string, key string) (string, bool) {
 value, present := adamicOriginalQueryValue(url, key)
 if path := os.Getenv("ADAMIC_QUERY_CAPTURE"); path != "" {
  _, caller, line, _ := runtime.Caller(1)
  row := struct { URL string `json:"url"`; Key string `json:"key"`; Value string `json:"value"`; Present bool `json:"present"`; Caller string `json:"caller"`; Line int `json:"line"` }{url,key,value,present,caller,line}
  f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); if err != nil { panic(err) }
  if err := json.NewEncoder(f).Encode(row); err != nil { panic(err) }; f.Close()
 }
 return value,present
}
'''
 patched = scratch / 'query.go'; patched.write_text(text)
 control_input = scratch / 'controls.json'; control_input.write_text(json.dumps(controls))
 control_test = scratch / 'query_test.go'; control_test.write_text('''package next
import ("encoding/json"; "os"; "testing")
func TestAdamicQueryControls(t *testing.T) {
 data,err:=os.ReadFile(os.Getenv("ADAMIC_QUERY_CONTROLS")); if err!=nil {t.Fatal(err)}
 var rows []struct{URL string; Key string}; if err=json.Unmarshal(data,&rows); err!=nil {t.Fatal(err)}
 for _,row:=range rows {urlQueryValue(row.URL,row.Key)}
}
''')
 overlay = scratch / 'overlay.json'
 overlay.write_text(json.dumps({'Replace': {str(source):str(patched),str(source.parent/'adamic_query_test.go'):str(control_test)}}))
 harness = COHERE / 'internal/lint/testing/rule_testing.go'
 anchor = 'return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
 htext = harness.read_text(); assert htext.count(anchor) == 1
 hside = scratch / 'harness.go'; hside.write_text(htext.replace(anchor, 'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result'))
 mapping = json.loads(overlay.read_text()); mapping['Replace'][str(harness)] = str(hside); overlay.write_text(json.dumps(mapping))
 capture = scratch / 'calls.jsonl'
 env = os.environ | {'ADAMIC_QUERY_CAPTURE':str(capture),'ADAMIC_QUERY_CONTROLS':str(control_input),'COHERE_DOCS_CAPTURE':str(scratch/'upstream')}
 log = HERE / 'capture.log'
 with log.open('w') as output:
  subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/next','-run','^(TestGoogleFontDisplay.*|TestNoUnwantedPolyfillio.*|TestAdamicQueryControls)$','-count=1','-v','-timeout=10m'],cwd=COHERE,env=env,stdout=output,stderr=subprocess.STDOUT,check=True)
 rows = [json.loads(line) for line in capture.read_text().splitlines()]
 counts = {}
 for row in rows:
  row['caller'] = Path(row['caller']).name
  counts[row['caller']] = counts.get(row['caller'],0)+1
 assert counts.get('google_font_display.go',0)>0 and counts.get('no_unwanted_polyfillio.go',0)>0, counts
 (HERE/'calls.json').write_text(json.dumps(rows,ensure_ascii=True,indent=2)+'\n')
 (HERE/'go-output.txt').write_text(''.join(row['value']+'|'+str(row['present']).lower()+'\n' for row in rows))
 upstream = [json.loads(line) for file in (scratch/'upstream').glob('*.jsonl') for line in file.read_text().splitlines() if line]
 unique = {}
 for row in upstream:
  rule = row['rule']; unique.setdefault(rule,set()).add((row['file'],row['source'],json.dumps(row.get('options'),sort_keys=True)))
 rule_counts = {rule:{'unique':len(values),'runs':sum(r['rule']==rule for r in upstream)} for rule,values in unique.items()}
 (HERE/'upstream-counts.json').write_text(json.dumps(rule_counts,indent=2)+'\n')
 (HERE/'counts.json').write_text(json.dumps({'pin':PIN,'calls':len(rows),'byCaller':counts,'controls':len(controls)},indent=2)+'\n')
 print('captured',len(rows),'actual Go calls',counts)
