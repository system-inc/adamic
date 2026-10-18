#!/usr/bin/env python3
"""Hold ambient package unions to stock TypeScript and run union mutants."""
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent


def main():
    scratch = Path(tempfile.mkdtemp(prefix='adamic-reference-types-'))
    tsc = os.environ['PROJECT_LOADER_TSC']

    def run(name, args, expected=0):
        log = scratch / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(args, cwd=ROOT, stdout=output, stderr=subprocess.STDOUT)
        assert result.returncode == expected, (args, result.returncode, log.read_text())
        return log.read_text()

    probe = scratch / 'probe'
    compiler = scratch / 'adamic'
    run('build-probe', ['go', 'build', '-o', str(probe), './stage3/project-loader/probe'])
    run('build-compiler', ['go', 'build', '-o', str(compiler), './cmd/adamic'])
    # Authored fixtures must also pass the gate's standalone .a check, before
    # they are adapted into tsconfig-owned sources for the project oracle.
    authored_files = sorted((ROOT / 'stage3/project-loader/fixture').rglob('*.a'))
    authored_files += sorted((HERE / 'fixture').rglob('*.a'))
    for index, authored in enumerate(authored_files):
        run('authored-check-' + str(index), [str(compiler), 'c', str(authored)])
    work = scratch / 'fixture'
    for authored in (HERE / 'fixture/types').rglob('*.a'):
        relative = authored.relative_to(HERE / 'fixture/types')
        if relative.parts[0] == 'packages':
            target = work / 'node_modules/@types' / authored.stem / 'index.d.ts'
        else:
            target = work / relative.with_suffix('.ts')
        target.parent.mkdir(parents=True, exist_ok=True)
        text = authored.read_text().replace('.a\'', '.js\'')
        # Authored .a fixtures resolve aliases through type-only module imports.
        # The project-loading oracle supplies those same aliases exclusively
        # through each project's selected @types package instead.
        if relative.parts[0] == 'packages':
            assert text.startswith('export type '), text
            text = text.removeprefix('export ')
        else:
            lines = text.splitlines(keepends=True)
            assert lines[0].startswith('import type { '), text
            text = ''.join(lines[1:])
        target.write_text(text)
    (work / 'prelude.d.ts').write_text((ROOT / 'internal/load/prelude.d.ts').read_text())
    (work / 'package.json').write_text('{"type":"module"}\n')
    options = dict(strict=True, target='es2020', module='esnext', moduleResolution='bundler',
                   lib=['es2020'], skipLibCheck=False, declaration=True)
    for project, source in [('app', 'main.ts'), ('dependency', 'value.ts')]:
        config = dict(compilerOptions=dict(options, composite=True, rootDir='.',
                      outDir='../out/' + project, types=['entry' if project == 'app' else 'dependency']),
                      files=[source, '../prelude.d.ts'],
                      references=[dict(path='../dependency')] if project == 'app' else [])
        (work / project / 'tsconfig.json').write_text(json.dumps(config))
    entry = str(work / 'app/main.ts')
    assert json.loads(run('clean-loader', [str(probe), entry])) == dict(loaded=True, entries=1)
    run('clean-stock-solution', ['node', tsc, '--build', str(work / 'app'), '--pretty', 'false'])
    assert run('clean-stock-solution-output', ['node', str(work / 'out/app/main.js')]) == '7\n'
    # This is the external oracle for the one shared program, with exactly the
    # union selected by the ruling. Solution builds keep ambient scopes separate.
    flat = dict(compilerOptions=dict(options, types=['entry', 'dependency'], rootDir='.', outDir='flat-out'),
                files=['app/main.ts', 'dependency/value.ts', 'prelude.d.ts'])
    (work / 'tsconfig.json').write_text(json.dumps(flat))
    run('clean-stock-union', ['node', tsc, '-p', str(work), '--pretty', 'false'])
    assert run('clean-stock-union-output', ['node', str(work / 'flat-out/app/main.js')]) == '7\n'
    executable = work / 'native'
    run('clean-native-build', [str(compiler), 'build', entry, '-o', str(executable), '--sanitize'])
    assert run('clean-native-output', [str(executable)]) == '7\n'
    javascript = run('clean-javascript-build', [str(compiler), 'js', entry])
    (work / 'compiled.js').write_text(javascript)
    assert run('clean-javascript-output', ['node', str(ROOT / 'oracle/node.mjs'), str(work / 'compiled.js')]) == '7\n'
    for package in ('entry', 'dependency'):
        declaration = work / 'node_modules/@types' / package / 'index.d.ts'
        declaration.write_text(declaration.read_text() + 'declare const sharedAmbient: number;\n')
    # Each separate project still checks; the union must now report the two
    # duplicate declarations rather than picking whichever package came first.
    run('duplicate-stock-solution', ['node', tsc, '--build', str(work / 'app'), '--force', '--pretty', 'false'])
    rejected = json.loads(run('duplicate-loader', [str(probe), entry]))
    assert not rejected['loaded'], rejected
    stock = run('duplicate-stock-union', ['node', tsc, '-p', str(work), '--pretty', 'false'], 2)
    loader_diagnostics = sorted(re.findall(r'(@types/[^:]+):(\d+):(\d+): error TS(\d+): ([^\n]+)', rejected['error']))
    stock_diagnostics = sorted(re.findall(r'(@types/[^()]+)\((\d+),(\d+)\): error TS(\d+): ([^\n]+)', stock))
    assert loader_diagnostics == stock_diagnostics and len(loader_diagnostics) == 2, (loader_diagnostics, stock_diagnostics)
    # Declaration skipping in separate projects cannot erase conflicts newly
    # introduced by sharing their ambient packages in a native program.
    for project in ('app', 'dependency'):
        path = work / project / 'tsconfig.json'
        config = json.loads(path.read_text())
        config['compilerOptions']['skipLibCheck'] = True
        path.write_text(json.dumps(config))
    skipped = json.loads(run('duplicate-skipped-loader', [str(probe), entry]))
    assert skipped == rejected, skipped
    killed = []
    for name, file, needle, replacement, test, witness in [
        ('entry-types-only', 'project_references.go', 'return roots, owners, types, nil',
         'return roots, owners, entry.CompilerOptions().Types, nil', 'clean', 'ambient types union lost'),
        ('audit-entry-types-only', 'project_options.go', 'base.Types = types',
         'base.Types = config.CompilerOptions().Types; _ = types', 'clean', 'audit lost ambient types union'),
        ('skip-union-declarations', 'load.go', 'options.SkipLibCheck = core.TSFalse',
         'options.SkipLibCheck = core.TSTrue', 'duplicate-skipped', 'want both checker duplicate declarations'),
        ('audit-skip-union-declarations', 'project_options.go', 'base.SkipLibCheck = core.TSFalse',
         'base.SkipLibCheck = core.TSTrue', 'duplicate-skipped', 'want ordinary duplicate declarations'),
    ]:
        original = ROOT / 'internal/load' / file
        text = original.read_text()
        assert text.count(needle) == 1
        changed = scratch / (name + '.go')
        changed.write_text(text.replace(needle, replacement))
        overlay = scratch / (name + '.json')
        overlay.write_text(json.dumps(dict(Replace={str(original): str(changed)})))
        log = run(name, ['go', 'test', '-overlay', str(overlay), './internal/load', '-count=1',
                         '-run', '^' + ('TestProjectReferencesTypesUnion' if test == 'clean' else 'TestProjectReferencesTypesUnionDuplicateSkipped') + '$'], 1)
        assert '--- FAIL:' in log and witness in log and '[build failed]' not in log, log
        killed.append(name)
    summary = dict(authored_checked=[str(path.relative_to(ROOT)) for path in authored_files], clean_output='7\n', duplicate_diagnostics=loader_diagnostics, mutants_caught=killed, logs=str(scratch))
    (scratch / 'RESULT.json').write_text(json.dumps(summary, indent=2) + '\n')
    print(json.dumps(summary, indent=2))


if __name__ == '__main__':
    main()
