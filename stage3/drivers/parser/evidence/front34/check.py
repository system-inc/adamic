from pathlib import Path
import json
import gzip

root=Path(__file__).resolve().parent
r=json.loads((root/'parser-front34-minimals/report.json').read_text())
assert r['passing']==8 and r['total']==21
probes={p['file']:p for p in r['probes']}
for p in probes.values():
    assert p['node']['exit']==0
    if p['build']['exit']==0:
        assert p['pass'] and p['native']['exit']==0
        assert p['node']['stdout']==p['native']['stdout']
        assert p['node']['stderr']==p['native']['stderr']==''
        assert p['native_byte_mutant_cmp']==1
        assert all(value==0 for value in p['comparison'].values())
for name in ['native-namespace-object-receiver.a','native-namespace-class.a','native-callable-namespace.a','native-predicate-overload.a','native-predicate-callback.a','native-arguments-length.a','native-arguments-length-value.a']:
    assert probes[name]['pass']
for name in ['native-error-capture-stack.a','native-enum-map.a','native-call-before-enum.a','native-nonnull.a','native-predicate-callback-parameter.a','native-debug-namespace.a','native-map-before-namespace.a']:
    assert probes[name]['build']['exit']!=0
assert 'core.ts:11:52:' in (root/'parser-front34-behind-error.stderr').read_text()
metrics=json.loads((root/'parser-front34-build-modes/report.json').read_text())
assert metrics['generated_c'] is None and metrics['emit_exit']!=0
for name in ['unsplit','split']:
    with gzip.open(root/f'parser-front34-build-modes/{name}.stderr.gz','rt') as f:text=f.read()
    assert 'debug.ts:113:19:' in text and 'debug.ts:114:19:' in text
for name in ['parser-front34-full-node','parser-front34-slice-node']:
    report=json.loads((root/name/'report.json').read_text())
    assert report['node']['bytes']==36429231 and report['node']['sha256']=='686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615'
    assert report['node']['stderr_bytes']==0 and report['mutation']['changed_lines']==1
    mutants=json.loads((root/name/'jsdoc-mutants/report.json').read_text())
    assert all(row['comparison_exit']==1 for row in mutants.values())
print(json.dumps({'native_matching':8,'byte_mutants_caught':8,'namespace_declaration_stops_cleared':3,'enum_stop_confirmed_behind_throw':True,'parser_native':'checker red','extended_node_identity':True}))
