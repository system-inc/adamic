"""Reconcile receiver classifications with the pinned raw root CSV."""
import csv
import io
import json
import subprocess
from pathlib import Path

table = 'origin/codex/stage3-notyet-table'
raw = subprocess.check_output(['git', 'show', table + ':stage3/notyet-table/roots/raw.csv'], text=True)
roots = {r['where'] for r in csv.DictReader(io.StringIO(raw)) if r['reason'] == 'an ElementAccessExpression' and not r['blocked_by_reason']}
assert len(roots) == 91, len(roots)
classified = {r['site']:r for r in csv.DictReader(Path('stage3/notyet-element-access/sites.csv').open()) if r['site'] in roots}
assert set(classified) == roots
counts = {}
rows = []
for site in sorted(roots):
 r = classified[site]
 receiver = r['receiver']
 if any(name in receiver for name in ('NodeArray', 'Readonly<PathPathComponents>', 'TemplateStringsArray', 'SortedReadonlyArray', 'JSDocArray')):
  group, covered = 'array-derived', False
 elif receiver.startswith('readonly ['):
  group, covered = 'fixed-tuple-union', True
 elif receiver.startswith(('Record<number', 'MapLike', 'CompilerOptions')):
  group, covered = 'open-dictionary', False
 else:
  group, covered = 'finite-data-fields', True
 counts[group] = counts.get(group,0) + 1
 rows.append({'site':site, 'receiver':receiver, 'key':r['key'], 'group':group, 'rule_shape_covered':covered})
with Path('stage3/notyet-element-access/resumed-coverage.csv').open('w') as output:
 writer = csv.DictWriter(output, fieldnames=list(rows[0]))
 writer.writeheader(); writer.writerows(rows)
summary = {'table_sha':subprocess.check_output(['git','rev-parse',table],text=True).strip(), 'raw_csv':'stage3/notyet-table/roots/raw.csv', 'root_sites':len(roots), 'groups':counts, 'rule_shape_covered':sum(r['rule_shape_covered'] for r in rows), 'limit':'Rule shape coverage is not a count of complete units compiling or all sites replayed successfully.'}
Path('stage3/notyet-element-access/resumed-coverage.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary))
