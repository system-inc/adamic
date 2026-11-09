"""Summarize the paired census without changing the pair tool's identities."""
from collections import Counter
import json
from pathlib import Path
import sys

folder = Path(sys.argv[1])
pair = json.loads((folder / 'paired.json').read_text())
raw = pair['raw_findings']
newly = [row for row in raw['unit_first_stop_delta'] if not row['base'] and row['head']]
changed = [row for row in raw['unit_first_stop_delta'] if row['base'] and row['head']]
groups = {}
for view in ('raw_findings', 'distinct_reads'):
    groups[view] = Counter((row['finding']['reason'], row['identity'][0]) for row in pair[view]['new_findings'])
summary = dict(lowered_to_stopped=len(newly), changed_existing_stops=len(changed),
               units=[dict(unit=row['unit'], stop=row['head']) for row in newly],
               new_findings_by_reason_and_file=[dict(reason=reason, file=file,
                   raw=groups['raw_findings'][reason,file], distinct=groups['distinct_reads'][reason,file])
                   for reason,file in sorted(set(groups['raw_findings']) | set(groups['distinct_reads']))],
               production_root_first_stops_moved=0, production_roots=73,
               production_first_stop='core.ts:2599:15',
               production_note='Production lower and refusal sources are byte-identical to 018b7c21. The 73 production roots were not recompiled; their first stops cannot change in this measurement-only overlay change.')
(folder / 'counts.json').write_text(json.dumps(summary, indent=2)+'\n')
lines = ['# Deferred refusal census', '',
    f"{len(newly)} units move from lowered to stopped; {len(changed)} existing unit first stops change.", '',
    '| Unit | Deferred reason |', '| --- | --- |']
for row in newly:
    lines.append('| '+row['unit']['where']+' | '+row['head']['reason'].replace('|','\\|')+' |')
lines += ['', 'New unique findings by reason and file:', '',
          '| Reason | File | Raw | Distinct |', '| --- | --- | ---: | ---: |']
for row in summary['new_findings_by_reason_and_file']:
    lines.append('| '+row['reason'].replace('|','\\|')+' | '+row['file']+' | '+str(row['raw'])+' | '+str(row['distinct'])+' |')
lines += ['', 'Production first stops: 0 of 73 move from `core.ts:2599:15`. '+summary['production_note'], '',
          'Lowered here means an attempted unit with no first stop in phases lowering or deferred. These census observations do not establish successful compilation. The new deferred phase participates in the pair tool first-stop comparison. Module-owned deferred findings are retained as findings but are not independent units.']
(folder / 'counts.md').write_text('\n'.join(lines)+'\n')
