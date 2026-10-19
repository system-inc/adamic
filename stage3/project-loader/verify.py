#!/usr/bin/env python3
"""Node project-build oracle and real loader mutants; all subprocess output goes to logs."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent


def run(args, log, *, expected=0):
    with log.open('w') as out:
        result = subprocess.run(args, cwd=ROOT, stdout=out, stderr=subprocess.STDOUT)
    if result.returncode != expected:
        raise AssertionError(f'{args}: exit {result.returncode}, wanted {expected}; {log}')
    return log.read_text()


def main():
    scratch = Path(tempfile.mkdtemp(prefix='adamic-project-loader-'))
    options = dict(composite=True, strict=True, module='esnext', moduleResolution='bundler',
                   target='es2020', types=[], skipLibCheck=True, rootDir='.')
    for project, source in [('app', 'main'), ('dependency', 'value')]:
        directory = scratch / project
        directory.mkdir()
        text = (HERE / 'fixture' / project / (source + '.a')).read_text()
        (directory / (source + '.ts')).write_text(text.replace('value.a', 'value.js'))
        config = dict(compilerOptions=dict(options, outDir='../out/' + project), files=[source + '.ts'])
        if project == 'app':
            config['references'] = [dict(path='../dependency')]
        (directory / 'tsconfig.json').write_text(json.dumps(config))
    (scratch / 'package.json').write_text('{"type":"module"}')
    probe = scratch / 'probe'
    run(['go', 'build', '-o', str(probe), './stage3/project-loader/probe'], scratch / 'build.log')
    entry = str(scratch / 'app/main.ts')
    before = json.loads(run([str(probe), entry], scratch / 'before.log'))
    assert not before['loaded'] and 'app/main.ts:1:23: error TS6305:' in before['error'], before
    baseline_source = scratch / 'main-load.go'
    baseline_source.write_text(run(['git', 'show', '3ffb1a835184713998a34874e86326cd21db971f:internal/load/load.go'], scratch / 'main-source.log'))
    baseline_overlay = scratch / 'main-overlay.json'
    baseline_overlay.write_text(json.dumps(dict(Replace={str(ROOT / 'internal/load/load.go'): str(baseline_source)})))
    baseline_probe = scratch / 'main-probe'
    run(['go', 'build', '-overlay', str(baseline_overlay), '-o', str(baseline_probe), './stage3/project-loader/probe'], scratch / 'main-build.log')
    baseline = json.loads(run([str(baseline_probe), entry], scratch / 'main-baseline.log'))
    assert not baseline['loaded'] and 'TS2345' in baseline['error'] and 'TS6305' not in baseline['error'], baseline
    tsc = os.environ['PROJECT_LOADER_TSC']
    run(['node', tsc, '--build', str(scratch / 'app'), '--verbose'], scratch / 'node-build.log')
    output = run(['node', str(scratch / 'out/app/main.js')], scratch / 'node-output.log')
    assert output == '7\n', output
    after = json.loads(run([str(probe), entry], scratch / 'after.log'))
    assert after == dict(loaded=True, entries=1), after
    mutants = [
        ('composite-roots', 'load.go', 'checkRoots = append(append([]tspath.RootedFilePath{}, projectConfig.FileNames()...), roots...)', 'checkRoots = roots', 'TestProjectLoaderHostConsoleAndCompositeEntries'),
        ('references', 'load.go', 'projectConfig.ProjectReferences()', 'nil', 'TestProjectLoaderReferenceNeedsOutput'),
        ('audit', 'load.go', 'for _, site := range sites {', 'for _, site := range []OptionSite{} {', 'TestProjectLoaderOptionAuditFailsClosed'),
        ('no-check', 'project_loader.go', 'if options.NoCheck == core.TSTrue {', 'if false {', 'TestProjectLoaderOwnershipAndNoCheck/noCheck'),
        ('host-console', 'source_fs.go', 'if s.projectConsole {', 'if false {', 'TestProjectLoaderHostConsoleAndCompositeEntries'),
    ]
    killed = []
    for name, file, needle, replacement, test in mutants:
        original = ROOT / 'internal/load' / file
        text = original.read_text()
        assert text.count(needle) == 1, (name, text.count(needle))
        changed = scratch / (name + '.go')
        changed.write_text(text.replace(needle, replacement))
        overlay = scratch / (name + '.json')
        overlay.write_text(json.dumps(dict(Replace={str(original): str(changed)})))
        log = scratch / (name + '.log')
        run(['go', 'test', '-overlay', str(overlay), './internal/load', '-count=1', '-run', '^' + test + '$'], log, expected=1)
        assert '--- FAIL:' in log.read_text(), log.read_text()
        killed.append(name)
    summary = dict(main_baseline=baseline, before=before, node_output=output, after=after, mutants_caught=killed,
                   logs=str(scratch), source_roots_redirect=False)
    (scratch / 'RESULT.json').write_text(json.dumps(summary, indent=2) + '\n')
    print(json.dumps(summary, indent=2))


if __name__ == '__main__':
    main()
