"""Conservative, ordered attribution of pinned diagnostic reasons to wall steps."""
import hashlib
import json
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parent
PIN = 'ec0b16c0'
STEPS = {11: 'branded types', 15: 'namespaces', 16: 'generics',
         17: 'counted and uncounted unions', 18: 'optional calls',
         19: 'Maps and Sets with any key', 20: 'iteration', 21: 'exceptions',
         22: 'object semantics', 23: 'regex', 30: 'long tail'}
# First matching rule wins. These describe the diagnostic operation, not identifiers
# in tsc's program (reading excludeRegex is not a regex implementation failure).
RULES = [
    ('object-relation', 22, r'^a value of type .* seen as |^an unproven relation .*optional field'),
    ('generic-mapper', 16, r'^a generic function|^overload .*additional implementation type parameters'),
    ('overload', 30, r'^overload |^an overloaded function as a value'),
    ('object-operation', 22, r'^a method call through a structural signature in a program with statics;|^an assignment value to a member$|^Object\.|^hasOwnProperty |^a computed field name$|^a method in object destructuring$|^replacing a represented method|^a constructor object escaping|^writing a possibly absent optional own field|^in when |^a DeleteExpression|^a spread after |^a spread that adds |^inherited library member|^a checked view with callable field|^a literal method through a view|^a class method through a view|^an uninitialized object field|^a structural Object.keys'),
    ('iteration', 20, r'^for\.\.\.(?:of|in) |^a for\.\.\.of |^iterating a value$'),
    ('optional-call', 18, r'^a call through \?\. \(an optional call\)$'),
    ('collection-storage', 19, r'^a Map |^a Set |^new Map '),
    ('regex-operation', 23, r'^RegExp '),
    ('exception-operation', 21, r'^(?:throw|try|catch|finally)(?: |$)'),
    ('namespace-operation', 15, r'^namespace(?: |$)'),
    ('generic-type', 16, r'^(?:a value of type |a function returning |a call returning |an array of ).*\b(?:T|U|V|K|R|X|Y|T[A-Z]\w*)\b'),
    ('known-brand', 11, r'^(?:a value of type |a function returning |a call returning |an array of )(?:__String|Path|ResolvedConfigFileName|ResolvedConfigFilePath|PathPathComponents)(?:$| \| undefined$)'),
    ('union-storage', 17, r'^a narrowed .*union|^a narrowed scalar in a boxed union field|^a union of differently held|^null comparison with a scalar$|^(?:a value of type |a function returning |a call returning |an array of |checked view field .* of type ).* \| |^a BinaryExpression with .*union of differently held'),
    ('object-view', 22, r'^a function viewed as unknown or object|^an object with a nullable field viewed as unknown or object|^JSON.stringify object references'),
    ('other-explicit-operation', 30, r'^var$|^a cast the runtime can.t check$|^a destructured |^destructuring |^a parameter that isn.t|^a rest parameter|^a SpreadElement$|^spreading |^a tagged template|^a template interpolating|^toLocaleTimeString:|^JSON.parse:|^a ClassExpression$|^this outside|^a block-scoped nested function|^a YieldExpression|^a (?:PostfixUnaryExpression|PrefixUnaryExpression|BinaryExpression|ConditionalExpression)|^returning a property or element assignment|^incrementing a NonNullExpression|^new (?:an Identifier|a ParenthesizedExpression)|^(?:push|sort|join|slice|lastIndexOf|indexOf|findIndex|apply) |^a (?:filter|some) callback|^a comparator|^typed array element type|^Array as a value'),
]


def placement(row):
    if row['kind'] not in ('Refused', 'NotYet'):
        return None, 'measurement-cause'
    for name, step, pattern in RULES:
        if re.search(pattern, row['reason']):
            return step, name
    return None, 'insufficient-diagnostic'


