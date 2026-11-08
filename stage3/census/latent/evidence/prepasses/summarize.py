"""Summarize the pair tool's exact identities without renormalizing observations."""
from collections import Counter
import json
from pathlib import Path
import sys

folder = Path(sys.argv[1])
lines = []
summary = {}
for label in ['first-error', 'core-full']:
    pair = json.loads((folder / (label+'-paired.json')).read_text())
    lines += ['\n## '+label, '', '| View | Cleared | New | Changed | Newly stopped units | Changed existing first stops | Panics before | Panics after |',
              '| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |']
    summary[label] = {}
    groups = {}
    for view in ['raw_findings','distinct_reads']:
        data = pair[view]
        row = dict(cleared=len(data['lowered']), new=len(data['new_findings']), changed=len(data['moved']),
                   newly_stopped=sum(not r['base'] and bool(r['head']) for r in data['unit_first_stop_delta']),
                   changed_existing_stops=sum(bool(r['base']) and bool(r['head']) for r in data['unit_first_stop_delta']),
                   panics_before=len(data['panics']['before']), panics_after=len(data['panics']['after']),
                   totals=data['totals'])
        summary[label][view] = row
        lines.append('| '+view+' | '+' | '.join(str(row[k]) for k in ['cleared','new','changed','newly_stopped','changed_existing_stops','panics_before','panics_after'])+' |')
        groups[view] = Counter((r['finding']['reason'],r['identity'][0]) for r in data['new_findings'])
    lines += ['', 'New findings by exact reason and source file:', '', '| Reason | File | Raw | Distinct |', '| --- | --- | ---: | ---: |']
    for reason, file in sorted(set(groups['raw_findings']) | set(groups['distinct_reads'])):
        key = (reason,file)
        lines.append('| '+reason.replace('|','\\|')+' | '+file+' | '+str(groups['raw_findings'][key])+' | '+str(groups['distinct_reads'][key])+' |')
    if not any(groups.values()):
        lines.append('| None | | 0 | 0 |')
    lines += ['', 'Cleared is an old finding with no remaining reason at that identity. Changed retains another reason at the identity. New includes replacement reasons. A cleared first-error observation can be hidden by an earlier new stop. Only lowering-phase findings define unit first stops. Raw and distinct are unique identities, not event multiplicity.']
(folder/'counts.json').write_text(json.dumps(summary,indent=2)+'\n')
(folder/'counts.md').write_text('\n'.join(lines)+'\n')
