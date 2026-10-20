#!/usr/bin/env python3
"""Extract compiler-owned void-widening reasons without interpreting their truth."""
import gzip
import json
from pathlib import Path
import re
import sys

PREFIXES = ('a void value at a return site whose function returns ',
            'a void-returning function at the ')

def extract(stream):
    opener = gzip.open if str(stream).endswith('.gz') else open
    sites = {}
    with opener(stream, 'rt') as source:
        for line in source:
            record = json.loads(line)
            for finding in (record.get('findings') or []):
                if finding.get('kind') != 'Refused' or not finding.get('reason', '').startswith(PREFIXES):
                    continue
                file, row, column = finding['where'].rsplit(':', 2)
                file = 'src/' + file.split('/src/', 1)[1]
                key = (file, int(row), int(column))
                observation = {k: finding[k] for k in ('kind', 'reason', 'text')}
                observation['text'] = re.sub(r'^.*?/src/', 'src/', observation['text'], count=1)
                observation.update(file=file, line=int(row), column=int(column))
                refusal = {k: observation[k] for k in ('kind', 'reason', 'text')}
                if key not in sites:
                    sites[key] = {**observation, 'refusals': []}
                if refusal not in sites[key]['refusals']:
                    sites[key]['refusals'].append(refusal)
    return sorted(sites.values(), key=lambda row: (row['file'], row['line'], row['column']))

if __name__ == '__main__':
    rows = extract(sys.argv[1])
    Path(sys.argv[2]).write_text(json.dumps(rows, indent=2) + '\n')
    print(json.dumps({'sites': len(rows)}))
