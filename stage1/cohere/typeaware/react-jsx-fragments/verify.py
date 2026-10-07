#!/usr/bin/env python3
"""Private typed-runner comparison; the shared syntax context has no checker lease."""
import argparse
import gzip
import json
from pathlib import Path
import re
import subprocess
import statistics
import time

parser = argparse.ArgumentParser()
parser.add_argument('--artifacts', type=Path, required=True)
parser.add_argument('--compiler', type=Path, required=True)
parser.add_argument('--checker', type=Path, required=True)
parser.add_argument('--sanitized-checker', type=Path, required=True)
args = parser.parse_args()
root = Path(__file__).resolve().parents[4]
own = Path(__file__).resolve().parent
out = args.artifacts.resolve()
out.mkdir(parents=True, exist_ok=True)
sequence = 0

def run(label, command, cwd=root, expected=0):
    global sequence
    sequence += 1
    stem = out / f'{sequence:03d}-{label}'
    started = time.monotonic()
    with stem.with_suffix('.stdout').open('wb') as stdout, stem.with_suffix('.stderr').open('wb') as stderr:
        result = subprocess.run(list(map(str, command)), cwd=cwd, stdout=stdout, stderr=stderr, timeout=600)
    assert result.returncode == expected, (label, result.returncode, expected)
    return stem.with_suffix('.stdout').read_bytes(), stem.with_suffix('.stderr').read_bytes(), time.monotonic() - started

virtual = root / 'cohere/wave20_jsx_fragments_oracle.go'
overlay = out / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {str(virtual): str(own / 'testdata/oracle.go')}}))
oracle = out / 'oracle'
run('oracle-build', ['go', 'build', '-overlay', overlay, '-o', oracle, virtual], root / 'cohere')
entry = own / 'suite.a'
native = out / 'suite'
san = out / 'suite-asan'
run('native-build', [args.compiler, 'build', entry, '-o', native, '--tsgo', args.checker])
run('asan-build', [args.compiler, 'build', entry, '-o', san, '--tsgo', args.sanitized_checker, '--sanitize'])
source = (root / 'cohere/internal/lint/rules/react/jsx_fragments_test.go').read_text()
literal = r'"(?:[^"\\]|\\.)*"'
rows = []
for match in re.finditer(r'\{\s*(' + literal + r'),\s*(' + literal + r')', source):
    label, text = map(json.loads, match.groups())
    if '<' not in text:
        continue
    tail = source[match.end():].split('},', 1)[0]
    flags = ['--element'] if 'JsxFragmentsElement' in tail else []
    rows.append((label, text, '.tsx', flags))
assert len(rows) >= 35, len(rows)
rows += [
 ('unicode before named fragment', 'const note = "😀"; <React.Fragment />;', '.tsx', []),
 ('unicode before imported declaration', 'const note = "😀"; import {Fragment as F} from "react"; <F />;', '.tsx', []),
 ('comments in pragma', '<React /* a */ . /* b */ Fragment />;', '.tsx', []),
 ('type annotation and initializer', 'const F: any = React.Fragment; <F />;', '.tsx', []),
 ('bare pragma initializer', 'const F = React; <F />;', '.tsx', []),
 ('no initializer', 'let React; <React />;', '.tsx', []),
 ('array destructuring is declined', 'const [F] = React; <F />;', '.tsx', []),
 ('wrong imported name', 'import {Other as Fragment} from "react"; <Fragment />;', '.tsx', []),
 ('escaped module name', 'import {Fragment} from "r\\u0065act"; <Fragment />;', '.tsx', []),
 ('foreign ambient declaration without initializer', '<ExternalF />;', '.tsx', []),
 ('foreign declaration initialized with pragma', '<ExternalFragment />;', '.tsx', []),
 ('binding alias name is not tested by Go', 'const {Other: F} = React; <F />;', '.tsx', []),
]

