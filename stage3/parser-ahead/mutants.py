#!/usr/bin/env python3
"""Independent source-proof mutants. No backend acceptance credit."""
import json,pathlib,subprocess,tempfile
root=pathlib.Path(__file__).resolve().parents[2]
p=root/'internal/lower/factory_completion.go'
source=p.read_text()
mutants={
 'branch-union':('yes.written[field] && no.written[field]','yes.written[field] || no.written[field]'),
 'omit-escape-missing':('a.result.BeforeEscape[field] = false','a.result.BeforeEscape[field] = true'),
 'erase-earlier-read':('Checked: !state.written[node.Name().Text()] || state.escaped', 'Checked: false'),
 'trust-assertion':('return !forged && interfaceScalar(actual)', 'return (forged || !forged) && interfaceScalar(actual)'),
}
rows=[]
with tempfile.TemporaryDirectory(prefix='step24-mutants-') as scratch:
 for name,(old,new) in mutants.items():
  assert source.count(old)==1,(name,source.count(old))
  changed=pathlib.Path(scratch)/'changed.go.txt';changed.write_text(source.replace(old,new))
  overlay=pathlib.Path(scratch)/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(p):str(changed)}}))
  log=root/'review/compiler/step24-parser-main/logs'/('completion-'+name+'.log')
  with log.open('w') as output:
   code=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lower','-run','^TestParserFactory(Completion.*|ReadBeforeCompletion)$','-count=1','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT).returncode
  text=log.read_text()
  assert code!=0 and '--- FAIL: TestParserFactory' in text and '[build failed]' not in text,(name,code,text)
  rows.append({'mutant':name,'exit':code,'catcher':'TestParserFactoryCompletion behavior assertion','log':str(log)})
(root/'review/compiler/step24-parser-main/completion-mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))
