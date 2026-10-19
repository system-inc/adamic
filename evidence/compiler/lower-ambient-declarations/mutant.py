from pathlib import Path
import subprocess,json
p=Path('internal/lower/ambient.go')
original=p.read_text()
try:
 s=original.replace('func (l *lowering) ambientRefusal(node *ast.Node) error {','func (l *lowering) ambientRefusal(node *ast.Node) error {\n return nil')
 s=s.replace('func (l *lowering) referenceRefusal(module *ast.SourceFile) error {','func (l *lowering) referenceRefusal(module *ast.SourceFile) error {\n return nil')
 p.write_text(s)
 with open('/tmp/adamic-ambient-evidence/mutant.jsonl','w') as log:
  result=subprocess.run(['go','test','./internal/oracle','-run','^TestAmbientDeclarationsAreRefused$','-count=1','-timeout','30m','-json'],stdout=log,stderr=subprocess.STDOUT)
 events=[json.loads(s) for s in Path('/tmp/adamic-ambient-evidence/mutant.jsonl').read_text().splitlines() if s.startswith('{')]
 failed=sorted(e['Test'].split('/',1)[1] for e in events if e.get('Action')=='fail' and e.get('Test','').startswith('TestAmbientDeclarationsAreRefused/'))
 expected=sorted(p.name for p in Path('internal/oracle/testdata/ambient_refused').iterdir() if p.suffix in ('.a','.ts') and not p.name.endswith('.d.ts'))
 print('mutant exit:',result.returncode,'failed fixtures:',len(failed),'expected:',len(expected))
 print('\n'.join(failed))
 if result.returncode==0 or failed!=expected: raise SystemExit('mutant did not fail every fixture')
finally:
 p.write_text(original)
 print('restored ambient.go')
