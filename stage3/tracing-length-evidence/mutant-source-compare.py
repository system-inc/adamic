from pathlib import Path
from run import raw,root
import json
source=Path('/workspace/scratch/tracing-length-after');pins=json.loads((root/'stage3/adapt/99-correctness-tracing-length/pins.json').read_text());files=sorted(str(p.relative_to(source/'src')) for p in (source/'src').rglob('*') if p.is_file());summary={}
for pin in pins:
 owner=pin['owner'];tree=raw/('mutant-'+owner);assert files==sorted(str(p.relative_to(tree/'src')) for p in (tree/'src').rglob('*') if p.is_file())
 old,new={'pop':('eventStack.splice(eventStack.length - 1);','eventStack.length--;'),'popAll':('eventStack.splice(0);','eventStack.length = 0;')}[owner]
 original=(source/'src/compiler/tracing.ts').read_bytes().decode();assert pin['after'].count(old)==1
 expected=original.replace(pin['after'],pin['after'].replace(old,new));assert (tree/'src/compiler/tracing.ts').read_bytes()==expected.encode()
 changed=[n for n in files if (source/'src'/n).read_bytes()!=(tree/'src'/n).read_bytes()];assert changed==['compiler/tracing.ts']
 assert 'if (eventStack.length === 0) throw new RangeError("Invalid array length");' in expected
 summary[owner]={'sourceFiles':len(files),'onlyChangedFile':changed[0],'onlyLiteralRestored':{'from':old,'to':new},'rangeErrorGuardRetained':True}
(raw/'mutant-source-comparison-report.json').write_text(json.dumps(summary,indent=2)+'\n');print('both restoring mutants differ only by their own length-write literal; guard retained')
