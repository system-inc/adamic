#!/usr/bin/env python3
"""Compare source-root programs with stock tsc and kill real loader mutants."""
import json
import os
import re
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent


def main():
    scratch = Path(tempfile.mkdtemp(prefix='adamic-reference-sources-'))
    tsc = os.environ['PROJECT_LOADER_TSC']
    def run(name, args, expected=0):
        log = scratch / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(args, cwd=ROOT, stdout=output, stderr=subprocess.STDOUT)
        assert result.returncode == expected, (args, result.returncode, log.read_text(), str(log))
        return log.read_text()
    probe = scratch / 'probe'
    compiler = scratch / 'adamic'
    run('build-probe', ['go', 'build', '-o', str(probe), './stage3/project-loader/probe'])
    run('build-compiler', ['go', 'build', '-o', str(compiler), './cmd/adamic'])
    run('stock-version', ['node', tsc, '--version'])
    observations = {}
    for kind in ('two', 'chain'):
        work = scratch / kind
        source = ROOT / 'stage3/project-loader/fixture' if kind == 'two' else HERE / 'fixture/chain'
        for authored in source.rglob('*.a'):
            target = work / authored.relative_to(source).with_suffix('.ts')
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text(authored.read_text().replace('.a\'', '.js\'').replace('.a"', '.js"'))
        references = {'app': ['dependency']} if kind == 'two' else {'app': ['middle'], 'middle': ['leaf']}
        for directory in sorted(work.iterdir()):
            options = dict(composite=True, declaration=True, strict=True, target='es2020', module='esnext',
                           moduleResolution='bundler', types=[], skipLibCheck=True, rootDir='.',
                           outDir='../out/' + directory.name)
            config = dict(compilerOptions=options, files=sorted(p.name for p in directory.glob('*.ts')),
                          references=[dict(path='../' + name) for name in references.get(directory.name, [])])
            (directory / 'tsconfig.json').write_text(json.dumps(config))
        (work / 'package.json').write_text('{"type":"module"}\n')
        entry = str(work / 'app/main.ts')
        before = json.loads(run(kind + '-fresh', [str(probe), entry]))
        assert before == dict(loaded=True, entries=1), before
        run(kind + '-stock-build', ['node', tsc, '--build', str(work / 'app'), '--verbose'])
        output = run(kind + '-stock-output', ['node', str(work / 'out/app/main.js')])
        assert output == '7\n', output
        after = json.loads(run(kind + '-built', [str(probe), entry]))
        assert after == before, after
        # A configured, unimported source must have the same diagnostic location
        # and code on a solution build and on the flattened checker.
        error_project = 'dependency' if kind == 'two' else 'leaf'
        unused = work / error_project / 'unused.ts'
        unused.write_text('export const unused: number = "bad";\n')
        config_path = unused.parent / 'tsconfig.json'
        config = json.loads(config_path.read_text())
        if 'unused.ts' not in config['files']:
            config['files'].append('unused.ts')
        config_path.write_text(json.dumps(config))
        rejected = json.loads(run(kind + '-source-error', [str(probe), entry]))
        assert not rejected['loaded'] and 'unused.ts:1:14: error TS2322' in rejected['error'], rejected
        stock_error = run(kind + '-stock-error', ['node', tsc, '--build', str(work / 'app'), '--force', '--pretty', 'false'], 2)
        assert 'unused.ts(1,14): error TS2322' in stock_error, stock_error
        loader_diagnostics = re.findall(r'([\w.-]+\.ts):(\d+):(\d+): error TS(\d+): ([^\n]+)', rejected['error'])
        stock_diagnostics = re.findall(r'([\w.-]+\.ts)\((\d+),(\d+)\): error TS(\d+): ([^\n]+)', stock_error)
        assert loader_diagnostics == stock_diagnostics, (loader_diagnostics, stock_diagnostics)
        unused.write_text('export const unused: number = 1;\n')
        # The original DOM console is a separate lowering limitation in this
        # base. Preserve that observation instead of treating a type check as execution.
        for backend, args in [('native', ['build', entry, '-o', str(work / 'native')]), ('javascript', ['js', entry])]:
            refusal = run(kind + '-' + backend + '-host-console', [str(compiler)] + args, 1)
            assert ('console.log' in refusal or 'library member log' in refusal) and 'TS6305' not in refusal, refusal
        # A shared string-console profile uses Adamic's actual prelude on both
        # checkers. A local copy is a declaration alias, never another runtime entry.
        (work / 'prelude.d.ts').write_text((ROOT / 'internal/load/prelude.d.ts').read_text())
        main = work / 'app/main.ts'
        main.write_text(main.read_text().replace('console.log(value)', 'console.log(`${value}`)'))
        for directory in references.keys() | {error_project}:
            path = work / directory / 'tsconfig.json'
            config = json.loads(path.read_text())
            config['compilerOptions']['lib'] = ['es2020']
            config['files'].append('../prelude.d.ts')
            path.write_text(json.dumps(config))
        shutil.rmtree(work / 'out')
        run(kind + '-string-stock-build', ['node', tsc, '--build', str(work / 'app'), '--force', '--pretty', 'false'])
        string_output = run(kind + '-string-stock-output', ['node', str(work / 'out/app/main.js')])
        assert string_output == output
        shutil.rmtree(work / 'out')
        executable = work / 'native'
        run(kind + '-native-build', [str(compiler), 'build', entry, '-o', str(executable), '--sanitize'])
        native_output = run(kind + '-native-output', [str(executable)])
        javascript = run(kind + '-javascript-build', [str(compiler), 'js', entry])
        (work / 'compiled.js').write_text(javascript)
        javascript_output = run(kind + '-javascript-output', ['node', str(ROOT / 'oracle/node.mjs'), str(work / 'compiled.js')])
        assert native_output == javascript_output == output, (native_output, javascript_output, output)
        dependency_path = work / error_project / 'tsconfig.json'
        conflicting = json.loads(dependency_path.read_text())
        conflicting['compilerOptions']['strictNullChecks'] = False
        dependency_path.write_text(json.dumps(conflicting))
        conflict = json.loads(run(kind + '-option-conflict', [str(probe), entry]))
        referencing = 'app' if kind == 'two' else 'middle'
        expected_conflict = (f"load: project reference options conflict: {work / referencing / 'tsconfig.json'} "
                             f"references {dependency_path}: compiler option strictNullChecks differs; "
                             "separate checker ownership is required")
        assert conflict == dict(loaded=False, error=expected_conflict), conflict
        # A solution build can keep separate option ownership; the native loader
        # deliberately refuses to flatten that same otherwise-valid graph.
        run(kind + '-stock-conflict-build', ['node', tsc, '--build', str(work / 'app'), '--force', '--pretty', 'false'])
        assert run(kind + '-stock-conflict-output', ['node', str(work / 'out/app/main.js')]) == output
        observations[kind] = dict(fresh=before, built=after, stock_output=output, native_output=native_output,
                                  javascript_output=javascript_output, bad_source='unused.ts:1:14: TS2322')
    mutants = [
        ('drop-transitive', 'project_references.go', 'for _, reference := range config.ProjectReferences() {',
         'for _, reference := range config.ProjectReferences() { if config != entry { continue }', 'TestProjectReferencesTransitive', 'transitive'),
        ('declarations-only', 'load.go', 'nil, currentDirectory.ResolveDirectory',
         'projectConfig.ProjectReferences(), currentDirectory.ResolveDirectory', 'TestProjectLoaderReferenceSources', 'TS6305'),
        ('allow-conflicts', 'project_references.go', 'if option := conflictingProjectOption(config.CompilerOptions(), dependency.CompilerOptions()); option != "" {',
         'if option := conflictingProjectOption(config.CompilerOptions(), dependency.CompilerOptions()); false {', 'TestProjectReferencesConflictingOptions', 'want'),
        ('drop-reference-audit', 'load.go', 'for _, site := range sites {',
         'for _, site := range []OptionSite{} {', 'TestProjectReferencesAudit', 'admitted'),
        ('duplicate-roots', 'project_references.go', 'if !seen[key] {',
         'if true {', 'TestProjectReferencesDiamond', 'duplicate'),
        ('physical-prelude', 'project_references.go', 'if root != preludePath && root.IsDeclarationFile() && fs.FileExists(preludePath) {',
         'if false {', 'TestProjectReferencesSharedPrelude', 'physical prelude console'),
        ('allow-cycle', 'project_references.go', 'if active[name] {',
         'if false {', 'TestProjectReferencesCycle', 'circular'),
    ]
    killed = []
    for name, file, needle, replacement, test, witness in mutants:
        original = ROOT / 'internal/load' / file
        text = original.read_text()
        assert text.count(needle) == 1, (name, text.count(needle))
        changed = scratch / (name + '.go')
        changed.write_text(text.replace(needle, replacement))
        overlay = scratch / (name + '.json')
        overlay.write_text(json.dumps(dict(Replace={str(original): str(changed)})))
        log = run(name, ['go', 'test', '-overlay', str(overlay), './internal/load', '-count=1', '-run', '^' + test + '$'], 1)
        assert '--- FAIL:' in log and witness in log and '[build failed]' not in log, log
        killed.append(name)
    summary = dict(observations=observations, mutants_caught=killed, logs=str(scratch))
    (scratch / 'RESULT.json').write_text(json.dumps(summary, indent=2) + '\n')
    print(json.dumps(summary, indent=2))


if __name__ == '__main__':
    main()
