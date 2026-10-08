"""Recount raw latent observations; never equate observed sites with syntax tokens."""
import collections
import json
import pathlib
import sys

ANY = {'a value of type any', 'an array of any', 'a function returning any',
       'a call returning any', 'storing any in a field'}

def recount(raw, root):
    unique = {}
    for line in pathlib.Path(raw).read_text().splitlines():
        for original in json.loads(line).get('findings', []):
            if original.get('kind') not in ('Refused', 'NotYet'):
                continue
            row = {k: str(original.get(k, '')).replace(root + '/', '')
                   for k in ('kind', 'where', 'reason', 'text')}
            unique[tuple(row.values())] = row
    rows = list(unique.values())
    families = collections.Counter((r['kind'], r['reason']) for r in rows)
    any_rows = [r for r in rows if r['reason'] in ANY]
    reasons = []
    for (kind, reason), count in sorted(families.items(), key=lambda item: (-item[1], item[0])):
        disposition = 'compiler lesson or unresolved contract'
        if kind == 'Refused':
            if 'seen as' in reason or reason.startswith(('optional property ', 'an unproven relation ')):
                disposition = 'variance: other worker'
            elif reason.startswith('a method read as a value') or reason.endswith(' as a condition') or reason in {
                    'a namespace', 'a parameter property', 'var', 'an index signature',
                    'a generator function', 'yield (generators)', 'a definite assignment assertion !'}:
                disposition = 'source edit candidate: needs owner and behavior proof'
            elif reason in {'debugger', 'with', 'the void operator'}:
                disposition = 'adaptation 41'
            elif reason == "a cast the runtime can't check":
                disposition = 'mixed: source contract or dynamic boundary; inspect each site'
        reasons.append(dict(kind=kind, reason=reason, count=count, disposition=disposition,
                            sites=[r['where'] for r in rows if r['kind'] == kind and r['reason'] == reason]))
    return dict(measurement='measured on a checker-rejected program',
                totals=dict(collections.Counter(r['kind'] for r in rows)),
                any_count=len(any_rows), any_sites=any_rows,
                any_files=dict(sorted(collections.Counter(r['where'].rsplit(':', 2)[0] for r in any_rows).items())),
                reasons=reasons)

if __name__ == '__main__':
    report = recount(sys.argv[1], sys.argv[2])
    pathlib.Path(sys.argv[3]).write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps({k: report[k] for k in ('totals', 'any_count', 'any_files')}))
