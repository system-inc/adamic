"""Recount the panic-fix delta against the previous committed census evidence."""
import gzip, json, pathlib, subprocess
root = pathlib.Path(__file__).resolve().parent.parent
before_sha = 'a09ca4eac49fe58e3495a9587fce8e0b352d69e0'
def previous(path):
    return subprocess.check_output(['git', 'show', before_sha + ':stage3/census/predicates/' + path], cwd=root)
before = json.loads(previous('REPORT.json'))
after = json.loads((root / 'REPORT.json').read_text())
assert before['source'] == after['source'], 'source bytes changed'
assert before['measurement'] == after['measurement']
key = lambda finding: (finding['kind'], finding['where'], finding['reason'], finding['text'])
old_runs = {run['name']: run for run in before['runs']}
for run in after['runs'][:6]:
    assert {key(f) for f in run['findings']} == {key(f) for f in old_runs[run['name']]['findings']}, run['name']
old = old_runs['predicates']
new = next(run for run in after['runs'] if run['name'] == 'predicates')
assert old['lowering_counts']['panic'] == 3
assert new['lowering_counts']['panic'] == 0, 'compiler panic remains'
assert new['features']['codex/proven-predicates'] == '04b3f073992d6712bfb850677def5d47673fbaea'
assert new['checker_total'] == old['checker_total']
for name in ['total', 'functions', 'skipped_checker_body', 'attempted_functions', 'attempted_statements']:
    assert new['units'][name] == old['units'][name], name
with gzip.open(root / 'data/predicates.jsonl.gz', 'rt') as source:
    raw = [json.loads(line) for line in source]
unit = 'src/compiler/emitter.ts:685:1'
expected = 'instantiating a generic function makes a value of type string | undefined written where string is read'
assert any(f['unit'] == unit and f['phase'] == 'lowering' and f['kind'] == 'Refused' and f['where'] == 'src/compiler/core.ts:288:33' and f['reason'] == expected for row in raw[1:] for f in row['findings'])
old_sites = {key(f) for f in old['findings']}
new_sites = {key(f) for f in new['findings']}
old_families = json.loads(previous('FAMILIES.json'))
new_families = json.loads((root / 'FAMILIES.json').read_text())
def finding(site):
    kind, where, reason, text = site
    return dict(kind=kind, where=where, reason=reason, text=text)
result = dict(measurement=after['measurement'], before_census=before_sha,
    fixed_compiler=new['features']['codex/proven-predicates'], source_unchanged=True,
    original_six_finding_sets_unchanged=True,
    before_counts=old['lowering_counts'], after_counts=new['lowering_counts'],
    before_family_counts=old_families['counts']['predicates'], after_family_counts=new_families['counts']['predicates'],
    added=[finding(site) for site in sorted(new_sites-old_sites)],
    removed=[finding(site) for site in sorted(old_sites-new_sites)])
(root / 'PANIC_DELTA.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps({k: result[k] for k in ['before_counts', 'after_counts', 'before_family_counts', 'after_family_counts']}, indent=2))
print('panic delta audit passed: original six baselines unchanged, source and eligibility unchanged, three panics removed, emitter refusal preserved')