def aggregate(rows, mapping):
    source = {}
    for row in rows:
        key = (row['kind'], row['reason'])
        if key in source:
            raise ValueError('duplicate input reason: ' + repr(key))
        value = row['bytes_revealed_if_fixed_alone']
        if type(value) is not int or value < 0:
            raise ValueError('invalid credited bytes')
        source[key] = row
    assigned = {}
    for item in mapping:
        key = (item['kind'], item['reason'])
        if key in assigned:
            raise ValueError('reason mapped more than once: ' + repr(key))
        if item['step'] is not None and item['step'] not in STEPS:
            raise ValueError('unknown roadmap step')
        assigned[key] = item['step']
    if source.keys() != assigned.keys():
        raise ValueError('mapping must cover exactly the input reasons')
    groups = []
    for step in [*STEPS, None]:
        reasons = [r for k, r in source.items() if assigned[k] == step]
        reasons.sort(key=lambda r: (-r['bytes_revealed_if_fixed_alone'], r['kind'], r['reason']))
        groups.append({'step': step, 'title': STEPS.get(step, 'unplaced'),
                       'hidden_bytes': sum(r['bytes_revealed_if_fixed_alone'] for r in reasons),
                       'reason_count': len(reasons), 'top_three': reasons[:3]})
    groups.sort(key=lambda g: (-g['hidden_bytes'], g['step'] if g['step'] is not None else 999))
    return groups


def write_json(name, value):
    (ROOT / name).write_text(json.dumps(value, indent=2) + '\n')


def main():
    raw = subprocess.check_output(['git', 'show', PIN + ':stage3/census/hidden-ranking/RESULT.json'])
    source = json.loads(raw)
    rows = source['ranked_reasons'] + source['other_causes']
    # All input records are parsed; only already credited reason rows are projected.
    mapping = []
    for row in rows:
        step, rule = placement(row)
        mapping.append({'kind': row['kind'], 'reason': row['reason'], 'step': step, 'rule': rule})
    groups = aggregate(rows, mapping)
    total = sum(g['hidden_bytes'] for g in groups)
    if total != source['hidden_bytes']:
        raise ValueError('credited reasons do not conserve hidden bytes')
    write_json('input.json', {'hidden_bytes': source['hidden_bytes'], 'reasons': rows,
                            'source_commit': subprocess.check_output(['git', 'rev-parse', PIN], text=True).strip(),
                            'source_sha256': hashlib.sha256(raw).hexdigest(),
                            'compiler_commit': source['provenance']['compiler_commit']})
    write_json('mapping.json', mapping)
    write_json('RESULT.json', {'hidden_bytes': total, 'ranked_groups': groups})
    lines = ['# Hidden bytes by roadmap step', '',
             'Generated from the pinned ranking. Unplaced is included to conserve every credited byte.', '',
             '| Rank | Step | Title | Hidden bytes | Reasons |',
             '|---:|---:|---|---:|---:|']
    for i, g in enumerate(groups, 1):
        lines.append(f"| {i} | {g['step'] if g['step'] is not None else 'unplaced'} | {g['title']} | {g['hidden_bytes']:,} | {g['reason_count']} |")
    for g in groups:
        lines.extend(['', f"## {g['step'] if g['step'] is not None else 'Unplaced'}: {g['title']}", ''])
        for r in g['top_three']:
            lines.append(f"- {r['bytes_revealed_if_fixed_alone']:,} bytes: {r['kind']}: {r['reason']}")
        if not g['top_three']:
            lines.append('No matching diagnostic reasons; zero is not evidence that this feature compiles.')
    lines.extend(['', '## Every unplaced reason', '', '| Kind | Exact reason | Bytes | Rule |', '|---|---|---:|---|'])
    indexed = {(m['kind'], m['reason']): m for m in mapping}
    for r in rows:
        m = indexed[r['kind'], r['reason']]
        if m['step'] is None:
            reason = r['reason'].replace('|', '&#124;').replace('\n', '<br>')
            lines.append(f"| {r['kind']} | {reason} | {r['bytes_revealed_if_fixed_alone']:,} | {m['rule']} |")
    (ROOT / 'REPORT.md').write_text('\n'.join(lines) + '\n')
    print(f'{len(rows)} reasons, {total} bytes conserved')
    for g in groups:
        print(g['step'], g['title'], g['hidden_bytes'], g['reason_count'])


if __name__ == '__main__':
    main()
