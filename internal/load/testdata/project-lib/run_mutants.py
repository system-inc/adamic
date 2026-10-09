#!/usr/bin/env python3
"""Prove the overlay/project checker regressions without modifying the checkout."""
import json
from pathlib import Path
import subprocess
import sys

repository = Path(__file__).resolve().parents[4]
logs = Path(sys.argv[1] if len(sys.argv) > 1 else '/tmp/adamic-overlay-mutants').resolve()
logs.mkdir(parents=True, exist_ok=True)
mutants = {
    'iterator-element': ('internal/load/regexp_library.go',
        'next(): IteratorResult<T, undefined>;',
        'next(): { done: false; value: T } | { done: true; value: undefined };',
        'TestCollectionIteratorElementInference|TestProjectIteratorWitnesses'),
    'project-set-extensions': ('internal/load/load.go',
        'if !project {', 'if project || !project {',
        '^TestProjectLibCustomSet$/es2024'),
    'project-options': ('internal/load/project.go',
        'GetParsedCommandLineOfConfigFile(selected, nil, nil, fs, nil)',
        'GetParsedCommandLineOfConfigFile(selected, compilerOptions(), nil, fs, nil)',
        'TestProjectLibAndOptions|TestProjectExtendsAndDeclarationRoots'),
    'project-console': ('internal/load/source_fs.go',
        'if s.projectConsole {', 'if s.projectConsole && false {',
        '^TestProjectInheritedConsoleDeclaration$'),
    'project-declarations': ('internal/load/load.go',
        'if name.IsDeclarationFile() {', 'if name.IsDeclarationFile() && false {',
        '^TestProjectExtendsAndDeclarationRoots$'),
    'declaration-diagnostics': ('internal/load/load.go',
        'if len(all) == 0 && p.compiler.Options().GetEmitDeclarations() {',
        'if len(all) == 0 && p.compiler.Options().GetEmitDeclarations() && false {',
        '^TestProjectLibAndOptions$/isolated_declaration_needs_annotation'),
    'mixed-projects': ('internal/load/project.go',
        'if typescript && found != selected {',
        'if typescript && found != selected && false {',
        '^TestProjectConfigErrorsAndMixedRoots$'),
}
for name, (file, original, replacement, pattern) in mutants.items():
    source = repository / file
    text = source.read_text()
    assert original in text, f'{name}: mutation site missing'
    changed = logs / (name + '.go')
    changed.write_text(text.replace(original, replacement))
    overlay = logs / (name + '.json')
    overlay.write_text(json.dumps({'Replace': {str(source): str(changed)}}))
    log_path = logs / (name + '.log')
    with log_path.open('w') as log:
        result = subprocess.run(['go', 'test', '-overlay', str(overlay),
            './internal/load', '-run', pattern, '-count=1', '-v'],
            cwd=repository, stdout=log, stderr=subprocess.STDOUT)
    output = log_path.read_text()
    assert result.returncode == 1 and '--- FAIL: Test' in output and '[build failed]' not in output, (
        f'{name}: not killed by a regression check, see {log_path}')
    print(f'{name}: caught by checking, exit 1, {log_path}', flush=True)
