"""Check rule.json against production listener kinds and the private driver."""
from pathlib import Path
import json,re
ROOT=Path(__file__).resolve().parents[4];HERE=Path(__file__).resolve().parent
values={name:int(value) for name,value in re.findall(r'x\[Kind(\w+)-(\d+)\]',(ROOT/'cohere/TypeScript/tsc/internal/ast/kind_stringer_generated.go').read_text())}
driver=(HERE/'main.a').read_text()
for module,callback in [('require_await','requireAwait'),('symbol_description','symbolDescription'),('valid_typeof','validTypeof')]:
    source=(ROOT/'cohere/internal/lint/rules/core'/(module+'.go')).read_text().split('return rule.Listeners{',1)[1]
    kinds=re.findall(r'^\t{3}ast.Kind(\w+)\s*:',source,re.M)
    expected=[values[name] for name in kinds]
    actual=json.loads((HERE/module/'rule.json').read_text())['kinds']
    registered=[values[name] for name in re.findall(r'\[(\w+), \['+callback+r'\]\]',driver)]
    assert actual==expected and registered==actual,(module,actual,expected,registered)
    rule=(HERE/module/'rule.a').read_text()
    assert not re.search(r'\.kind\s*(?:===|!==)\s*[\x22\x27]',rule),module
    mutant=actual.copy();mutant[0]+=1;assert mutant!=expected
    print(module,actual,'matches Go and driver; numeric metadata mutant caught')
print('PASS three kind-indexed callback declarations; no string-kind comparisons in rule files')
