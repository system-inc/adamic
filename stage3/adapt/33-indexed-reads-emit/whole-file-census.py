#!/usr/bin/env python3
"""Report every checker code for one file, without the indexed-read code filter."""
import json
from pathlib import Path
import re
import sys

source = json.loads(Path(sys.argv[1]).read_text())
file = sys.argv[2]
rows = []
counts = {}
for text in source['diagnostics']:
    match = re.match(r'.*/src/compiler/(.*?):(\d+):(\d+): error TS(\d+):', text)
    if match and match[1] == file:
        code = 'TS' + match[4]
        counts[code] = counts.get(code, 0) + 1
        rows.append({'code': code, 'line': int(match[2]), 'column': int(match[3]),
                     'diagnostic': 'src/compiler/' + text[match.start(1):]})
result = {'roots': source['roots'], 'file': file, 'code_filter': None,
          'counts': counts, 'total': len(rows), 'diagnostics': rows}
Path(sys.argv[3]).write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps({'file': file, 'counts': counts, 'total': len(rows)}))
