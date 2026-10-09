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
assert not before.keys() - after.keys(), 'own count rows lost'
assert not incoming.keys() - after.keys(), 'incoming count rows lost'
changes = {'added': [{'row':key,'after':value} for key,value in after.items() if key not in before],
           'changed': [{'row':key,'before':before[key],'after':value,'reason':reason} for key,value in after.items() if key in before and before[key] != value],
           'removed': [], 'rows':len(after)}
(root / (prefix + '-count-changes.json')).write_text(json.dumps(changes,indent=2)+'\n')
records = []
for path in root.glob(prefix + '*tests.jsonl'):
    for line in path.read_text().splitlines():
        try: item = json.loads(line)
        except ValueError: continue
        if item['Action'] in ('pass','fail','skip') and item.get('Test'):
            records.append({'test':item['Test'],'result':item['Action'],'seconds':item.get('Elapsed'), 'log':path.name})
assert not any(item['result']=='fail' for item in records), 'a control test failed'
(root / (prefix + '-results.json')).write_text(json.dumps({'branch':branch,'source':source,'tests':records,'counts':changes},indent=2)+'\n')
print(f'{branch}: {len(changes["added"])} count rows added, {len(changes["changed"])} changed; {len(after)} rows preserved')
