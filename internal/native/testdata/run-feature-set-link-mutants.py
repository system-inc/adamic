#!/usr/bin/env python3
"""Run real Go overlays, retaining link evidence without changing the checkout."""
import json
import os
from pathlib import Path
import subprocess
import sys

repository = Path(__file__).resolve().parents[3]
logs = repository / 'review/compiler/feature-set-link-main/mutants'
logs.mkdir(parents=True, exist_ok=True)
header = repository / 'internal/native/runtime/adamic.h'
original = header.read_text()
name = next(line for line in original.splitlines() if line.startswith('#define ADAMIC_FEATURE_NAME_PARTS('))
reference = next(line for line in original.splitlines() if line.startswith('static const void *const adamic_runtime_feature_reference '))
mutants = [
    ('fixed-symbol', header, original.replace(name, '#define ADAMIC_FEATURE_NAME_PARTS(a, b, c, d, e) adamic_runtime_features'), '^TestRuntimeFeatureMismatchProgramFeature$', 'mismatched runtime linked successfully'),
    ('drop-reference', header, original.replace(reference, '/* mutant: no unit reference */'), '^TestRuntimeFeatureMismatchProgramFeature$', 'mismatched runtime linked successfully'),
    ('drop-retain', header, original.replace('__attribute__((used, retain))', '__attribute__((used))', 1), '^TestRuntimeFeatureMismatchProgramFeature$', 'mismatched runtime linked successfully'),
]
library = repository / 'internal/native/library.go'
base = subprocess.check_output(['git', 'show', '8c76557e:internal/native/library.go'], cwd=repository, text=True)
start = base.index('\tfor _, feature := range []string', base.index('func cachedRuntime'))
end = base.index('\n\t// Headers live beside', start)
current = library.read_text()
needle = '\t// Keep header bytes unchanged.'
start_current = current.index(needle)
end_current = current.index('\n\t// Headers live beside', start_current)
mutants.append(('header-forces-runtime-features', library, current[:start_current] + base[start:end] + current[end_current:], '^TestRuntimeFeatureMismatchRuntimeFeature$', 'mismatched runtime linked successfully'))
environment = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
# Keep Go's ordinary action cache when isolating each runtime cache.
environment['GOCACHE'] = subprocess.check_output(['go', 'env', 'GOCACHE'], text=True).strip()

def run_overlay(name, path, contents, selection, extra=None, additional=None):
    directory = logs / name
    directory.mkdir(exist_ok=True)
    changed = directory / (path.name + ".txt")
    changed.write_text(contents)
    overlay = directory / 'overlay.json'
    replacements = {str(path): str(changed)}
    for other, text in (additional or {}).items():
        destination = directory / (other.name + ".txt")
        destination.write_text(text)
        replacements[str(other)] = str(destination)
    overlay.write_text(json.dumps({'Replace': replacements}) + '\n')
    env = dict(environment, XDG_CACHE_HOME=str(Path('/tmp/adamic-feature-set-link-cache') / name))
    if extra:
        env.update(extra)
    log = directory / 'test.log'
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', '-overlay', str(overlay), './internal/native', '-run', selection, '-count=1', '-timeout', '90s', '-v'], cwd=repository, env=env, stdout=output, stderr=subprocess.STDOUT, timeout=150)
    return result.returncode, log

for name, path, contents, selection, catcher in ([] if sys.argv[1:] == ["checker-original"] else mutants):
    code, log = run_overlay(name, path, contents, selection)
    evidence = log.read_text()
    if code == 0 or catcher not in evidence or 'build failed' in evidence or 'error: linker command failed' in evidence:
        raise SystemExit(f'{name}: not caught solely by {catcher}; see {log}')
    print(f'{name}: caught by {catcher}; {log}', flush=True)

archive = os.environ.get('ADAMIC_CLANG_TSGO_ARCHIVE')
if not archive:
    raise SystemExit('ADAMIC_CLANG_TSGO_ARCHIVE is required for the third split-build-flags mutant')
path = repository / 'internal/native/units_tsgo.go'
contents = path.read_text()
needle = 'flags := append(sourceFlags(source, options), "-DADAMIC_TSGO")'
if contents.count(needle) != 1:
    raise SystemExit('split runtime flag mutation site moved')
changed = contents.replace(needle, 'flags := append(Flags(options), "-DADAMIC_TSGO")')
if sys.argv[1:] != ['checker-original']:
    code, log = run_overlay('split-runtime-default-flags', path, changed, '^TestSplitTSGoRuntimeFeatures$', {'ADAMIC_TEST_FEATURE_RUNTIME_MISMATCH': '1'})
    evidence = log.read_text()
    if code != 0 or 'undefined reference' not in evidence or 'adamic_runtime_features_closure_convention_closure_receivers' not in evidence or 'Sanitize: true' in evidence:
        raise SystemExit(f'split default flags: did not stop at the requested link symbol; see {log}')
    print(f'split-runtime-default-flags: unsanitized link rejection verified; {log}', flush=True)
    code, log = run_overlay('split-runtime-default-flags-positive-control', path, changed, '^TestSplitTSGoRuntimeFeatures$', {'ADAMIC_TEST_FEATURE_RUNTIME_MISMATCH': '0'})
    if code == 0 or 'undefined reference' not in log.read_text() or 'adamic_runtime_features_closure_convention_closure_receivers' not in log.read_text():
        raise SystemExit(f'split default flags: ordinary positive test did not catch the mutant; see {log}')
    print(f'split-runtime-default-flags: ordinary matching-build test catches mutant; {log}', flush=True)

# The original checker fixture, using the exact third mutation and an unsanitized test overlay.
test = repository / 'internal/native/units_tsgo_test.go'
ordinary = test.read_text()
needle = 'options := Options{Sanitize: true, Jobs: 5}'
if ordinary.count(needle) != 1:
    raise SystemExit('original checker test options moved')
ordinary = ordinary.replace(needle, 'options := Options{Jobs: 5}')
code, log = run_overlay('split-original-checker-default-flags', path, changed, '^TestSplitTSGoAgreesUnit00$', additional={test: ordinary})
evidence = log.read_text()
if code == 0 or 'undefined reference' not in evidence or 'adamic_runtime_features_' not in evidence or 'ERROR: AddressSanitizer' in evidence:
    raise SystemExit(f'original third mutant did not fail at feature-set link: {log}')
print(f'split-original-checker-default-flags: original checker fixture fails at unsanitized link; {log}', flush=True)
