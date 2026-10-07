"""Derive only this slot's dependency removals from the frozen ledger."""
import json
from pathlib import Path
HERE = Path(__file__).resolve().parent
BASE = HERE.parents[1]
ledger = json.loads((BASE / 'readiness.json').read_text())
symbols = {
    'github.com/system-inc/cohere/internal/lint/ecmascript/jsx.ElementParts': 'jsx_element_parts.a',
    'github.com/system-inc/cohere/internal/lint/rules/tailwind.ClassLiteralSettings.key': 'tailwind_settings_key.a',
    'github.com/system-inc/cohere/internal/lint/rules/tailwind.compiledClassLiteralReader': 'tailwind_compiled_reader.a',
}
helpers = []
for symbol, file in symbols.items():
    consumers = [r['rule'] for r in ledger['remaining'] if symbol in r['remaining_helpers']]
    final = [r['rule'] for r in ledger['remaining'] if r['remaining_helpers'] == [symbol]]
    helpers.append({'symbol': symbol, 'file': file, 'consumers': consumers, 'alone_final_blockers_removed': final})
remaining = []
new_ready = []
for row in ledger['remaining']:
    revised = dict(row)
    revised['remaining_helpers'] = [h for h in row['remaining_helpers'] if h not in symbols]
    revised['helper_ready'] = len(revised['remaining_helpers']) == 0
    if not row['helper_ready'] and revised['helper_ready']: new_ready.append(row['rule'])
    remaining.append(revised)
result = {
    'cohort': ledger['cohort'],
    'base_helper_ready': ledger['helper_ready'],
    'slot04_helper_ready': ledger['helper_ready'] + len(new_ready),
    'new_helper_ready': new_ready,
    'removed_dependency_edges': sum(len(h['consumers']) for h in helpers),
    'distinct_consumers': len({r for h in helpers for r in h['consumers']}),
    'helpers': helpers,
    'remaining': remaining,
    'limits': ['Dependency arithmetic, not completed lint rules.', 'Requires the documented common AST adapter.', 'Compiled cache requires one shared instance and a Go-compatible reader factory; concurrent access is not implemented.'],
}
(BASE / 'slot04_readiness.json').write_text(json.dumps(result, indent=2)+'\n')
print('helpers',len(helpers),'dependency edges',result['removed_dependency_edges'],'distinct consumers',result['distinct_consumers'],'new ready',new_ready)