findings = 0
truths = []
fixture_records = []
for index, (label, text, extension, flags) in enumerate(rows):
    project = out / f'case-{index:03d}'
    project.mkdir(exist_ok=True)
    input_path = project / ('input' + extension)
    input_path.write_text(text)
    (project / 'globals.d.ts').write_text('// 😀 foreign declaration positions\ndeclare var Text: any;\ndeclare const ExternalF: any;\ndeclare const ExternalFragment = React.Fragment;\n')
    config = project / 'tsconfig.json'
    config.write_text(json.dumps({'compilerOptions': {'strict': True, 'allowJs': True, 'jsx': 'preserve', 'target': 'ES2022', 'types': []}, 'files': [input_path.name, 'globals.d.ts']}))
    manifest = project / 'manifest'
    manifest.write_text(str(input_path) + '\n')
    command = [config, manifest, *flags]
    truth = run(f'{index}-go', [oracle, *command])[0]
    for kind, binary in [('native', native), ('asan', san)]:
        got, errors, _ = run(f'{index}-{kind}', [binary, *command])
        assert not errors and got == truth, (label, kind)
    findings += int(truth.splitlines()[-1].split()[-1])
    truths.append((command, truth))
    fixture_records.append({'label': label, 'source': text, 'extension': extension, 'flags': flags})
print('PASS', len(rows), 'controls', findings, 'findings; complete Go/native/sanitizer bytes', flush=True)
(out / 'fixtures.json.gz').write_bytes(gzip.compress(json.dumps(fixture_records, indent=2).encode(), mtime=0))
mutated = own / 'analysis.a'
original = mutated.read_text()
assert original.count("base.text === 'React'") == 1
mutation = out / 'analysis-mutant.a'
mutation.write_text(re.sub(r"from '([^']+)'", lambda match: "from '" + str((own / match[1]).resolve()) + "'" if match[1].startswith('.') else match[0], original.replace("base.text === 'React'", "base.text === 'Preact'")))
changed = out / 'suite-mutant.a'
changed.write_text(re.sub(r"from '([^']+)'", lambda match: "from '" + str(mutation if match[1] == './analysis.a' else (own / match[1]).resolve()) + "'" if match[1].startswith('.') else match[0], entry.read_text()))
binary = out / 'mutant'
run('mutant-build', [args.compiler, 'build', changed, '-o', binary, '--tsgo', args.checker])
for command, truth in truths:
    got, errors, _ = run('mutant', [binary, *command])
    assert not errors
    if got != truth:
        print('KILLED pragma-object mutant: compiled, exit 0, empty stderr, Go bytes differ', flush=True)
        break
else:
    raise AssertionError('semantic mutant survived')
binary.unlink()
corpora = []
for name, config in [('repository', root / 'tsconfig.json'), ('compiler', Path('/workspace/wave20-typescript/src/compiler/tsconfig.json'))]:
    manifest = Path('/workspace/wave20-validation/harness-third') / (name + '.manifest')
    truth = run(name + '-go', [oracle, config, manifest])[0]
    for kind, binary in [('native', native), ('asan', san)]:
        got, errors, _ = run(name + '-' + kind, [binary, config, manifest])
        assert not errors and got == truth, (name, kind)
    print('PASS', name, len(truth), 'complete corpus bytes normal/sanitized', flush=True)
    corpora.append((name, config, manifest))
released_source = root / 'stage1/cohere/typeaware/prefer-promise-reject-errors/released.a'
released = out / 'released'
run('released-build', [args.compiler, 'build', released_source, '-o', released, '--tsgo', args.checker])
probe = out / 'probe.a'
probe.write_text('x();\n')
config = out / 'tsconfig.json'
config.write_text('{"compilerOptions":{"strict":true,"target":"ES2022","types":[]},"sourceExtensions":[".a"]}')
_, errors, _ = run('released-query', [released, config, probe], expected=70)
assert b'invalid or released checker handle' in errors
print('PASS released binding-origin query panics 70', flush=True)
bench = {}
for name, config, manifest in corpora:
    times = {'native': [], 'go': []}
    for round_index in range(3):
        outputs = {}
        for kind in (['native', 'go'] if round_index % 2 == 0 else ['go', 'native']):
            output, errors, elapsed = run(name + '-' + kind + '-timed', [native if kind == 'native' else oracle, config, manifest])
            if kind == 'native':
                assert not errors
            outputs[kind] = output
            times[kind].append(elapsed)
        assert outputs['native'] == outputs['go']
    bench[name] = {kind: statistics.median(values) for kind, values in times.items()}
    print('TIMING', name, bench[name], flush=True)
(out / 'summary.json').write_text(json.dumps({'controls': len(rows), 'findings': findings, 'semantic_mutant': 'caught by bytes', 'corpora': 'both normal/sanitized PASS', 'released_question': 'panic 70', 'bench': bench, 'shared_registration': False}, indent=2) + '\n')
print('PASS analysis controls, both corpora, semantic mutant, sanitizer and released question; shared checker integration pending', flush=True)
