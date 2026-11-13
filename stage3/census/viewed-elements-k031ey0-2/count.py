import collections
import itertools
import json
from pathlib import Path
import re
root = Path(__file__).resolve().parent
pattern = re.compile(r'^viewed-elements: build=(main|head) every-on-entry=(true|false) every=(true|false) viewed=(\d+) trip=(".*")$')
result = {}
for side in ('main', 'head'):
    lines = [line for line in (root / (side + '.stderr')).read_text().splitlines() if line.startswith('viewed-elements:')]
    counts = collections.Counter()
    for line in lines:
        match = pattern.fullmatch(line)
        if not match or match[1] != side:
            raise ValueError(line)
        counts[(match[2], match[3], match[5] != '""')] += 1
    result[side] = {'total': len(lines), 'classes': [dict(every_on_entry=entry, every=every, trip_nonempty=trip, count=counts[(entry, every, trip)]) for entry, every, trip in itertools.product(('false', 'true'), ('false', 'true'), (False, True))], 'lines': lines}
print(json.dumps(result, indent=2))
