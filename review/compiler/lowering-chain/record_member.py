"""Record serial landing observations without editing generated count values."""
from pathlib import Path
import json, sys
root = Path(__file__).resolve().parent
number, branch, source, reason = sys.argv[1:5]
prefix = 'member-' + number

def rows(path):
    result = {}
    for line in path.read_text().splitlines():
        if not line.startswith('| ') or '.a |' not in line:
            continue
        parts = [part.strip() for part in line.split('|')[1:-1]]
        if len(parts) == 7:
            result[parts[0]] = parts[1:]
    return result
before = rows(root / (prefix + '-before-counts.md'))
incoming = rows(root / (prefix + '-incoming-counts.md'))
after = rows(Path('internal/oracle/counts.md'))
retired_path = root / (prefix + '-retired-counts.json')
retired = json.loads(retired_path.read_text()) if retired_path.exists() else {}
assert before.keys() - after.keys() == retired.keys(), 'unrecorded own count rows lost'
assert not (incoming.keys() - after.keys()) - retired.keys(), 'unrecorded incoming count rows lost'
changes = {'added': [{'row':key,'after':value} for key,value in after.items() if key not in before],
           'changed': [{'row':key,'before':before[key],'after':value,'reason':reason} for key,value in after.items() if key in before and before[key] != value],
           'removed': [{'row':key,'before':before[key],'reason':value} for key,value in retired.items()], 'rows':len(after)}
(root / (prefix + '-count-changes.json')).write_text(json.dumps(changes,indent=2)+'\n')
records = []
for path in root.glob(prefix + '*tests.jsonl'):
    if path.name.endswith('-prior-tests.jsonl'):
        continue  # Preserve the initial interaction failure, count its final replay separately.
    for line in path.read_text().splitlines():
        try: item = json.loads(line)
        except ValueError: continue
        if item['Action'] in ('pass','fail','skip') and item.get('Test'):
            records.append({'test':item['Test'],'result':item['Action'],'seconds':item.get('Elapsed'), 'log':path.name})
assert not any(item['result']=='fail' for item in records), 'a control test failed'
(root / (prefix + '-results.json')).write_text(json.dumps({'branch':branch,'source':source,'tests':records,'counts':changes},indent=2)+'\n')
print(f'{branch}: {len(changes["added"])} count rows added, {len(changes["changed"])} changed; {len(after)} rows preserved')
