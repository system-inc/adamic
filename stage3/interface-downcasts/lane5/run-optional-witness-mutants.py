#!/usr/bin/env python3
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[3]
pin = sys.argv[1]
logs = Path('/tmp/lane5-optional-witness-mutants')
logs.mkdir(exist_ok=True)
for directory, before, after in [
    ('parameter-declaration', 'modifiers: readonly ModifierLike[] | undefined,', 'modifiers: readonly ModifierLike[],'),
    ('variable-statement', 'declarationList: VariableDeclarationList | readonly VariableDeclaration[]', 'declarationList: VariableDeclarationList'),
]:
    fixture = root / 'stage3/interface-downcasts/lane5/optional-aggregates' / directory / 'good.a'
    if fixture.with_suffix('.ts').is_file():
        fixture = fixture.with_suffix('.ts')
    original = fixture.read_text()
    assert original.count(before) == 1
    try:
        fixture.write_text(original.replace(before, after))
        with (logs / (directory + '.log')).open('w') as log:
            result = subprocess.run(['node', 'stage3/interface-downcasts/lane5/verify-optional-aggregate-fixtures.cjs', pin], cwd=root, stdout=log, stderr=subprocess.STDOUT)
        evidence = (logs / (directory + '.log')).read_text()
        assert result.returncode != 0 and 'original declaration changed' in evidence, evidence
        print(directory + ': shortened declaration rejected by independent original-span verifier', flush=True)
    finally:
        fixture.write_text(original)
