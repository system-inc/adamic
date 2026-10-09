#!/usr/bin/env python3
"""Run with the repository toolchain environment sourced; output stays in scratch."""
import gzip, json, os, pathlib, shutil, subprocess, sys, tempfile

root = pathlib.Path(__file__).resolve().parents[3]
evidence = pathlib.Path(__file__).resolve().parent
scratch = pathlib.Path(tempfile.mkdtemp(prefix='adamic-outline-'))
real_clang = shutil.which('clang')
assert real_clang, 'source the setup toolchain environment first'
for label in ('before', 'after'):
    (scratch / (label + '.c')).write_bytes(gzip.decompress((evidence / (label + '.c.gz')).read_bytes()))
for mode in ('full', 'line'):
    folder = scratch / mode
    folder.mkdir()
    wrapper = folder / 'clang'
    wrapper.write_text((evidence / 'clang.py').read_text().replace("real = '/workspace/adamic-tools/llvm/bin/clang'", 'real = ' + repr(real_clang)))
    wrapper.chmod(0o755)
env = os.environ.copy()
env.update(OUTLINE_MEASURE=str(scratch), ADAMIC_GATE_UNCACHED='1', ADAMIC_NATIVE_SPLIT='1', ADAMIC_NATIVE_JOBS='5')
for source, test, label in [('measure.go.txt', 'TestMeasureOutlinedBuild', 'measure'), ('debug.go.txt', 'TestOutlinedSanitizerLocations', 'debug')]:
    overlay = scratch / (label + '.json')
    overlay.write_text(json.dumps({'Replace': {str(root / ('internal/native/outline_' + label + '_test.go')): str(evidence / source)}}))
    command = ['go', 'test', '-overlay=' + str(overlay), './internal/native', '-run', '^' + test + '$', '-v', '-count=1', '-timeout=30m']
    with (scratch / (label + '.log')).open('wb') as log:
        result = subprocess.run(command, cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT)
    print(label, 'exit', result.returncode, 'log', scratch / (label + '.log'), flush=True)
    if result.returncode:
        raise SystemExit(result.returncode)
if '--end-to-end' in sys.argv:
    baseline = scratch / 'baseline-emit.go'
    baseline.write_bytes(subprocess.check_output(['git', 'show', 'c4e133659f737fb43d0d02da3811211bfa20a2d0:internal/native/emit.go'], cwd=root))
    outline = scratch / 'baseline-outline.go'
    outline.write_text((root / 'internal/native/outline.go').read_text().replace('func (e *emitter) moduleMain() string {', 'func (e *emitter) outlinedModuleMain() string {'))
    env['OUTLINE_ROOT'] = str(root)
    for version in ('before', 'after'):
        env['OUTLINE_VERSION'] = version
        replacements = {str(root / 'internal/native/outline_end_to_end_test.go'): str(evidence / 'end_to_end.go.txt')}
        if version == 'before':
            replacements.update({str(root / 'internal/native/emit.go'): str(baseline), str(root / 'internal/native/outline.go'): str(outline)})
        overlay = scratch / ('end-to-end-' + version + '.json')
        overlay.write_text(json.dumps({'Replace': replacements}))
        command = ['go', 'test', '-overlay=' + str(overlay), './internal/native', '-run', '^TestOutlineEndToEnd$', '-v', '-count=1', '-timeout=30m']
        with (scratch / ('end-to-end-' + version + '.log')).open('wb') as log:
            result = subprocess.run(command, cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT)
        print('end-to-end', version, 'exit', result.returncode, flush=True)
        if result.returncode:
            raise SystemExit(result.returncode)
print('Saved timings, generated units, sanitizer reports:', scratch)
