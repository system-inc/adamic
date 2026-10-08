#!/usr/bin/env python3
"""Select this rule's upstream capture and witnesses with a temporary Go overlay.

The normal whole-package run has no overlay. No shared repository file is edited.
"""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[5]
package = root / 'stage1/cohere/lint'
with tempfile.TemporaryDirectory(prefix='typescript-optional-chain-selection-') as scratch:
    scratch = Path(scratch)
    replacements = {}
    for name, anchor in [
        ('shared_test.go', '\tfor _, d := range descriptors {\n\t\tdiscovered[d.Name] = true'),
        ('registration_test.go', '\tfor _, d := range prepareRegistry(t, ".") {'),
    ]:
        original = package / name
        source = original.read_text()
        assert source.count(anchor) == 1
        if name == 'shared_test.go':
            changed = anchor.replace('\n\t\tdiscovered', '\n\t\tif d.Name != "@typescript-eslint/no-non-null-asserted-optional-chain" { continue }\n\t\tdiscovered')
        else:
            changed = anchor + '\n\t\tif d.Name != "@typescript-eslint/no-non-null-asserted-optional-chain" { continue }'
        target = scratch / name
        selected = source.replace(anchor, changed)
        if name == 'shared_test.go':
            assert selected.count('if len(rows) < 150 {') == 1
            selected = selected.replace('if len(rows) < 150 {', 'if len(rows) != 24 {')
        target.write_text(selected)
        replacements[str(original)] = str(target)
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': replacements}))
    command = ['go', 'test', '-overlay', str(overlay), './stage1/cohere/lint', '-run', '^(TestRulesAgree|TestOwnedWitnesses|TestMutants)$/^optional-chain-parentheses-not-skipped$', '-count=1', '-v', '-timeout=30m']
    raise SystemExit(subprocess.call(command, cwd=root))
