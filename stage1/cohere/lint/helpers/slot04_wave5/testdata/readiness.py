"""Derive residual dependencies without changing the frozen original ledger."""
import json
from pathlib import Path
HERE = Path(__file__).resolve().parent
BASE = HERE.parents[1]
prefix = 'github.com/system-inc/cohere/internal/lint/'
current = [prefix + 'rules/tailwind/collapse.' + name for name in ['Comment', 'Declaration', 'breakpointBucket']]
prior = [prefix + name for name in ['ecmascript/jsx.ElementParts', 'rules/tailwind.ClassLiteralSettings.key', 'rules/tailwind.compiledClassLiteralReader', 'rules/tailwind.classValuesUnder', 'rules/tailwind.collectClassValues', 'rules/tailwind/collapse.isJavaScriptSpace', 'rules/tailwind/collapse.isBlank', 'rules/tailwind/collapse.isValueSeparator', 'rules/tailwind/collapse.topOfStack', 'rules/tailwind/collapse.peekByte', 'rules/tailwind/collapse.sortedKeys']]
base = json.loads((BASE / 'readiness.json').read_text())
helpers = [{'symbol': symbol, 'consumers': [r['rule'] for r in base['remaining'] if symbol in r['remaining_helpers']], 'alone_final_blockers_removed': [r['rule'] for r in base['remaining'] if r['remaining_helpers'] == [symbol]]} for symbol in current]
remaining, new_ready, cumulative_ready = [], [], []
for row in base['remaining']:
    before = [s for s in row['remaining_helpers'] if s not in prior]
    after = [s for s in before if s not in current]
    result = dict(row, remaining_helpers=after, helper_ready=not after)
    remaining.append(result)
    if before and not after:
        new_ready.append(row['rule'])
    if not row['helper_ready'] and not after:
        cumulative_ready.append(row['rule'])
result = {'base_helper_ready': base['helper_ready'], 'new_ready': new_ready, 'cumulative_new_ready': cumulative_ready, 'cumulative_helper_ready': base['helper_ready'] + len(cumulative_ready), 'removed_dependency_edges': sum(len(h['consumers']) for h in helpers), 'distinct_consumers': len({r for h in helpers for r in h['consumers']}), 'helpers': helpers, 'remaining': remaining}
(HERE.parent / 'readiness.json').write_text(json.dumps(result, indent=2) + '\n')
print('edges', result['removed_dependency_edges'], 'consumers', result['distinct_consumers'], 'new ready', new_ready, 'cumulative helper ready', result['cumulative_helper_ready'])
