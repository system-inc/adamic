#!/usr/bin/env python3
"""Read-only queue audit: fresh origin refs, helper handoff, then syntax inventory."""
import json
import re
import subprocess
from pathlib import Path

repository = Path(__file__).resolve().parents[5]
def git(*args):
    return subprocess.check_output(['git', *args], cwd=repository, text=True)

def show(ref, path):
    return git('show', ref + ':' + path)

inventory = json.loads(show('origin/codex/lint-inventory', 'stage1/cohere/lint/inventory/inventory.json'))
report = show('origin/codex/lint-helpers', 'stage1/cohere/lint/helpers/REPORT.md')
assert '../HELPERS.md' in report
handoff = show('origin/codex/lint-helpers', 'stage1/cohere/lint/HELPERS.md')
section = handoff.split('### Options complete')[1].split('### Remaining option gaps')[0]
helper_names = re.findall(r'^- `([^`]+)`', section, re.M)
assert len(helper_names) == 46
refs = git('for-each-ref', '--format=%(refname:short)', 'refs/remotes/origin').splitlines()
claims = []
for ref in refs:
    for file in git('ls-tree', '-r', '--name-only', ref, 'stage1/cohere/lint/claims').splitlines():
        if file.endswith('.md'):
            claims.append((ref, file, show(ref, file)))
files = git('ls-tree', '-r', '--name-only', 'origin/main', 'stage1/cohere/lint').splitlines()
files = [file for file in files if file.endswith(('.a', '.ts', 'rule.json'))
         and all(part not in file for part in ['/inventory/', '/helpers/', '/testdata/'])]
main = '\n'.join(show('origin/main', file) for file in files)
queue = helper_names + [rule['name'] for rule in inventory['rules']
                       if not rule['needs_type_information'] and not rule['binding_only']
                       and rule['name'] not in helper_names]
rows = []
for name in queue:
    matches = sorted(set((ref, file) for ref, file, text in claims
                         if re.search(r'(?<![\w/-])' + re.escape(name) + r'(?![\w/-])', text)))
    # The slot's appended claim is intentionally ignored for its pre-code selection replay.
    matches = [(ref, file) for ref, file in matches
               if not (ref == 'origin/codex/lint-wave1-09' and file.endswith(('wave1-09.md', 'wave1-09-next-report.md'))
                       and name in ['structure/tailwind-no-physical-direction',
                                    '@eslint-community/eslint-comments/require-description',
                                    '@next/next/google-font-display'])]
    rows.append({'name': name, 'helper_ready': name in helper_names,
                 'ported_on_main': re.search(r"['\"]" + re.escape(name) + r"['\"]", main) is not None,
                 'claims': matches})
selected = [row['name'] for row in rows if not row['ported_on_main'] and not row['claims']][:3]
print(json.dumps({'refs': {ref: git('rev-parse', ref).strip() for ref in refs},
                  'main_source_files': files, 'claim_documents_scanned': len(claims),
                  'helper_count': len(helper_names), 'syntax_count': len(queue),
                  'selected': selected, 'queue': rows}, indent=2))
