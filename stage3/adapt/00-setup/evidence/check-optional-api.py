#!/usr/bin/env python3
"""Run the lane's sanctioned API check on actual emitted declaration files."""
import difflib
import json
from pathlib import Path
import sys
import tempfile

sys.dont_write_bytecode = True
REPO = Path(__file__).resolve().parents[4]
sys.path.insert(0, str(REPO / 'stage3/lane'))
import check

before, after = (Path(argument).resolve() for argument in sys.argv[1:])
old_api = (before / 'built/local/typescript.d.ts').read_bytes()
new_api = (after / 'built/local/typescript.d.ts').read_bytes()
assert old_api == new_api
sanctioned = json.loads((REPO / 'stage3/lane/sanctioned-api.json').read_text())
with tempfile.TemporaryDirectory(prefix='optional-api-proof-') as scratch:
    results = Path(scratch)
    tree = results / 'adapted-tree'
    tree.mkdir()
    (tree / '.git').symlink_to(after / '.git', target_is_directory=True)
    reference = tree / 'tests/baselines/reference/api/typescript.d.ts'
    local = tree / 'tests/baselines/local/api/typescript.d.ts'
    reference.parent.mkdir(parents=True)
    local.parent.mkdir(parents=True)
    reference.write_bytes((after / 'tests/baselines/reference/api/typescript.d.ts').read_bytes())
    local.write_text(new_api.decode().replace('\r\n', '\n'))
    (results / 'oracle').mkdir()

    def observe():
        (results / 'oracle/baseline.diff').write_text(''.join(difflib.unified_diff(
            reference.read_text().splitlines(True), local.read_text().splitlines(True),
            fromfile='reference/api/typescript.d.ts', tofile='local/api/typescript.d.ts')))
        return check.api_check(results, sanctioned)

    errors, report = observe()
    assert not errors, errors
    assert report['sanctioned_declarations'] == report['composed_declarations'] == 222
    print(json.dumps(dict(status='pass', API=report, beforeAfterByteIdentical=True)))
    original = 'directoryExists?(directoryName: string): boolean;'
    text = local.read_text()
    assert text.count(original) == 1
    local.write_text(text.replace(original, 'directoryExists?: (directoryName: string) => boolean;'))
    errors, _ = observe()
    assert errors == ['unsanctioned API declaration: member:["ts","ModuleResolutionHost","directoryExists"]#0'], errors
    print(json.dumps(dict(mutant='public method converted to property', caught=errors)))
