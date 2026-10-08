#!/usr/bin/env python3
"""Join pinned source spans, never checker IDs from different census runs."""
import gzip
import json
from pathlib import Path
root = Path(__file__).resolve().parents[3]
destination = Path(__file__).resolve().parent
lane2 = json.loads((root / 'stage3/interface-downcasts/lane2/read-census/pairs.json').read_text())
lane4 = json.loads(gzip.decompress((destination / 'lane4-read-demand-sites.json.gz').read_bytes()))
key = lambda site: (site['file'], site['start'], site['end'])
a = {key(s): s for p in lane2 if 'callable contracts' in p['families'] for s in p['sites']}
b = {key(s): s for s in lane4 if 'callable contracts' in s['families']}
result = dict(lane2_reads=len(a), lane4_reads=len(b), shared_spans=len(a.keys() & b.keys()),
              lane2_only_spans=len(a.keys() - b.keys()), lane4_only_spans=len(b.keys() - a.keys()),
              union_spans=len(a.keys() | b.keys()),
              lane2_only_examples=[a[k] for k in sorted(a.keys() - b.keys())[:8]],
              lane4_only_examples=[b[k] for k in sorted(b.keys() - a.keys())[:8]])
assert (len(a), len(b), result['shared_spans'], result['union_spans']) == (1503, 11063, 918, 11648)
(destination / 'census-reconciliation.json').write_text(json.dumps(result, indent=2) + '\n')
pairs = json.loads(gzip.decompress((destination / 'lane4-read-demand-pairs.json.gz').read_bytes()))
pairs = [p for p in pairs if 'callable contracts' in p['families']]
pairs.sort(key=lambda p: (-p['reads'], p['type'], p['field']))
assert len(pairs) == 2818 and sum(p['reads'] for p in pairs) == 11063
(destination / 'unknown-callable-pairs-ranked.json').write_text(json.dumps([
    dict(p, rank=i, certified=False) for i, p in enumerate(pairs, 1)], indent=2) + '\n')
print('308 lane2 pairs; 2818 Unknown pairs; 11648 deduplicated explicit reads; zero certified')
