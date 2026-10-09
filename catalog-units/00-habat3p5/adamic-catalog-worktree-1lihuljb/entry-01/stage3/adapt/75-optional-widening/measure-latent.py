#!/usr/bin/env python3
"""Count the refusal rule's exact reason, rather than optional syntax or paths."""
import collections
import json
from pathlib import Path
import sys
before, after, output = map(Path, sys.argv[1:])
result = {}
remaining = []
for name, file in [('before', before), ('after', after)]:
    rows = [json.loads(line) for line in file.read_text().splitlines()]
    findings = [finding for row in rows for finding in row.get('findings', [])
                if finding.get('reason', '').startswith('optional property ')
                and 'absent from structural source' in finding['reason']]
    result[name] = {'optional_widening_lowering_sites': len(findings),
                    'measurement': 'checker-rejected program',
                    'phases': dict(collections.Counter(f['phase'] for f in findings))}
    if name == 'after':
        remaining = findings
output.mkdir(parents=True, exist_ok=True)
(output / 'latent-counts.json').write_text(json.dumps(result, indent=2) + '\n')
(output / 'latent-remaining-sites.json').write_text(json.dumps(remaining, indent=2) + '\n')
print(json.dumps(result))
