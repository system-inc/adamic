"""Read pinned exception findings without treating AST node names as runtime errors."""
import gzip
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[2]
PIN = '6c4fc1af'
PREFIXES = ('throwing a ', 'throwing an Error ', 'a catch that ', 'new Error with ', 'a try around ', 'assigning to what a catch caught')


def belongs(reason):
    return reason.startswith(PREFIXES)


def report():
    ranking = json.loads(subprocess.check_output([
        'git', 'show', f'{PIN}:stage3/census/hidden-ranking/RESULT.json'], cwd=ROOT))
    reasons = [r for r in ranking['ranked_reasons'] if belongs(r['reason'])]
    roots = {}
    path = ROOT / 'stage3/meter/runs/20261008T035244Z.latent-full/compiler/full.jsonl.gz'
    with gzip.open(path, 'rt') as source:
        for line in source:
            row = json.loads(line)
            for finding in row.get('findings', []):
                if belongs(finding.get('reason', '')):
                    key = (finding['kind'], finding['reason'])
                    roots.setdefault(key, set()).add(finding['unit'])
    diagnostics = json.loads((ROOT / 'stage3/census/data/diagnostics.json').read_text())
    catch_unknown = [r for r in diagnostics if r['code'] == 18046 and r['profile'] == 'stock']
    return {'base_exception_roots': {str(k): len(v) for k, v in roots.items()},
            'hidden_exception_reasons': reasons,
            'catch_unknown': {'diagnostics': len(catch_unknown),
                              'file_roots': len({r['file'] for r in catch_unknown}),
                              'witnesses': sorted({f"{r['file']}:{r['line']}" for r in catch_unknown})}}


if __name__ == '__main__':
    result = report()
    assert belongs('throwing a number')
    assert belongs('a try around repeat, whose failure is a panic')
    assert not belongs('reading exception')
    assert not belongs('checked view field parent of type CatchClause | VariableDeclarationList')
    assert result['catch_unknown']['diagnostics'] == 8
    assert result['catch_unknown']['file_roots'] == 5
    assert result['base_exception_roots'] == {}
    assert result['hidden_exception_reasons'] == []
    print(json.dumps(result, indent=2, sort_keys=True))
