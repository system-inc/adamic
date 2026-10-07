"""Standalone byte comparisons of the two native setter validators."""
import argparse
import json
import pathlib
import subprocess
import time

ROOT = pathlib.Path(__file__).resolve().parents[4]
OWN = pathlib.Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('--scratch', required=True)
parser.add_argument('--tools', required=True)
parser.add_argument('--rule', choices=['render', 'effect'], required=True)
parser.add_argument('--compiler-manifest', required=True)
parser.add_argument('--repository-manifest', required=True)
args = parser.parse_args()
scratch = pathlib.Path(args.scratch).resolve()
scratch.mkdir(parents=True, exist_ok=True)
tools = pathlib.Path(args.tools).resolve()
records = []


def run(name, command):
    started = time.monotonic()
    with (scratch / (name + '.stdout')).open('wb') as out, (scratch / (name + '.stderr')).open('wb') as err:
        result = subprocess.run([str(x) for x in command], cwd=ROOT, stdout=out, stderr=err)
    records.append(dict(name=name, command=[str(x) for x in command], exit=result.returncode, seconds=time.monotonic() - started))
    (scratch / 'runs.json').write_text(json.dumps(records, indent=2) + '\n')
    assert result.returncode == 0, (name, result.returncode, (scratch / (name + '.stderr')).read_text())
    return (scratch / (name + '.stdout')).read_bytes()


stage0 = tools / 'adamic'
archive = tools / 'checker.a'
virtual = ROOT / ('cohere/adamic_wave26_' + args.rule + '_oracle.go')
overlay = scratch / 'oracle-overlay.json'
overlay.write_text(json.dumps({'Replace': {str(virtual): str(OWN / ('testdata/' + args.rule + '_oracle.go'))}}))
oracle = scratch / 'oracle'
run('oracle-build', ['go', '-C', ROOT / 'cohere', 'build', '-overlay', overlay, '-o', oracle, virtual])
run('extract-build', ['go', 'build', '-o', scratch / 'extract', OWN / 'testdata/extract_cases.go'])
test = ROOT / ('cohere/internal/lint/rules/react/set_state_in_' + args.rule + '_test.go')
extracted = json.loads(run('extract', [scratch / 'extract', test]))
common = extracted['Constants']
if args.rule == 'render':
    ambient = common['reactStateDeclarations']
    body = ambient[ambient.index('{') + 1:ambient.rindex('}')]
else:
    body = common['reactStub']
    ambient = "declare module 'react' {\n" + body + '\n}'
(scratch / 'ambient.d.ts').write_text(ambient)
(scratch / 'react.d.ts').write_text(body)
(scratch / 'other.d.ts').write_text(common.get('otherModuleStub', 'export declare function other():void;'))
examples = extracted['Inputs']
if args.rule == 'render':
    # Preserve untyped originals alongside the typed versions production tests use.
    examples = [s for s in examples if 'import ' not in s] + [s if 'import ' in s else "import {useState,useMemo,useCallback,useRef,useReducer,useEffect} from 'react';\n" + s for s in examples]
files = []
for i, text in enumerate(examples):
    file = scratch / ('control-%03d.tsx' % i)
    file.write_text(text + '\nexport {};\n')
    files.append(file)
config = scratch / 'tsconfig.json'
config.write_text(json.dumps({'compilerOptions': {'strict': True, 'target': 'ES2022', 'module': 'ESNext', 'jsx': 'preserve'}, 'files': [str(files[0]), str(scratch / 'ambient.d.ts'), str(scratch / 'react.d.ts'), str(scratch / 'other.d.ts')]}))
manifest = scratch / 'controls.manifest'
manifest.write_text(''.join(str(file) + '\n' for file in files))
entry = OWN / (args.rule + '_suite.a')
binary = scratch / 'native'
asan = scratch / 'native-asan'
run('native-build', [stage0, 'build', entry, '-o', binary, '--tsgo', archive])
run('asan-build', [stage0, 'build', entry, '-o', asan, '--tsgo', archive, '--sanitize'])
truth = None
for name, paths in [('controls', manifest), ('compiler', args.compiler_manifest), ('repository', args.repository_manifest)]:
    expected = run(name + '-go', [oracle, config, paths])
    actual = run(name + '-native', [binary, config, paths])
    if actual != expected:
        print(name + ': byte mismatch; inspect retained stdout files', flush=True)
    assert actual == expected, name
    assert (scratch / (name + '-native.stderr')).read_bytes() == b''
    sanitized = run(name + '-asan', [asan, config, paths])
    assert sanitized == expected, name + '-asan'
    assert (scratch / (name + '-asan.stderr')).read_bytes() == b''
    print(name + ': ' + str(len(expected)) + ' exact bytes, ' + expected.splitlines()[-1].decode(), flush=True)
    if name == 'controls':
        truth = expected
        assert b'findings 0\n' not in truth
mutant_dir = scratch / ('set_state_in_' + args.rule)
mutant_dir.mkdir(exist_ok=True)
source = (OWN / ('set_state_in_' + args.rule + '/rule.a')).read_text()
source = source.replace("from '../../preference_structure.a'", 'from ' + repr(str(OWN.parent / 'preference_structure.a')))
source = source.replace("from '../../diagnostic.ts'", 'from ' + repr(str(OWN.parent / 'diagnostic.ts')))
for name in ['react_hir.a', 'hir_data.a']:
    source = source.replace("from '../" + name + "'", 'from ' + repr(str(OWN / name)))
source = source.replace("from './index.a'", 'from ' + repr(str(OWN / ('set_state_in_' + args.rule + '/index.a'))))
needle = "=== 'Dispatch'"
assert source.count(needle) == 1
source = source.replace(needle, "=== 'DispatchMutant'")
(mutant_dir / 'rule.a').write_text(source)
source = entry.read_text().replace("from './set_state_in_" + args.rule + "/rule.a'", 'from ' + repr(str(mutant_dir / 'rule.a')))
source = source.replace("from '../../../typescript/parser/nodes.ts'", 'from ' + repr(str(ROOT / 'stage1/typescript/parser/nodes.ts')))
source = source.replace("from '../preference_structure.a'", 'from ' + repr(str(OWN.parent / 'preference_structure.a')))
for name in ['react_hir.a', 'hir_data.a', 'react_candidate.a']:
    source = source.replace("from './" + name + "'", 'from ' + repr(str(OWN / name)))
mutant_entry = scratch / 'mutant.a'
mutant_entry.write_text(source)
mutant_binary = scratch / 'mutant'
run('mutant-build', [stage0, 'build', mutant_entry, '-o', mutant_binary, '--tsgo', archive])
assert run('mutant-run', [mutant_binary, config, manifest]) != truth
assert (scratch / 'mutant-run.stderr').read_bytes() == b''
print('setter-alias mutant: exit 0, empty stderr, byte comparison catches it', flush=True)
